package daemon

import (
	"context"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/caarlos0/env/v6"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// WatcherDaemon specifies what methods must be implemented
type WatcherDaemon interface {
	Watch(ctx context.Context, sigCh chan os.Signal)
}

// verifying a Daemon implements all required methods, ie is a WatcherDaemon
var _ (WatcherDaemon) = (*Daemon)(nil)

// Daemon contains configuration for running the watcher
type Daemon struct {
	BasePath  string `env:"WATCHER_DAEMON_BASE_PATH" envDefault:"."`
	Extention string `env:"WATCHER_DAEMON_EXTENSION" envDefault:".go"`
	Excluded  string `env:"WATCHER_DAEMON_EXCLUDED" envDefault:""`   // provided as a comma separated string
	Frequency string `env:"WATCHER_DAEMON_FREQUENCY" envDefault:"5"` // run frequency in seconds

	excluded  []string
	frequency time.Duration

	logger   *logrus.Logger
	LogLevel string `env:"WATCHER_DAEMON_LOG_LEVEL" envDefault:""`

	// mutex protects sending on the doneChan
	doneMux  *sync.Mutex
	doneChan chan struct{}

	// mutex protects running of the command
	cmdMux  *sync.Mutex
	Command string `env:"WATCHER_DAEMON_COMMAND" envDefault:"echo \"Hello world\""`
}

// New is a constructor providing a new instance of a Daemon
func New() (*Daemon, error) {
	d := &Daemon{}
	err := env.Parse(d)
	if err != nil {
		return nil, errors.Wrap(err, "error creating a Daemon instance")
	}

	d.excluded = strings.Split(d.Excluded, ",")

	fr, err := strconv.Atoi(d.Frequency)
	if err != nil {
		return nil, err
	}
	d.frequency = time.Duration(time.Duration(fr) * time.Second)

	d.initialiseLogger()

	d.cmdMux = &sync.Mutex{}
	d.doneMux = &sync.Mutex{}

	d.doneChan = make(chan struct{})

	return d, err
}

// Watch watches for changes in files at regular intervals
func (d *Daemon) Watch(ctx context.Context, sigCh chan os.Signal) {
	d.logger.Infof("\nStarting the watcher daemon ⌚ 👀 ... \n\n")

	cmdParts := strings.Split(d.Command, " ")

	// use when a change is detected to avoid processing further files
	doneCh := make(chan struct{})
	// use when a change is detected, after successfully running the command,
	// to cancel already created goroutines
	cancelCh := make(chan struct{})

	// Starts a gouroutine checking on the run outcome, running the command as required
	d.runOutcomeChecker(cmdParts, sigCh, doneCh, cancelCh)

	tick := time.NewTicker(d.frequency)
	for {
		ctxR, cancel := context.WithCancel(ctx)
		select {
		case <-tick.C:
			files, err := d.CollectFiles(ctxR)
			if err != nil {
				d.logger.Warn(err)
				continue
			}

			//nolint:lll
			// Creating a buffered channel will avoid leaking goroutines. This would  // happen if there are still running goroutines after one finds a change
			// and sends to a done channel. If the done channel is not buffered, then
			// some of the running gouroutines may also try to send to the done
			// channel and would be blocked forever, ie would start leaking.
			doneAllCh := make(chan struct{}, len(files))

			go func(doneAllCh, doneCh chan struct{}) {
				select {
				case <-doneAllCh:
					d.doneMux.Lock()
					doneCh <- struct{}{}
					d.doneMux.Unlock()
				case <-time.After(d.frequency * 2):
					return
				}
			}(doneAllCh, doneCh)

			d.ProcessFilesInParallel(ctxR, files, doneAllCh)
		case <-cancelCh:
			cancel()
		}
	}
}
