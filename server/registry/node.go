package registry

import (
	"context"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/xinzf/consumer_proxy/utils"
	"net"
	"os"
)

type Node struct {
	Id       string            `json:"id"`
	Name     string            `json:"name"`
	Address  string            `json:"address"`
	Port     int               `json:"port"`
	Metadata map[string]string `json:"metadata"`
}

func InitNode(ctx context.Context) *Node {

	project := ctx.Value("project_info").(map[string]string)

	host, _ := os.Hostname()
	uuid := utils.UUID()

	node := &Node{
		Id:   uuid,
		Name: project["name"],
		Metadata: map[string]string{
			"host": host,
		},
	}

	node.Address = viper.GetString("server.addr")
	port := viper.GetInt("server.port")

	if node.Address == "" {
		ips, err := utils.GetIP()
		if err != nil {
			log.Panic(err)
		}

		node.Address = ips[0]
	}

	if port == 0 {
		l, _ := net.Listen("tcp", ":0")
		port = l.Addr().(*net.TCPAddr).Port
		l.Close()
	}

	node.Port = port

	return node
}
