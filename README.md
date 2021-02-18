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

Tests are provided.

Quality of the Go code is checked using the golangci-lint utility.

## Usage

make all

WATCHER_DAEMON_EXCLUDED=internal/daemon/fixtures/basepath  WATCHER_DAEMON_FREQUENCY=3 make run
