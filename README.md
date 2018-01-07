# cmdlinebeat
A beat that will execute commands periodically and send every line as a single event to the output

## Disclamer
This program executes commands and send the output to whatever you configured.
Cmdlinebeat doesn't do any security checks nor sandboxing, so be really careful to
**not run this as root** and remember that:

  - what ever you execute can impact your system -- or remote systems
  - what ever you execute can leak sensitive informations
  - what ever you execute will be executed with the user that runs `cmdlinebeat` so make sure the user has a restricted access to what it should have
  - running `cmdlinebeat` in a container with the `root` user **must still be avoided**

I'm planing on dropping the privileges on every command so it is safe to run, until then please respect the guidelines.

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
cmdlinebeat.commands:
  - command: ls      # Command you want to run in a shell.
    sleep: 0s        # Sleep for x after running command. By default the command will be re-executed just after so you might want to put some sleep if the command is fast.
    fields:          # Extra data to be merged with .fields field.
      command: tata  # Usefull to "tag" the event since the command is not exposed to the event for security reasons and you cant find it back in elasticsearch
    shell: ${SHELL}  # Use the shell specified if the environment variable SHELL is not found
    env:             # additional environment variable to pass to the child process
    copy_env: false  # copy all environment to child process
    timeout: 0       # NYI: kill process if timeout is reached
```

## Example Configuration

```
cmdlinebeat.commands:
# Will get the size of all directories at depth=1 every hour from my_test_bucket
  - command: s3-du -d 1 -b my_test_bucket
    sleep: 1h
    fields:
      command: s3_du_my_test_bucket_1h
    env:
      AWS_PROFILE: dev
    shell: /bin/sh
# Will get the size of all directories at depth=0 every half hour from my_test_bucket
  - command: s3-du -d 0 -b my_test_bucket
    sleep: 30m
    fields:
      command: s3_du_my_test_bucket_30m
    env:
      AWS_PROFILE: dev
    shell: /bin/sh
```

## Output Fields:

```
{
  "line": string   // line output from the command
  "number": number // line number from the output
}
```
