## spm - Simple Process Manager
Similar to Foreman (https://github.com/ddollar/foreman) but with _stop_ feature. Create a Procfile to define your jobs.

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

Documentation can be found at https://github.com/bytegust/tools
```