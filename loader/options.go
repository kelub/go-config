package loader

import (
	"github.com/spf13/viper"
	"time"
)

var goptions *Options

type Options struct {
	ServerName    string   `flag:"server_name"`
	RPCAddress    string   `flag:"rpc-address"`
	RPCPort       int      `flag:"rpc-port"`
	ConsulAddress string   `flag:"tcp-port"`
	HealthPort    int      `flag:"HealthPort"`
	ProfPort      int      `flag:"prof_port"`
	KafkaAddress  []string `flag:"kafka-address"`
	Watch         struct {
		PollInterval time.Duration
	}
}

func NewOptions() *Options {
	interval := viper.GetInt("poll_interval")
	if interval == 0 {
		interval = 30
	}
	PollInterval := time.Duration(interval)

	goptions := &Options{
		ServerName:    viper.GetString("server_name"),
		RPCAddress:    viper.GetString("rpc_addr"),
		RPCPort:       viper.GetInt("rpc_port"),
		ConsulAddress: viper.GetString("consul_port"),
		HealthPort:    viper.GetInt("health_port"),
		ProfPort:      viper.GetInt("prof_port"),
		KafkaAddress:  viper.GetStringSlice("kafka_addrs"),
		Watch: struct {
			PollInterval time.Duration
		}{
			PollInterval: PollInterval * time.Second,
		},
	}
	return goptions
}

func GetOptions() *Options {
	return goptions
}

func GetTCPAddress() string {
	return fmt.Sprintf("%s:%d", viper.GetString("tcp_addr"), viper.GetInt("tcp_port"))
}
