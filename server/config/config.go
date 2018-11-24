package config

import (
	"os"
	"strings"

	"context"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/xinzf/consumer_proxy/utils"
)

var (
	project map[string]string
)

type Config struct {
	Name string
}

func Init(ctx context.Context) error {

	project = ctx.Value("project_info").(map[string]string)

	c := Config{}

	c.initLog()

	cfg := os.Getenv(project["name"] + "_CONFIG")

	if utils.PathExists(cfg) == false {
		log.Panic("缺少配置文件或配置文件不存在")
	}

	c.Name = cfg

	// 初始化配置文件
	if err := c.initConfig(); err != nil {
		return err
	}

	return nil
}

func (c *Config) initConfig() error {
	if c.Name != "" {
		viper.SetConfigFile(c.Name) // 如果指定了配置文件，则解析指定的配置文件
	}
	viper.SetConfigType("yaml")         // 设置配置文件格式为YAML
	viper.AutomaticEnv()                // 读取匹配的环境变量
	viper.SetEnvPrefix(project["name"]) // 读取环境变量的前缀
	replacer := strings.NewReplacer(".", "_")
	viper.SetEnvKeyReplacer(replacer)
	if err := viper.ReadInConfig(); err != nil { // viper解析配置文件
		return err
	}

	return nil
}

func (c *Config) initLog() {
	log.SetFormatter(&log.TextFormatter{})
	log.SetOutput(os.Stdout)

	runMode := viper.GetString("runmode")
	if runMode != "release" {
		log.SetLevel(log.DebugLevel)
	} else {
		log.SetLevel(log.ErrorLevel)
	}
}
