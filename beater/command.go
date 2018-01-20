package beater

import (
  "fmt"
  "github.com/elastic/beats/libbeat/beat"
  "github.com/elastic/beats/libbeat/common"
  "github.com/elastic/beats/libbeat/logp"
  "time"
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
  Name string `config:"name"`
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
    logp.Err(errors.Wrapf(err, "Enable to connect the publisher for command %s", command.Name).Error())
    return
  }

  tries := 3
  env := ForkEnv(command.Env, command.CopyEnv)

  for {
    if tries == 0 {
      logp.Err("Stop retrying command %s after 3 tries", command.Name)

      break
    }

    cmd := exec.Command(command.Shell, "-c", command.Command)
    cmd.Env = env
    now := time.Now()
    id := GenerateId(8)
    if id == "" {
      logp.Err(errors.Errorf("Error generating new command id in command %s", command.Name).Error())
      tries = decrementAfterSleep(tries, SLEEP_TIME)
      continue
    }

    stderrChan, err := CreateAndReadAllFromFn(cmd.StderrPipe)
    if err != nil {
      logp.Err(errors.Wrapf(err, "Error in reading from stderr in command %s", command.Name).Error())
      tries = decrementAfterSleep(tries, SLEEP_TIME)
      continue
    }

    go func() {
      err := <- stderrChan
      if err != nil {
        logp.Err(errors.Wrapf(err, "Error in command %s, retrying...", command.Name).Error())
      }
    }()

    doneReading, err := ReadLineFromReaderFnAndPublish(cmd.StdoutPipe, client, &command, now, id)
    if err != nil {
      logp.Err(errors.Wrapf(err, "Unable to open stdout in command %s, retrying...", command.Name).Error())

      tries = decrementAfterSleep(tries, SLEEP_TIME)
      continue
    }

    lineRead, err := StartAndWaitCommand(cmd, doneReading)
    if err != nil {
      logp.Err(errors.Wrapf(err, "Error starting or waiting command %s after %d line read, retrying...", command.Name, lineRead).Error())

      tries = decrementAfterSleep(tries, SLEEP_TIME)
      continue
    }

    logp.Info(fmt.Sprintf("Command %s has sent %d lines", command.Name, lineRead))
    time.Sleep(command.Sleep)
  }

  sync <- struct{}{}
}

func StartAndWaitCommand(cmd *exec.Cmd, wait chan int64) (int64, error) {
  err := cmd.Start()
  if err != nil {
    return 0, errors.Wrap(err, "Error creating command")
  }

  lineRead := <- wait

  err = cmd.Wait()
  if err != nil {
    return lineRead, errors.Wrap(err, "Error executing command")
  }

  return lineRead, nil
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

func ReadLineFromReaderFnAndPublish(fn func() (io.ReadCloser, error), client beat.Client, command *Command, now time.Time, id string) (chan int64, error) {
  r, err := fn()
  if err != nil {
    return nil, errors.Wrap(err, "Error creating reader")
  }

  done := make(chan int64)

  reader := bufio.NewReader(r)
  go func() {

    var i int64 = 0

    for ;; i++ {
      line, err := reader.ReadString('\n')
      if err != nil {
        if err == io.EOF {
          break
        }

        logp.Err(errors.Wrapf(err, "Error reading line in command %s, killing command and retring...", command.Name).Error())

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
            "id": id,
            "name": command.Name,
            "started_at": now,
          },
        },
      })
    }

    done <- i
  }()

  return done, nil
}
