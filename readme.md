### spm - Simple Process Manager
spm is somewhere between systemd and [foreman](https://github.com/ddollar/foreman).

It runs as a daemon. You can start spm daemon by running `spm` command with no arguments. Once you've done that you are ready to start and manage your jobs(processes) with spm commands. spm reads its jobs from an extended Procfile syntax. You are eligible to start or stop jobs, see running jobs and their logs. Your processes executed as shell commands so you're allowed to use shell syntax in your Procfile. 

Our extended version of Procfile syntax supports comments with `#` sign and multilines with `\`.

## Why spm
All other foreman like process managers don't have a proper stop feature to end running jobs which spm has. spm works in a client/server convention, spm daemon listens for job commands(listed at `spm -h`) from it's clients through unix sockets. You are also able to use multiple Procfiles and start/stop multiple jobs in a Procfile which gives a huge flexiblity when you need to work with bunch of long running processes. Jobs are recognized by their names, make sure that your jobs has unique names otherwise spm will not start an already running job with same name that you started before.

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

## Example

Following Procfile has two jobs (tabs for legibility): 

```
# Serve spm's github repository on local machine
repo:   rm -rf spm && \
        git clone https://github.com/bytegust/spm.git && \
        http-server ./spm -p 8080

# Download and serve a webpage on local machine
apod: \

        # Create directory that will be served
        rm -rf astropix && \
        mkdir astropix && \
        cd astropix && \

        # Downloads the content of https://apod.nasa.gov/apod/astropix.html
        wget -A.jpg -e robots=off -k \ 
        --user-agent="Mozilla/5.0 (compatible; Konqueror/3.0.0/10; Linux)" \
        --no-check-certificate https://apod.nasa.gov/apod/astropix.html && \

        # Rename file as index
        mv $(ls *.html | head -1) index.html && \

        # Serve webpage
        http-server -p 8081
```

Suppose that we have the Procfile above inside a folder named _test_. After starting the daemon by `spm` command we will be able to run jobs, inside our Procfile, from the clients, namely, other terminal windows or tabs.

1. Run `spm` command to start daemon:

    ```
    $ spm
    2016/11/07 22:20:22 spm.go:209: deamon started
    ```

    ![](https://cloud.githubusercontent.com/assets/7649229/20076221/4d9cb57e-a540-11e6-911b-d5b33722b131.png)

1. Run `spm start` (or a specific job e.g. `spm start apod`) command from different terminal tab. Since our Procfile reside in the folder named _test_, we have to define its path using `-f` flag:

    ```
    $ spm -f test start
    2016/11/07 22:20:23 spm.go:122: done
    ```

    ![](https://cloud.githubusercontent.com/assets/7649229/20076317/a00a0370-a540-11e6-9bc8-21640f097168.png)

1. Log files are being saved under /tmp folder with a job specific name, so that it's possible to see logs by `spm logs <jobname>` command:

    ```
    $ spm logs apod
    ```

1. List all running jobs using `spm list` command:

    ```
    $ spm list
    ```
    
1. Stop running jobs using `spm stop` command (or a specific job e.g. `spm stop apod`):

    ```
    $ spm stop
    ```

    ![](https://cloud.githubusercontent.com/assets/7649229/20076337/b64b1b42-a540-11e6-9a39-80a3235d696c.png)
