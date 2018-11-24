package registry

import (
	"context"
	"crypto/tls"
	"fmt"
	consul "github.com/hashicorp/consul/api"
	log "github.com/sirupsen/logrus"
	"sync"
	"time"
)

type Consul struct {
	Address   string
	Client    *consul.Client
	Timeout   time.Duration
	Intval    time.Duration
	Secure    bool
	TLSConfig *tls.Config

	// Other options for implementations of the interface
	// can be stored in a context
	Context context.Context

	// connect enabled
	connect bool

	queryOptions *consul.QueryOptions

	sync.Mutex
	//register map[string]uint64
	// lastChecked tracks when a node was last checked as existing in Consul
	lastChecked time.Time
}

func NewRegistry() *Consul {
	config := consul.DefaultConfig()
	client, err := consul.NewClient(config)
	if err != nil {
		log.Panic(err)
	}

	return &Consul{
		Timeout: 3 * time.Second,
		Intval:  3 * time.Second,
		Client:  client,
	}
}

func (this *Consul) Register(node *Node) error {

	// @todo tags 是针对 Node 还是针对 service ?
	tags := this.encodeMetadata(node.Metadata)
	asr := &consul.AgentServiceRegistration{
		ID:      node.Id,
		Name:    node.Name,
		Tags:    tags,
		Port:    node.Port,
		Address: node.Address,
		Check: &consul.AgentServiceCheck{
			HTTP:                           fmt.Sprintf("http://%s:%d%s", node.Address, node.Port, "/check"),
			Timeout:                        this.Timeout.String(),
			Interval:                       this.Intval.String(),
			DeregisterCriticalServiceAfter: "30s",
		},
	}

	if this.connect {
		asr.Connect = &consul.AgentServiceConnect{
			Native: true,
		}
	}

	if err := this.Client.Agent().ServiceRegister(asr); err != nil {
		return err
	}

	this.Lock()
	this.lastChecked = time.Now()
	this.Unlock()

	return nil
}

func (this *Consul) Degister(node *Node) error {
	this.Lock()
	err := this.Client.Agent().ServiceDeregister(node.Id)
	this.Unlock()
	return err
}

func (this *Consul) encodeMetadata(md map[string]string) []string {
	var tags []string
	for k, v := range md {
		tags = append(tags, fmt.Sprintf("%s:%s", k, v))
	}

	return tags
}
