package daemon

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/pkg/errors"
)

// FileInfo captures file path, name and modification time.
// This information is required for the watch functionality.
type FileInfo struct {
	Path    string
	Name    string
	ModTime time.Time
}

// CollectFiles checks if any watched file has changed
func (d *Daemon) CollectFiles(ctx context.Context) ([]FileInfo, error) {
	var files []FileInfo

	err := filepath.Walk(d.BasePath, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() ||
			strings.HasPrefix(path, ".git") ||
			(!info.IsDir() && filepath.Ext(path) != d.Extention) {
			return err // this will be nil if there is no problem with the file
		}

		if len(d.excluded) != 0 {
			isExcl, err := d.IsExcluded(ctx, path, info.Name())
			if err != nil {
				d.logger.Panic(errors.Wrap(err, "cannot proccess exclusion of files"))
			}
			if isExcl {
				return nil
			}
		}

		files = append(files, FileInfo{
			Path:    path,
			Name:    info.Name(),
			ModTime: info.ModTime(),
		})
		return nil
	})

	if err != nil {
		return nil, errors.Wrapf(err, "error collecting files from %s", d.BasePath)
	}
	return files, nil
}

// ProcessFilesInParallel checks files in parallel.
func (d *Daemon) ProcessFilesInParallel(ctx context.Context, files []FileInfo, doneCh chan struct{}) {
	wg := &sync.WaitGroup{}

	stopCh := make(chan struct{})
	continueCh := make(chan struct{})

	// Files are checked in parallel. When a change is found, a message is sent
	// to the doneCh channel to interrupt the looping through the rest of the
	// files. When no chenge is found, a message is sent to the continueCh
	// channel to continue looping.
	// Note: I tried to use select default to continue the looping but that
	// did not work.
	d.logger.Infof("---------------")
LOOP:
	for _, f := range files {
		d.logger.Infof(">>> processing file %s", f.Path)

		wg.Add(1)
		go func(wg *sync.WaitGroup, f FileInfo, doneCh chan struct{}, stopCh chan struct{}) {
			defer wg.Done()
			time.Sleep(100 * time.Millisecond)

			lastChecked := time.Now().Add(-d.frequency)
			if f.ModTime.After(lastChecked) {
				d.logger.Infof("File %s has changed", f.Name)
				stopCh <- struct{}{}
				return
			}
			continueCh <- struct{}{}
		}(wg, f, doneCh, stopCh)

		select {
		case <-stopCh:
			doneCh <- struct{}{}
			d.logger.Debugf("\t--> finishing with file %s", f.Name)
			break LOOP
		case <-continueCh:
		}
	}
	d.logger.Infof("---------------")

	wg.Wait()
}

// IsExcluded filters files based on custom exclusion configuration
func (d *Daemon) IsExcluded(ctx context.Context, path, name string) (bool, error) {
	toExclude := false

	for _, ex := range d.excluded {
		if ex == "" {
			continue
		}

		// deal with regex
		if strings.ContainsAny(ex, "*?{}[]()+") {

			r, err := regexp.Compile(ex)
			if err != nil {
				return false, errors.Wrap(err, "cannot exclude files")
			}
			if r.MatchString(path) {
				return true, nil
			}
			// deal with string matches
		} else {
			if name == ex || path == ex || strings.Contains(path, ex) {
				return true, nil
			}
		}
		toExclude = false
	}

	return toExclude, nil
}

func (d *Daemon) runOutcomeChecker(cmdParts []string, sigCh chan os.Signal, doneCh, cancelCh chan struct{}) {
	go func() {
		for {
			select {
			case <-sigCh:
				d.logger.Info("You interrupted me 👹!")
				os.Exit(0)
			case <-doneCh:
				d.cmdMux.Lock()

				cmd := exec.Command(cmdParts[0], cmdParts[1:]...)
				// these can be commented out if not needed
				cmd.Stdout = os.Stdout
				cmd.Stderr = os.Stderr
				err := cmd.Run()
				if err != nil {
					d.logger.Errorf("%s", errors.Wrap(err, "error occurred processing during file watch"))
					cancelCh <- struct{}{}
					d.cmdMux.Unlock()
					continue
				}
				d.logger.Info("command completed successfully")
				d.cmdMux.Unlock()
			}
		}
	}()
}
