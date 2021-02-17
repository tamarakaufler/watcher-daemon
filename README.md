# watcher-daemon

Based on https://github.com/tamarakaufler/go-files-watcher. Version 2 of the same functionality.

### TODO
Add more details

## Synopsis

Go implementation of a daemon for montoring file changes and running a command when a change is detected.

The configurable options are:

|                |                  |                default                                        |
|:---------------|:-----------------|:-------------------------------------------------------------:|
|  BasePath      |  string          |   current dir (directory that the watcher daemon starts monitoring) |
|  Extension     |  string          |   .go (currently only one)                                    |
|  Command       |  string          |   echo "Hello world" (command to run upon detected change)    |
|  Excluded      |  list of strings |   none (a list of strings/regexes specifying files to exclude) |                            |
|  Frequency     |  int32           |   5 (sec) (repeat of the check)                               |

## Implementation

There are 3 progressive implementations, from the initial one using directly filepath.Walk,
an intemediate one as a preparation for the third parallelized third implementation. First two
versions are commented out (in the (*Daemon).Watch method).

### Details

The base directory, file extension and exclusions (path, file name (wildcard character * can be used))
provide the check criteria, together with the frequency, at which the check run happens.

File information (path, file name, modification time) is collected into a list.
The list is processed and the file checks are parallelized, each running in a goroutine. When
the first change is detected, this particular run finishes, stopping the check of the rest
of the files and cancelling already running gouroutines.

Tests are provided.

Quality of the Go code is checked using the golangci-lint utility.
