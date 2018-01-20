package beater

import (
  "time"
  "fmt"
  "os"
)

var (
  SLEEP_TIME = 5 * time.Second
)

func decrementAfterSleep(i int, sleep time.Duration) (int) {
  time.Sleep(sleep)

  i--
  return i
}

func ForkEnv(env map[string]string, inherit bool) ([]string) {
  newEnv := make([]string, 0)

  if inherit {
    newEnv = os.Environ()
  }

  for k, v := range env {
    newEnv = append(newEnv, fmt.Sprintf("%s=%s", k, v))
  }

  return newEnv
}
