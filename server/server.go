package server

import (
	"net/http"

	"github.com/xinzf/consumer_proxy/server/registry"
	"github.com/xinzf/consumer_proxy/server/router"

	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"time"
)

var (
	node        *registry.Node
	Registry    *registry.Consul
)

func Init(ctx context.Context) {
	node = registry.InitNode(ctx)

	log.Infof("Node: %+v", node)

	Registry = registry.NewRegistry()
}

func Start() error {

	g := gin.New()

	runmode := viper.GetString("server.runmode")
	if runmode == "" {
		runmode = "debug"
	}
	gin.SetMode(runmode)

	router.Load(g)

	go func() {
		address := fmt.Sprintf("%s:%d", node.Address, node.Port)
		log.Infoln("http server listen on ", address, ":", node.Port)
		err := http.ListenAndServe(address, g).Error()
		if err != "" {
			log.Panic(err)
		}
	}()

	return nil
}

func Stop() error {
	return nil
}

func Register() error {
	return Registry.Register(node)
}

func Deregister() error {
	return Registry.Degister(node)
}
