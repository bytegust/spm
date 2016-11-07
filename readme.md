### spm - Simple Process Manager
spm is somewhere between systemd and [foreman](https://github.com/ddollar/foreman).

It runs as a daemon. You can start spm daemon by running `spm` command with no arguments. Once you done that you are ready to start and manage your jobs(processes) with spm commands. spm reads it's jobs from an extended Procfile syntax. You are eligible to start or stop jobs, see running jobs and their logs. Your processes executed as shell commands so you're allowed to use shell syntax in your Procfile. 

Our extended version of Procfile syntax supports comments with `#` sign and multilines with `\`.

## Why spm
All other foreman like process managers doesn't have a proper stop feature to end running jobs which spm has. spm works in a client/server convention, spm daemon listens for job commands(listed at `spm -h`) from it's clients through unix sockets. You are also able to use multiple Procfiles and start/stop multiple jobs in a Procfile which gives a huge flexiblity when you need to work with bunch of long running processes. Jobs are recognized by their names, make sure that your jobs has unique names otherwise spm will not start an already running job with same name that you started before.

## Installation and Usage
```
$ go get github.com/bytegust/spm/cmd/spm
$ spm -h

Simple Process Manager.

Usage: spm [OPTIONS] COMMAND [arg...]
       spm [ --help ]

Options:

  -f, --file=~/.Procfile    Location of Procfile
  -h, --help                Print usage

Commands:

    start           Start all jobs defined in Procfile
    start <jobs>    Starts jobs if present in Procfile
    stop            Stop all current jobs
    stop <jobs>     Stop jobs if currently running
    list            Lists all running jobs
    logs <job>      Prints last 200 lines of job's logfile 

Documentation can be found at https://github.com/bytegust/spm
```