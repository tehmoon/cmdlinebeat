package beater

import (
  "os"
  "github.com/tehmoon/errors"
  "github.com/elastic/beats/libbeat/beat"
  "github.com/elastic/beats/libbeat/common"
  "strings"
  "time"
  "runtime"
)

type Cmdlinebeat struct {
  Commands []*Command `config:"commands"`
  Env map[string]string `config:"env"`
}

func (cmdlinebeat *Cmdlinebeat) Run(b *beat.Beat) (error) {
  client, err := b.Publisher.Connect()
  if err != nil {
    return errors.Wrap(err, "Error connecting to the publisher")
  }

  sync := make(chan struct{})
  events := make(chan *Event)

  go func() {
    for {
      event := <- events

      client.Publish(beat.Event{
        Timestamp: time.Now(),
        Fields: common.MapStr{
          "fields": event.Fields,
          "cmdlinebeat": event.BeatEvent,
        },
      })
    }
  }()

  for _, command := range cmdlinebeat.Commands {
    go command.Run(events, sync)
  }

  for range cmdlinebeat.Commands {
    <- sync
  }

  return nil
}

func New(b *beat.Beat, config *common.Config) (beat.Beater, error) {
  switch strings.ToLower(runtime.GOOS) {
    case "linux":
    case "darwin":
    case "freebsd":
    case "openbsd":
    case "netbsd":
    default:
      return nil, errors.Errorf("Operating system %s is not supported", runtime.GOOS)
  }

  cmdlinebeat := &Cmdlinebeat{
    Commands: make([]*Command, 0),
    Env: make(map[string]string),
  }

  err := config.Unpack(cmdlinebeat)
  if err != nil {
    return nil, errors.Wrap(err, "Error unpacking configuration")
  }

  for i, command := range cmdlinebeat.Commands {
    entryNumber := i + 1

    if command.Command == "" {
      return nil, errors.Errorf("Config #%d is missing a command entry", entryNumber)
    }

    if command.Name == "" {
      return nil, errors.Errorf("Config #%d is missing a name entry", entryNumber)
    }

    if command.Shell == "" {
      shell := os.Getenv("SHELL")
      if shell == "" {
        return nil, errors.Errorf("Config for command %s is missing a shell entry and SHELL environment variable is not found", command.Name)
      }

      command.Shell = shell
    }

    if command.Env == nil {
      command.Env = make(map[string]string)
    }

    command.uid, command.gid, err = GetUserGroupIds(command.User, command.Group)
    if err != nil {
      return nil, errors.Wrapf(err, "Config for command %s has an error in user or group field", command.Name)
    }

    for k, v := range cmdlinebeat.Env {
      if _, found := command.Env[k]; ! found {
        command.Env[k] = v
      }
    }

    command.entryNumber = entryNumber
  }

  return cmdlinebeat, nil
}

func (cmdlinebeat *Cmdlinebeat) Stop() {
  os.Exit(1)
}
