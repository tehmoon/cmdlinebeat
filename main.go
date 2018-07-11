package main

import (
  "os"
  "./beater"
  "github.com/elastic/beats/libbeat/cmd"
)

var RootCmd = cmd.GenRootCmd("cmdlinebeat", "6.3.1", beater.New)

func main() {
  if err := RootCmd.Execute(); err != nil {
    os.Exit(1)
  }
}
