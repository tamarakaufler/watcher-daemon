package daemon

import (
	"sync"
	"time"
)

// Daemon contains configuriation for running the watcher
type Daemon struct {
	BasePath  string
	Extention string
	Excluded  []string
	Frequency int32
	frequency time.Duration

	// mutex protects sending on the doneChan
	doneMux  *sync.Mutex
	doneChan chan struct{}

	// mutex protects running of the command
	cmdMux  *sync.Mutex
	Command string
}

// Option provides a way to customise the
type Option func(*Daemon)

// New is a constructor providing a new instance of a Daemon
func New(ops ...Option) *Daemon {
	f := int32(15)
	d := &Daemon{
		BasePath:  ".",
		Extention: ".go",
		Excluded:  []string{},
		Frequency: f,
		frequency: time.Duration(time.Duration(f) * time.Second),

		cmdMux:  &sync.Mutex{},
		Command: "echo \"Hello world\"",

		doneMux:  &sync.Mutex{},
		doneChan: make(chan struct{}),
	}

	for _, o := range ops {
		o(d)
	}

	return d
}

// WithBasePath allows to override default BasePath configuration.
func WithBasePath(bp string) Option {
	return func(d *Daemon) {
		d.BasePath = bp
	}
}

// WithExtension allows to override default file extension configuration.
func WithExtension(ex string) Option {
	return func(d *Daemon) {
		d.Extention = ex
	}
}

// WithCommand allows to override default configuration of a command
// to run when a file change is detected.
func WithCommand(c string) Option {
	return func(d *Daemon) {
		d.Command = c
	}
}

// WithExcluded allows to provide a list of paths to exclude.
func WithExcluded(ex []string) Option {
	return func(d *Daemon) {
		d.Excluded = ex
	}
}

// WithFrequency allows to override default frequency configuration.
func WithFrequency(f int32) Option {
	return func(d *Daemon) {
		d.Frequency = f
		d.frequency = time.Duration(time.Duration(f) * time.Second)
	}
}
