package mq

import (
	"context"
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
	"sync"
	"sync/atomic"
	"time"
)

type ConsumerOptions struct {
	// @todo name = name + node
	Name  string
	Queue struct {
		Name    string
		Durable bool
	}
	Exchange struct {
		Name    string
		Etype   string
		Durable bool
	}
	BindKey   string
	Consumer  string
	WorkerNum int
	TargetUrl string
	RetryNum  int
	Log       struct {
		Path    string
		Maxsize int
	}
}

type Consumer struct {
	name          string
	channel       *amqp.Channel
	options       ConsumerOptions
	ctx           context.Context
	cancle        context.CancelFunc
	once          sync.Once
	status        int32
	closeNotifies []chan bool
	startTime     time.Time
	stopTime      time.Time
	workerBucket  chan int  // 令牌桶
	workerPool    sync.Pool // worker 临时对象池
	logger        *Logger
}

func NewConsumer(options ConsumerOptions) (*Consumer, error) {
	var err error
	if options.Name == "" {
		err = errors.New("缺少消费者名称")
		return nil, err
	}

	if options.Queue.Name == "" {
		err = errors.New("缺少消费队列")
		return nil, err
	}

	if options.Exchange.Name == "" {
		err = errors.New("缺少绑定 Exchange")
		return nil, err
	}

	if options.Exchange.Etype == "" {
		err = errors.New("缺少 Exchange 类型")
		return nil, err
	}

	if options.Exchange.Etype != "direct" && options.Exchange.Etype != "fanout" {
		err = fmt.Errorf("Exchange 类型：%s 错误", options.Exchange.Etype)
		return nil, err
	}

	if options.Log.Path == "" {
		err = errors.New("缺少日志目录")
		return nil, err
	}

	if options.RetryNum == 0 {
		options.RetryNum = 1
	}

	return &Consumer{
		name:          options.Name,
		options:       options,
		stopTime:      time.Now(),
		closeNotifies: make([]chan bool, 0),
		status:        0,
	}, nil
}

func (this *Consumer) Start() error {
	defer func() {
		if err := recover(); err != nil {
			switch err.(type) {
			case error:
				logrus.Errorln(err)
			}
		} else {
			logrus.Infof("Jobber: %s exits", this.name)
		}
	}()

	var (
		msg <-chan amqp.Delivery
		err error
	)

	if this.Status() == RUNNING {
		return fmt.Errorf("消费者：%s 正在运行，不能重复开启", this.name)
	}

	if msg, err = this.preparStart(); err != nil {
		return err
	}

	logrus.Infoln("Jobber started successful.")

	notify := make(chan *amqp.Error)
	var runErr error
BREAK:
	for {
		select {
		case <-this.ctx.Done():
			break BREAK
		case runErr = <-this.channel.NotifyClose(notify):
			break BREAK
		case delivery, ok := <-msg:
			if !ok {
				runErr = errors.New("delivery channel has closed")
				logrus.Errorln(runErr)
				break BREAK
			}

			i, ok := <-this.workerBucket
			if !ok {
				runErr = errors.New("workers channel has closed")
				logrus.Errorln(runErr)
				break BREAK
			}

			if this.Status() != RUNNING {
				break BREAK
			}

			go this.do(delivery, i)
		}
	}

	// 等待所有工作线程退出
	// 如果能成功取车 workerNum 的令牌，说明所有的 worker 都已经执行完毕了
	for i := 0; i < this.options.WorkerNum; i++ {
		<-this.workerBucket
	}
	close(this.workerBucket)

	// 根据运行中的错误情况判定，程序是正常退出还是异常退出
	if runErr != nil {
		atomic.StoreInt32(&this.status, int32(ABNORMAL))
	} else {
		atomic.StoreInt32(&this.status, int32(STOPPED))
	}
	this.stopTime = time.Now()

	// 通知所有需要得知当前 Jobber 退出情况的监听者
	if len(this.closeNotifies) > 0 {
		for _, c := range this.closeNotifies {
			close(c)
		}
	}

	if runErr != nil {
		logrus.Errorln("Jobber exits with error: ", runErr.Error())
	} else {
		logrus.Infoln("Jobber exits.")
	}
	this.channel.Close()

	return nil
}

