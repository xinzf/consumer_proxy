package mq

import "context"

type State int32

const (
	ABNORMAL State = iota - 1
	STOPPED
	STARTING
	RUNNING
)

func (c State) String() (txt string) {
	switch c {
	case STARTING:
		txt = "Starting"
	case STOPPED:
		txt = "Stopped"
	case RUNNING:
		txt = "Running"
	}
	return
}

var (
	Options = struct {
		Brokers []string
		User    string
		Pswd    string
		Vhost   string
	}{
		Brokers: []string{"127.0.0.1:5672", "127.0.0.1:5673", "127.0.0.1:5674"},
		User:    "guest",
		Pswd:    "guest",
		Vhost:   "/",
	}
	Connection *amqpConnection
)

func Init(ctx context.Context) error {
	Connection = &amqpConnection{
		afterConnected: make([]func(connection *amqpConnection), 0),
		afterClosed:    make([]func(connection *amqpConnection), 0),
	}
	return nil
}
