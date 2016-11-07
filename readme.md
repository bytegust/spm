## spm - Simple Process Manager
spm is somewhere between systemd and foreman.

It runs as a daemon. You can start spm daemon by running `spm` command with no arguments. Once you done that you are ready to start and manage your jobs(processes) with spm commands. spm reads it's jobs from an extended Procfile syntax. You are eligible to start or stop jobs, see running jobs and their logs. Your processes executed as shell commands so you're allowed to use shell syntax in your Procfile. 

Our extended version of Procfile syntax supports comments with `#` sign and multilines with `\`.

### Usage
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