func (this *Consumer) Stop() error {
	return nil
}

func (this *Consumer) Status() State {
	status := atomic.LoadInt32(&this.status)
	return State(status)
}

func (this *Consumer) Name() string {
	return this.name
}

func (this *Consumer) Options() ConsumerOptions {
	return this.options
}

func (this *Consumer) StartTime() time.Time {
	return this.startTime
}

func (this *Consumer) StopTime() time.Time {
	return this.stopTime
}

func (this *Consumer) preparStart() (msg <-chan amqp.Delivery, err error) {
	atomic.StoreInt32(&this.status, int32(STARTING))
	ctx, cancle := context.WithCancel(context.Background())
	this.ctx = ctx
	this.cancle = cancle

	// 获取一个 channel
	this.channel, err = Connection.getChannel()
	if err != nil {
		return
	}

	// 创建队列
	_, err = this.channel.QueueDeclare(
		this.options.Queue.Name,
		this.options.Queue.Durable,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return
	}

	// 创建路由
	err = this.channel.ExchangeDeclare(
		this.options.Exchange.Name,
		this.options.Exchange.Etype,
		this.options.Exchange.Durable,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return
	}

	// 绑定队列到路由
	err = this.channel.QueueBind(
		this.options.Queue.Name,
		this.options.BindKey,
		this.options.Exchange.Name,
		false,
		nil,
	)
	if err != nil {
		return
	}

	// 设置 QOS
	err = this.channel.Qos(this.options.WorkerNum, 0, false)
	if err != nil {
		return
	}

	// 订阅队列
	msg, err = this.channel.Consume(
		this.options.Queue.Name,
		this.options.Consumer,
		false,
		false,
		false,
		false,
		nil,
	)

	// 初始化工作线程池，线程池容量等于 mq.prefetchCount，先塞满 workerNum 的令牌
	this.workerBucket = make(chan int, this.options.WorkerNum)
	for i := 0; i < this.options.WorkerNum; i++ {
		this.workerBucket <- i
	}

	// 初始化 jobber stop 的阻塞通知池
	this.closeNotifies = make([]chan bool, 0)

	// 初始化 worker 临时对象池
	this.workerPool = sync.Pool{New: func() interface{} {
		return NewWorker(this.options.TargetUrl)
	}}

	// 设置状态和开始时间
	atomic.StoreInt32(&this.status, int32(RUNNING))
	this.startTime = time.Now()

	return
}

func (this *Consumer) do(msg amqp.Delivery, i int) {
	worker := this.workerPool.Get().(*Worker)
	defer func() {
		if err := recover(); err != nil {
			switch err.(type) {
			case error:
				logrus.Errorln(err)
				//this.logger.With("workerId", i).Errorln("Jobber do request has some error: ", err.(error).Error())
			}
		}

		// ack
		msg.Ack(false)

		// worker 对象放回对象池
		this.workerPool.Put(worker)

		// 放回令牌
		this.workerBucket <- i
	}()

	worker.SetBody(msg)
	if err := worker.Do(); err != nil {

		logrus.Errorln(err)
		// @todo 从 redis 中获取 messageid 的失败次数，
		// @todo 如果失败次数已经大于等于 retryNum，则写入数据库
		// @todo 否则，通过死信的方式做延迟队列，并 incre 失败次数

		this.channel.Publish(
			"",
			this.options.Queue.Name,
			false,
			false,
			amqp.Publishing{
				Headers:         msg.Headers,
				ContentType:     msg.ContentType,
				ContentEncoding: msg.ContentEncoding,
				DeliveryMode:    msg.DeliveryMode,
				//Priority:msg.Priority
				CorrelationId: msg.CorrelationId,
				ReplyTo:       msg.ReplyTo,
				Expiration:    msg.Expiration,
				MessageId:     msg.MessageId,
				Timestamp:     time.Now(),
				Type:          msg.Type,
				UserId:        msg.UserId,
				AppId:         msg.AppId,
				Body:          msg.Body,
			},
		)
	}

}
