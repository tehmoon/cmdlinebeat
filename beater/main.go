package beater

import (
  "os"
  "github.com/tehmoon/errors"
  "github.com/elastic/beats/libbeat/beat"
  "github.com/elastic/beats/libbeat/common"
  "strings"
  "runtime"
)

type Cmdlinebeat struct {
  Commands []*Command `config:"commands"`
}

func (cmdlinebeat *Cmdlinebeat) Run(b *beat.Beat) (error) {
  sync := make(chan struct{})

  for _, command := range cmdlinebeat.Commands {
    go command.Run(b, sync)
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
        return nil, errors.Errorf("Config #%d is missing a shell entry and SHELL environment variable is not found", entryNumber)
      }

      command.Shell = shell
    }

    command.entryNumber = entryNumber
  }

  return cmdlinebeat, nil
}

func (cmdlinebeat *Cmdlinebeat) Stop() {
  os.Exit(1)
}
