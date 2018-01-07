package beater

import (
  "time"
)

var (
  SLEEP_TIME = 5 * time.Second
)

func decrementAfterSleep(i int, sleep time.Duration) (int) {
  time.Sleep(sleep)

  i--
  return i
}
