package daemon

import (
	"os"

	"github.com/sirupsen/logrus"
)

var (
	defaultLogLevel = logrus.InfoLevel
)

func (d *Daemon) initialiseLogger() {
	defaultLogger := logrus.New()

	defaultLogger.SetLevel(defaultLogLevel)
	if d.LogLevel != "" {
		logL, err := logrus.ParseLevel(d.LogLevel)
		if err != nil {
			defaultLogger.SetLevel(logL)
		}
	}

	output := os.Stdout
	defaultLogger.SetOutput(output)

	defaultFields := logrus.Fields{
		"service":   "watcher-daemon",
		"base_dir":  d.BasePath,
		"frequency": d.Frequency,
		"excluded":  d.Excluded,
		"extension": d.Extention,
	}
	defaultLogger.WithFields(defaultFields)

	d.logger = defaultLogger
}
