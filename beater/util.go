package beater

import (
  "time"
  "fmt"
  "os"
  "crypto/rand"
  "io"
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

func GenerateId(n int64) (string) {
  if n < 1 {
    return ""
  }

  n = n / 2

  buff := make([]byte, n)

  _, err := io.ReadFull(rand.Reader, buff)
  if err != nil {
    return ""
  }

  return fmt.Sprintf("%x", buff)
}
