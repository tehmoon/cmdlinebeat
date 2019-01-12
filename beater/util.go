package beater

import (
  "strconv"
	"os/exec"
  "github.com/tehmoon/errors"
  "time"
  "github.com/elastic/beats/libbeat/common"
  "fmt"
  "os"
  "crypto/rand"
  "io"
)

var (
  SLEEP_TIME = 5 * time.Second
  MAX_TRIES = 3
)

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

type Event struct {
  Fields common.MapStr
  BeatEvent common.MapStr
}

func GetUserGroupIds(userStr string) (uint32, uint32, error) {
  maxUint32 := uint32((1<<32) - 1)

  if userStr == "" {
    userStr = "nobody"
  }

  uid, err := getUserId(userStr)
  if err != nil {
    return maxUint32, maxUint32, errors.Wrap(err, "Error resolving user id")
  }

  gid, err := getGroupId(userStr)
  if err != nil {
    return maxUint32, maxUint32, errors.Wrap(err, "Error resolving group id")
  }

  return uint32(uid), uint32(gid), nil
}

func getUserId(userStr string) (uint32, error) {
	maxUint32 := uint32((1<<32) - 1)

	cmd := exec.Command("/usr/bin/id", "-u", userStr)
	output, err := cmd.CombinedOutput()
  if err != nil {
    return maxUint32, errors.Wrapf(err, "Error calling %q: %s", "id", string(output[:]))
  }

	if len(output) <= 1 {
		return maxUint32, errors.Errorf("Something went wrong with the command %q", "id -u")
	}

  uid, err := strconv.ParseUint(string(output[:len(output) - 1]), 10, 32)
  if err != nil {
    return maxUint32, errors.Wrap(err, "Error parsing uid")
  }

	return uint32(uid), nil
}

func getGroupId(userStr string) (uint32, error) {
	maxUint32 := uint32((1<<32) - 1)

	cmd := exec.Command("/usr/bin/id", "-g", userStr)
	output, err := cmd.CombinedOutput()
  if err != nil {
    return maxUint32, errors.Wrapf(err, "Error calling %q: %s", "id", string(output[:]))
  }

	if len(output) <= 1 {
		return maxUint32, errors.Errorf("Something went wrong with the command %q", "id -g")
	}

  gid, err := strconv.ParseUint(string(output[:len(output) - 1]), 10, 32)
  if err != nil {
    return maxUint32, errors.Wrap(err, "Error parsing gid")
  }

	return uint32(gid), nil
}

func IsRoot() (bool) {
  uid := os.Geteuid()

  if uid == 0 {
    return true
  }

  return false
}

type MaxRunningLocker struct {
  sync chan struct{}
  n uint16
}

func NewMaxRunningLocker(n uint16) (*MaxRunningLocker) {
  return &MaxRunningLocker{
    sync: make(chan struct{}, int64(n)),
    n: n,
  }
}

func (mrl MaxRunningLocker) Lock() {
  if mrl.n == 0 {
    return
  }

  mrl.sync <- struct{}{}
}

func (mrl MaxRunningLocker) Unlock() {
  if mrl.n == 0 {
    return
  }

  <- mrl.sync
}
