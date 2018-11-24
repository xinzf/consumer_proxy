package main

import (
	"context"
	log "github.com/sirupsen/logrus"
	"github.com/xinzf/consumer_proxy/server"
	"github.com/xinzf/consumer_proxy/server/config"
	"os"
	"os/signal"
	"runtime"
	"syscall"
)

const (
	PROJECT_NAME = "CONSUMER_PROXY"
	VERSION      = "1.1.0"
)

func init() {

}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	ex := make(chan bool)
	go run(ex)

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)

	<-ch
	close(ex)

	if err := server.Deregister(); err != nil {
		log.Errorln(err)
	}

	if err := server.Stop(); err != nil {
		log.Errorln(err)
	}
}

func run(ex chan bool) {

	ctx := context.WithValue(context.TODO(), "project_info", map[string]string{
		"name":    PROJECT_NAME,
		"version": VERSION,
	})

	if err := config.Init(ctx); err != nil {
		log.Panic(err)
	}

	server.Init(ctx)
	if err := server.Start(); err != nil {
		log.Panic(err)
	}

	if err := server.Register(); err != nil {
		log.Panic(err)
	}

	for {
		select {
		case <-ex:
			return
		}
	}
}
