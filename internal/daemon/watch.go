package daemon

import (
	"context"
	"fmt"
	"io"
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

// CollectFiles checks if any watched file has changed.
// The Walk function continues the walk while theere is no error and stops
// when the filepath.WalkFunc exits with error.
//nolint:unused,errcheck
func (d *Daemon) walkThroughFiles(ctx context.Context, doneCh chan struct{}) {
	filepath.Walk(d.BasePath, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() ||
			strings.HasPrefix(path, ".git") ||
			(!info.IsDir() && filepath.Ext(path) != d.Extention) {
			return nil
		}

		fmt.Printf("FILE info:  %s\n", info.Name())

		lastChecked := time.Now().Add(-d.frequency)
		if info.ModTime().After(lastChecked) {
			fmt.Printf("\tFile %s has changed\n", info.Name())
			// trigger running of the command
			doneCh <- struct{}{}
			// return any known error to stop walking through the dir content
			return io.EOF
		}
		return nil
	})
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

		if len(d.Excluded) != 0 {
			isExcl, err := d.IsExcluded(ctx, path, info.Name())
			if err != nil {
				panic(errors.Wrap(err, "cannot proccess exclusion of files"))
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
		//fmt.Printf("FILE info:  %s - %s\n", path, info.Name())

		return nil
	})
	if err != nil {
		return nil, errors.Wrapf(err, "error collecting files from %s", d.BasePath)
	}

	return files, nil
}

//nolint:unused
func (d *Daemon) processFiles(ctx context.Context, files []FileInfo, doneCh chan struct{}) {

	fmt.Println("GOT to processing ...")

	for _, f := range files {
		time.Sleep(100 * time.Millisecond)

		lastChecked := time.Now().Add(-d.frequency)
		if f.ModTime.After(lastChecked) {
			fmt.Printf("File %s has changed\n", f.Name)
			doneCh <- struct{}{}
			break
		}
	}
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
	fmt.Println("---------------")
LOOP:
	for _, f := range files {
		fmt.Printf("--> processing file %s\n", f.Name)

		wg.Add(1)
		go func(wg *sync.WaitGroup, f FileInfo, doneCh chan struct{}, stopCh chan struct{}) {
			defer wg.Done()
			time.Sleep(100 * time.Millisecond)

			lastChecked := time.Now().Add(-d.frequency)
			if f.ModTime.After(lastChecked) {
				fmt.Printf("File %s has changed\n", f.Name)
				stopCh <- struct{}{}
				return
			}
			continueCh <- struct{}{}
		}(wg, f, doneCh, stopCh)

		select {
		case <-stopCh:
			doneCh <- struct{}{}
			fmt.Printf("\t--> finishing with file %s\n\n", f.Name)
			break LOOP
		case <-continueCh:
		}
	}
	fmt.Println("---------------")

	wg.Wait()
}

// IsExcluded filters files based on custom exclusion configuration
func (d *Daemon) IsExcluded(ctx context.Context, path, name string) (bool, error) {
	toExclude := false

	for _, ex := range d.excluded {
		// deal with regex
		if strings.ContainsAny(ex, "*?{}[]()+") {
			r, err := regexp.Compile(ex)
			if err != nil {
				return false, errors.Wrap(err, "cannot exclude files")
			}
			if r.MatchString(path) {
				return true, nil
			}
			// deal with exact matches
		} else if name == ex || path == ex {
			return true, nil
		}

	}
	return toExclude, nil
}

func (d *Daemon) runOutcomeChecker(cmdParts []string, sigCh chan os.Signal, doneCh, cancelCh chan struct{}) {
	go func() {
		for {
			select {
			case <-sigCh:
				fmt.Println("You interrupted me 👹!")
				os.Exit(0)
			case <-doneCh:
				d.cmdMux.Lock()

				cmd := exec.Command(cmdParts[0], cmdParts[1:]...)
				// these can be commented out if not needed
				cmd.Stdout = os.Stdout
				cmd.Stderr = os.Stderr
				err := cmd.Run()
				if err != nil {
					fmt.Printf("ERROR: %s\n", errors.Wrap(err, "error occurred processing during file watch"))
					cancelCh <- struct{}{}
					d.cmdMux.Unlock()
					continue
				}
				fmt.Print("command completed successfully\n\n")
				d.cmdMux.Unlock()
			}
		}
	}()
}
