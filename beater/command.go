package beater

import (
  "fmt"
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

func (command Command) Run(events chan *Event, sync chan struct{}) {
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
      logp.Err(errors.Errorf("Error generating new command id in command %s id %d", command.Name, id).Error())
      tries = decrementAfterSleep(tries, SLEEP_TIME)
      continue
    }

    stderrChan, err := CreateAndReadAllFromFn(cmd.StderrPipe)
    if err != nil {
      logp.Err(errors.Wrapf(err, "Error in reading from stderr in command %s id %s", command.Name, id).Error())
      tries = decrementAfterSleep(tries, SLEEP_TIME)
      continue
    }

    go func() {
      err := <- stderrChan
      if err != nil {
        logp.Err(errors.Wrapf(err, "Error in command %s id %s, retrying...", command.Name, id).Error())
      }
    }()

    doneReading, err := ReadLineFromReaderFnAndPublish(cmd.StdoutPipe, &command, now, id, events)
    if err != nil {
      logp.Err(errors.Wrapf(err, "Unable to open stdout in command %s id %s, retrying...", command.Name, id).Error())

      tries = decrementAfterSleep(tries, SLEEP_TIME)
      continue
    }

    lineRead, err := StartAndWaitCommand(cmd, doneReading)
    if err != nil {
      logp.Err(errors.Wrapf(err, "Error starting or waiting command %s id %s after %d line read, retrying...", command.Name, id, lineRead).Error())

      tries = decrementAfterSleep(tries, SLEEP_TIME)
      continue
    }

    logp.Info(fmt.Sprintf("Command %s id %s has sent %d lines", command.Name, id, lineRead))
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

func ReadLineFromReaderFnAndPublish(fn func() (io.ReadCloser, error), command *Command, now time.Time, id string, events chan *Event) (chan int64, error) {
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


      events <- &Event{
        Fields: command.Fields,
        BeatEvent: common.MapStr{
          "line": line,
          "number": i,
          "id": id,
          "name": command.Name,
          "started_at": now,
        },
      }
    }

    done <- i
  }()

  return done, nil
}
