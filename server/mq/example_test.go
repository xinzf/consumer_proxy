package mq

import (
	"context"
	log "github.com/sirupsen/logrus"
	"time"
)

//func ExampleAmqpConnection_Start() {
//	Init(context.Background())
//
//	Connection.AfterConnected(func(connection *amqpConnection) {
//		log.Println("I am the connected:First")
//	}, func(connection *amqpConnection) {
//		log.Println("I am the connected:Second")
//	})
//
//	Connection.AfterClosed(func(connection *amqpConnection) {
//		log.Println("I am the closed:First")
//	}, func(connection *amqpConnection) {
//		log.Println("I am the closed:Second")
//	})
//	go Connection.Start(context.Background())
//
//	time.Sleep(time.Hour)
//	// Output: nil
//}

func ExampleConsumer_Start() {
	Init(context.Background())

	Connection.AfterConnected(func(connection *amqpConnection) {
		log.Println("I am the connected:First")
	}, func(connection *amqpConnection) {
		log.Println("I am the connected:Second")
	})

	Connection.AfterClosed(func(connection *amqpConnection) {
		log.Println("I am the closed:First")
	}, func(connection *amqpConnection) {
		log.Println("I am the closed:Second")
	})
	go Connection.Start(context.Background())

	time.Sleep(5 * time.Second)

	consumer, err := NewConsumer(ConsumerOptions{
		Name: "haha",
		Queue: struct {
			Name    string
			Durable bool
		}{Name: "proxy_test", Durable: true},
		Exchange: struct {
			Name    string
			Etype   string
			Durable bool
		}{Name: "proxy_ex", Etype: string("direct"), Durable: true},
		BindKey:   "proxy_test",
		Consumer:  "hdhafd",
		WorkerNum: 4,
		TargetUrl: "http://127.0.0.1:8082/index.php",
		RetryNum:  2,
		Log: struct {
			Path    string
			Maxsize int
		}{Path: string("/Users/xiangzhi/Work/Go/src/github.com/xinzf/consumer_proxy/log.log"), Maxsize: 300},
	})

	if err != nil {
		panic(err)
	}

	err = consumer.Start()
	if err != nil {
		panic(err)
	}
	time.Sleep(time.Hour)
	// Output: nil
}
