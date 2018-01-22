# cmdlinebeat
A beat that will execute commands periodically and send every line as a single event to the output

## Disclamer
This program executes commands and send the output to whatever you configured.

`Cmdlinebeat` doesn't do any security checks nor sandboxing, so remember that:

  - what ever you execute can impact your system -- or remote systems
  - what ever you execute can leak sensitive informations
  - what ever you execute will be executed with the user that runs `cmdlinebeat` so make sure the user has a restricted access to what it should have
  - running `cmdlinebeat` in a container as `root` just because you're in a container will come back haunting you
  - when running `cmdlinbeat` as `root` it will by default executes everything as `nobody`. Use only `root` if you have multiple command to run
    with multiple different user

## Installation
There are two ways you can install cmdlinebeat:

  - From the [release page](https://github.com/tehmoon/cmdlinebeat/releases)
  - From the source -- requires Go:
```
git clone https://github.com/tehmoon/cmdlinebeat
cd cmdlinebeat
go get ./...
go build # A binary name cmdlinebeat will be generated in the directory
```

## Run

```
./cmdlinebeat -c cmdlinebeat.yml
```

## Configuration file

```
cmdlinebeat.env:                # Global variable to pass to commands
cmdlinebeat.max-running: uint16 # Max number of command to execute in parrallele
cmdlinebeat.commands:
  - command: ls /      # Command you want to run in a shell.
    name: ls_slash_1_s # Name of you command
    sleep: 1s          # Sleep for x after running command. By default the command will be re-executed just after so you might want to put some sleep if the command is fast.
    fields:            # Extra data to be merged with .fields field.
    shell: ${SHELL}    # Use the shell specified if the environment variable SHELL is not found
    env:               # additional environment variable to pass to the child process. Will override cmdlinebeat.env if found.
    copy_env: false    # copy all environment to child process
    timeout: 0         # NYI: kill process if timeout is reached
    user: user/uid     # Username to execute the command with. Require root privs. Default to "nobody"
    group: group/gid   # Groupname to execute the group with. Require root privs, default to main user group
```

## Example Configuration

```
cmdlinebeat.env:
  AWS_PROFILE: dev
cmdlinebeat.max-running: 1
cmdlinebeat.commands:
# Will get the size of all directories at depth=1 every hour from my_test_bucket
  - command: s3-du -d 1 -b my_test_bucket
    sleep: 1h
    name: s3_du_my_test_bucket_1h
    shell: /bin/sh
# Will get the size of all directories at depth=0 every half hour from my_test_bucket
  - command: s3-du -d 0 -b my_test_bucket
    sleep: 30m
    name: s3_du_my_test_bucket_30m
    shell: /bin/sh
```

## Output Fields:

```
{
  "line": string,     // line output from the command
  "number": number,   // line number from the output
  "name": string,     // name of the command
  "id": string,       // uniq id per command
  "started_at": time, // when the command started executing
  "status": string    // when the command is executing status is at "running", when it is done it is at "stopped"
}
```
