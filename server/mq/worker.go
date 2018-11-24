package mq

import (
	"bytes"
	"fmt"
	"github.com/json-iterator/go"
	"github.com/streadway/amqp"
	"io/ioutil"
	"net"
	"net/http"
	"time"
)

type WorkerResponse struct {
	Msg_Code   int         `json:"msg_code"`
	Message    string      `json:"message"`
	Attachment interface{} `json:"attachment"`
}

var WorkerOptions struct {
	MaxIdleConns        int
	MaxIdleConnsPerHost int
	IdleConnTimeout     int
	Timeout             int64
	KeepAlive           int64
}

func init() {
	WorkerOptions.MaxIdleConns = 100
	WorkerOptions.MaxIdleConnsPerHost = 100
	WorkerOptions.IdleConnTimeout = 90
	WorkerOptions.Timeout = 30
	WorkerOptions.KeepAlive = 30
}

func NewWorker(targetURI string) *Worker {
	return &Worker{
		targetURI: targetURI,
		client: &http.Client{
			Transport: &http.Transport{
				Proxy: http.ProxyFromEnvironment,
				DialContext: (&net.Dialer{
					Timeout:   time.Duration(WorkerOptions.Timeout) * time.Second,
					KeepAlive: time.Duration(WorkerOptions.KeepAlive) * time.Second,
				}).DialContext,
				MaxIdleConns:        WorkerOptions.MaxIdleConns,
				MaxIdleConnsPerHost: WorkerOptions.MaxIdleConnsPerHost,
				IdleConnTimeout:     time.Duration(WorkerOptions.IdleConnTimeout) * time.Second,
			},
		},
	}
}

type Worker struct {
	targetURI string
	delivery  amqp.Delivery
	client    *http.Client
}

func (this *Worker) SetBody(delivery amqp.Delivery) {
	this.delivery = delivery
}

func (this *Worker) Do() error {
	rawData := bytes.NewBuffer(this.delivery.Body)
	req, err := http.NewRequest("POST", this.targetURI, rawData)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", this.delivery.ContentType)

	rsp, err := this.client.Do(req)
	if err != nil {
		return err
	}
	defer rsp.Body.Close()

	if rsp.StatusCode != 200 {
		return fmt.Errorf("Request failed, The http status is: %d.", rsp.StatusCode)
	}

	body, _ := ioutil.ReadAll(rsp.Body)

	var response WorkerResponse
	err = jsoniter.Unmarshal(body, &response)
	if err != nil {
		return err
	}

	if response.Msg_Code != 200 {
		return fmt.Errorf("Response failed: %s", string(body))
	}

	return nil
}
