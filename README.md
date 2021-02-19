# watcher-daemon

Based on https://github.com/tamarakaufler/go-files-watcher. Version 2 of the same functionality.

Configuration of Daemon values is done through environment variables.

## Synopsis

Go implementation of a daemon for montoring file changes and running a command when a change is detected.

The configurable options are:

|                |                  |                default                                        |
|:---------------|:-----------------|:-------------------------------------------------------------:|
|  BasePath      |  WATCHER_DAEMON_BASE_PATH  |   current dir (directory that the watcher daemon starts monitoring) |
|  Extension     |  WATCHER_DAEMON_EXTENSION  |   .go (currently only one)                                    |
|  Command       |  WATCHER_DAEMON_COMMAND    |   echo "Hello world" (command to run upon detected change)    |
|  Excluded      |  WATCHER_DAEMON_EXCLUDED   |   none (comma separated strings/regexes specifying files to exclude) |                            |
|  Frequency     |  WATCHER_DAEMON_FREQUENCY  |   5 (sec) (repeat of the check)                               |

## Implementation

The base directory, file extension and exclusions (path, file name (wildcard character * can be used))
provide the check criteria, together with the frequency, at which the check run happens.

File information (path, file name, modification time) is collected into a list.
The list is processed and the file checks are parallelized, each running in a goroutine. When
the first change is detected, this particular run finishes, stopping the check of the rest
of the files and cancelling already running gouroutines.

Quality of the Go code is checked using the golangci-lint utility.

Makefile provides useful CLI commands for dev tasks:

  * deps
  * lint
  * test
  * cover
  * build
  * run
  * docker-run

### Containerization

  * For golangci-lint to work either gcc and libc-dev needs to be installed or CGO_ENABLED=0 must be set
  * To be able to use the -race flag when running tests, gcc and libc-dev must be installed

 image build for:

 * ENV GOOS=linux
 * ENV GOARCH=amd64

## Usage

a) without docker (example)

  * make all
  * WATCHER_DAEMON_EXCLUDED=internal/daemon/fixtures/basepath  WATCHER_DAEMON_FREQUENCY=3 make run

b) with docker (example)

  * docker build -t watcher-daemon:v1.0.0 .
  * docker run -w /basedir -v $PWD:/basedir --env WATCHER_DAEMON_EXCLUDED=vendor --env WATCHER_DAEMON_FREQUENCY=3 watcher-daemon:v1.0.0

      OR

  * docker run -w /basedir -v $PWD:/basedir --env-file watcher_daemon.env watcher-daemon:v1.0.0

  where watcher_daemon.env file contains:

  WATCHER_DAEMON_EXCLUDED=vendor
  WATCHER_DAEMON_FREQUENCY=3

c) using docker image in the Quay registry
    docker run -w /basedir -v $PWD:/basedir --env WATCHER_DAEMON_EXCLUDED=vendor --env WATCHER_DAEMON_FREQUENCY=3 quay.io/tamarakaufler/watcher-daemon:v1.0.0