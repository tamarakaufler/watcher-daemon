package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/tamarakaufler/watcher-daemon/internal/daemon"
)

func main() {

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)

	d := daemon.New(
		daemon.WithCommand("echo \"Hello world\""),
		//daemon.WithCommand("go build -o watcher-daemon cmd/watcher-daemon/main.go"))
		daemon.WithExcluded([]string{"internal/daemon/fixtures/*"}),
		daemon.WithFrequency(5),
	)

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(d.Frequency)*time.Second)
	defer cancel()
	d.Watch(ctx, sigCh)
}
