package go_config

import (
	"github.com/Sirupsen/logrus"
	"google.golang.org/grpc"
	"kelub/go-config/consul"
	"kelub/go-config/loader"
	"kelub/go-config/server"
	"kelub/go-config/util"
)

func Main() {
	wg := &util.WaitGroupWrapper{}
	opts := loader.NewOptions()
	ex := loader.CreateExporter(opts)
	err := consul.InitConsul(opts.ConsulAddress)
	if err != nil {
		logrus.Errorf("InitConsul error", err)
		return
	}
	RegisterRPC(ex.RPCServer)

	wg.Wrap(func() {
		ex.HTTPServer.Main(opts)
	})
	wg.Wrap(func() {
		loader.RunRPCServer(ex.RPCServer, opts.RPCAddress, opts.RPCPort)
	})
	wg.Wait()
}

func RegisterRPC(s *grpc.Server) {
	logEntry := logrus.WithFields(logrus.Fields{
		"func_name": "RegisterRPC",
	})
	server.RegisterGetConfig(s)
	logEntry.Infoln("注册RPC服务")
}
