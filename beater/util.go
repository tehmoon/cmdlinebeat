package beater

import (
  "strconv"
  "os/user"
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

func GetUserGroupIds(userStr, groupStr string) (uint32, uint32, error) {
  maxUint32 := uint32((1<<32) - 1)

  if userStr == "" {
    userStr = "nobody"

    if groupStr == "" {
      groupStr = "nobody"
    }
  }

  u, err := getUserId(userStr)
  if err != nil {
    return maxUint32, maxUint32, errors.Wrap(err, "Error resolving user")
  }

  if groupStr == "" {
    groupStr = u.Gid
  }

  g, err := getGroupId(groupStr)
  if err != nil {
    return maxUint32, maxUint32, errors.Wrap(err, "Error resolving group")
  }

  ok, err := checkUserInGroup(u, g)
  if err != nil {
    return maxUint32, maxUint32, errors.Wrap(err, "Error checking if user belongs to group")
  }

  if ! ok {
    return maxUint32, maxUint32, errors.New("User doesn't belong to group")
  }

  uid, err := strconv.ParseUint(u.Uid, 10, 32)
  if err != nil {
    return maxUint32, maxUint32, errors.Wrap(err, "Error parsing uid")
  }

  gid, err := strconv.ParseUint(g.Gid, 10, 32)
  if err != nil {
    return maxUint32, maxUint32, errors.Wrap(err, "Error parsing gid")
  }

  return uint32(uid), uint32(gid), nil
}

func getUserId(userStr string) (*user.User, error) {
  u, err := user.LookupId(userStr)
  if err == nil {
    return u, nil
  }

  u, err = user.Lookup(userStr)
  if err != nil {
    return nil, err
  }

  return u, nil
}

func getGroupId(groupStr string) (*user.Group, error) {
  g, err := user.LookupGroupId(groupStr)
  if err == nil {
    return g, nil
  }

  g, err = user.LookupGroup(groupStr)
  if err != nil {
    return nil, err
  }

  return g, nil
}

func checkUserInGroup(u *user.User, g *user.Group) (bool, error) {
  gids, err := u.GroupIds()
  if err != nil {
    return false, err
  }

  for _, gid := range gids {
    if gid == g.Gid {
      return true, nil
    }
  }

  return false, nil
}

func IsRoot() (bool) {
  u, err := user.Current()
  if err != nil {
    return false
  }

  return u.Uid == "0"
}
