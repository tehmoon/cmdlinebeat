package beater

import (
  "github.com/elastic/beats/libbeat/beat"
  "github.com/elastic/beats/libbeat/common"
  "github.com/elastic/beats/libbeat/logp"
  "time"
  "fmt"
  "io"
  "github.com/tehmoon/errors"
  "bufio"
  "os/exec"
  "io/ioutil"
)

type Command struct {
  Command string `config:"command"`
  Shell string `config:"shell"`
  Env map[string]string `config:"env"`
  CopyEnv bool `config:"copy_env"`
  Sleep time.Duration `config:"sleep"`
  Timeout time.Duration `config:"timeout"`
  Fields common.MapStr `config:"fields"`
  entryNumber int
}

func (command Command) Run(b *beat.Beat, sync chan struct{}) {
  config := beat.ClientConfig{
    EventMetadata: common.EventMetadata{
      Fields: command.Fields,
    },
  }

  client, err := b.Publisher.ConnectWith(config)
  if err != nil {
    logp.Err(errors.Wrapf(err, "Enable to connect the publisher to config #%d", command.entryNumber).Error())
    return
  }

  tries := 3

  for {
    if tries == 0 {
      logp.Err("Stop retrying command #%d after 3 tries", command.entryNumber)

      break
    }

    cmd := exec.Command(command.Shell, "-c", command.Command)

    if ! command.CopyEnv {
      cmd.Env = make([]string, 0)
    }

    for k, v := range command.Env {
      cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, v))
    }

    stderrChan, err := CreateAndReadAllFromFn(cmd.StderrPipe)
    if err != nil {
      logp.Err(errors.Wrapf(err, "Error in reading from stderr in command #%d", command.entryNumber).Error())
      tries = decrementAfterSleep(tries, SLEEP_TIME)
      continue
    }

    go func() {
      err := <- stderrChan
      if err != nil {
        logp.Err(errors.Wrapf(err, "Error in command #%d, retrying...", command.entryNumber).Error())
      }
    }()

    doneReading, err := ReadLineFromReaderFnAndPublish(cmd.StdoutPipe, client, &command)
    if err != nil {
      logp.Err(errors.Wrapf(err, "Unable to open stdout in command #%d, retrying...", command.entryNumber).Error())

      tries = decrementAfterSleep(tries, SLEEP_TIME)
      continue
    }

    err = StartAndWaitCommand(cmd, doneReading)
    if err != nil {
      logp.Err(errors.Wrapf(err, "Error starting or waiting command #%d, retrying...", command.entryNumber).Error())

      tries = decrementAfterSleep(tries, SLEEP_TIME)
      continue
    }

    time.Sleep(command.Sleep)
  }

  sync <- struct{}{}
}

func StartAndWaitCommand(cmd *exec.Cmd, wait chan struct{}) (error) {
  err := cmd.Start()
  if err != nil {
    return errors.Wrap(err, "Error creating command")
  }

  <- wait

  err = cmd.Wait()
  if err != nil {
    return errors.Wrap(err, "Error executing command")
  }

  return nil
}

func CreateAndReadAllFromFn(fn func() (io.ReadCloser, error)) (chan error, error) {
  reader, err := fn()
  if err != nil {
    return nil, errors.Wrap(err, "Error in creating reader")
  }

  syncBack := make(chan error)

  go func() {
    output, err := ioutil.ReadAll(reader)
    if err != nil {
      syncBack <- errors.Wrapf(err, "Error reading stderr")
      return
    }

    if len(output) == 0 {
      syncBack <- nil
      return
    }

    if output[len(output) - 1] == '\n' {
      output = output[:len(output) - 1]
    }

    syncBack <- errors.Errorf("Stderr: %s", string(output[:]))
  }()

  return syncBack, nil
}

func ReadLineFromReaderFnAndPublish(fn func() (io.ReadCloser, error), client beat.Client, command *Command) (chan struct{}, error) {
  r, err := fn()
  if err != nil {
    return nil, errors.Wrap(err, "Error creating reader")
  }

  done := make(chan struct{})

  reader := bufio.NewReader(r)
  go func() {

    for i := int64(0);; i++ {
      line, err := reader.ReadString('\n')
      if err != nil {
        if err == io.EOF {
          break
        }

        logp.Err(errors.Wrapf(err, "Error reading line in command #%d, killing command and retring...", command.entryNumber).Error())

        break
      }

      if len(line) == 0 {
        break
      }

      if line[len(line) - 1] == '\n' {
        line = line[:len(line) - 1]
      }

      client.Publish(beat.Event{
        Timestamp: time.Now(),
        Fields: common.MapStr{
          "cmdlinebeat": &common.MapStr{
            "line": line,
            "number": i,
          },
        },
      })
    }

    done <- struct{}{}
  }()

  return done, nil
}
