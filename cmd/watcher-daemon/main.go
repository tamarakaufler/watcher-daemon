package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/tamarakaufler/watcher-daemon/internal/daemon"
)

func main() {

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)

	d, err := daemon.New()
	if err != nil {
		log.Panic(err)
	}

	fr, err := strconv.Atoi(d.Frequency)
	if err != nil {
		log.Panic(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(fr)*time.Second)
	defer cancel()
	d.Watch(ctx, sigCh)
}
