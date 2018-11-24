package mq

import (
	"context"
	"errors"
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
	"sync"
	"sync/atomic"
	"time"
)

type amqpConnection struct {
	status         int32
	conn           *amqp.Connection
	ctx            context.Context
	cancle         context.CancelFunc
	once           sync.Once
	afterConnected []func(connection *amqpConnection)
	afterClosed    []func(connection *amqpConnection)
}

func (this *amqpConnection) Start(ctx context.Context) {
	defer func() {
		if err := recover(); err != nil {
			switch err.(type) {
			case error:
				log.Panicln("RabbitMQ connection has disconnected with error: ", err.(error).Error())
			}
		}
	}()

	ctx, cancle := context.WithCancel(ctx)
	this.ctx = ctx
	this.cancle = cancle

	for this.Status() == STOPPED {
		this.connect()

		if this.Status() == STOPPED {
			time.Sleep(2 * time.Second)
		}
	}

	notify := make(chan *amqp.Error)
RETRY:
	for {
		select {
		case <-this.ctx.Done():
			this.close()
		case err, flag := <-this.conn.NotifyClose(notify):
			atomic.StoreInt32(&this.status, int32(STOPPED))

			for _, fn := range this.afterClosed {
				fn(this)
			}
			//Jobbers.StopAll()

			if !flag {
				log.Errorf("RabbitMQ connection has went away")
				return
			}

			if err != nil {
				log.Errorln("Closed by notify from RabbitMQ with error: ", err)
				break RETRY
			} else {
				log.Infoln("Closed by notify from RabbitMQ")
				return
			}
		}
	}

	go this.Start(this.ctx)
}

func (this *amqpConnection) Stop() error {
	this.close()
	return nil
}

func (this *amqpConnection) Status() State {
	status := atomic.LoadInt32(&this.status)
	return State(status)
}

func (this *amqpConnection) AfterConnected(fn ...func(connection *amqpConnection)) {
	this.afterConnected = append(this.afterConnected, fn...)
}

func (this *amqpConnection) AfterClosed(fn ...func(connection *amqpConnection)) {
	this.afterClosed = append(this.afterClosed, fn...)
}

func (this *amqpConnection) getChannel() (*amqp.Channel, error) {
	if this.Status() != RUNNING {
		return nil, errors.New("RabbitMQ has not connected.")
	}
	channel, err := this.conn.Channel()
	return channel, err
}

func (this *amqpConnection) connect() error {
	log.Info("Try to connect rabbitMQ server.")
	atomic.StoreInt32(&this.status, int32(STARTING))

	dial := func(addr string) (*amqp.Connection, error) {
		u := fmt.Sprintf(
			"amqp://%s:%s@%s%s",
			Options.User,
			Options.Pswd,
			addr,
			Options.Vhost,
		)
		conn, err := amqp.Dial(u)
		return conn, err
	}

	var conn *amqp.Connection
	var err error
	var connectHost string
	for _, host := range Options.Brokers {
		conn, err = dial(host)
		if err == nil {
			this.conn = conn
			connectHost = host
			break
		}
	}

	if err != nil {
		atomic.StoreInt32(&this.status, int32(STOPPED))
		return err
	}

	log.Infof("Connect host: %s success.", connectHost)
	atomic.StoreInt32(&this.status, int32(RUNNING))
	this.conn = conn

	for _, fn := range this.afterConnected {
		fn(this)
	}
	//vals := Jobbers.jobbers.Values()
	//for _, val := range vals {
	//	jb := val.(*Jobber)
	//	status, _ := jb.GetStatus()
	//	if status == -1 {
	//		Jobbers.Start(jb.name)
	//	}
	//}
	//
	//logrus.Info("Connect rabbitMQ server success.")
	return nil
}

func (this *amqpConnection) close() error {
	return this.conn.Close()
}
