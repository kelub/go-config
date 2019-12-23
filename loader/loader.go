package loader

import (
	"context"
	"fmt"
	"github.com/Sirupsen/logrus"
	"github.com/go-redis/redis"
	"github.com/go-xorm/xorm"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"kelub/go-config/server"
	"kelub/go-config/util"
	"net"
	"net/http"
	"strconv"
)

var gExporter *Exporter

type Options struct {
	ServerName    string `flag:"server_name"`
	RPCAddress    string `flag:"rpc-address"`
	RPCPort       int    `flag:"rpc-port"`
	ConsulAddress string `flag:"tcp-port"`
	HealthPort    int    `flag:"HealthPort"`
	ProfPort      int    `flag:"prof_port"`
}

func NewOptions() *Options {
	return &Options{
		ServerName:    viper.GetString("server_name"),
		RPCAddress:    viper.GetString("rpc_addr"),
		RPCPort:       viper.GetInt("rpc_port"),
		ConsulAddress: viper.GetString("consul_port"),
		HealthPort:    viper.GetInt("health_port"),
		ProfPort:      viper.GetInt("prof_port"),
	}
}

func GetTCPAddress() string {
	return fmt.Sprintf("%s:%d", viper.GetString("tcp_addr"), viper.GetInt("tcp_port"))
}

type Exporter struct {
	RPCServer *grpc.Server

	MysqlEngine *xorm.Engine
	RedisEngine *redis.Client

	HTTPServer *HTTPServer
}

func CreateExporter(opts *Options) *Exporter {
	exporter := &Exporter{}
	SetGExporter(exporter)

	exporter.RPCServer = createRPCServer()
	exporter.HTTPServer = createHTTPServer()
	return exporter
}

func SetGExporter(exporter *Exporter) {
	gExporter = exporter
}

func RegisterRPC(s *grpc.Server) {
	server.RegisterGetConfig(s)
}

func createRPCServer() *grpc.Server {
	rpcOption := make([]grpc.ServerOption, 0)
	rpcOption = append(rpcOption, grpc.UnaryInterceptor(GrpcInterceptor))
	s := grpc.NewServer(rpcOption...)
	return s
}

func GrpcInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
	resp, err = handler(ctx, req)
	return resp, err
}

func RunRPCServer(s *grpc.Server, addr string, port int) {
	logEntry := logrus.WithFields(logrus.Fields{
		"func_name": "runRPCServer",
		"addr":      addr,
		"port":      port,
	})

	logEntry.Infoln("启动 RPC 服务")
	if addr == "" {
		logEntry.Info("RPC 地址或者端口未配置 ")
		return
	}
	lis, err := net.Listen("tcp", fmt.Sprintf("%s:%d", addr, port))
	if err != nil {
		logEntry.Info(err)
		return
	}
	if err := s.Serve(lis); err != nil {
		logEntry.Infoln("启动 RPC 服务失败")
	}
}

type HTTPServer struct {
	opts *Options
}

func RunHttpServer(wg *util.WaitGroupWrapper, ex *Exporter, opts *Options) {
	wg.Wrap(func() {
		ex.HTTPServer.Main(opts)
	})
}

func createHTTPServer() *HTTPServer {
	return &HTTPServer{}
}

func (h *HTTPServer) Main(opts *Options) error {
	h.opts = opts
	h.httpfunc()
	httpPort := strconv.Itoa(h.opts.HealthPort)
	httpAddr := fmt.Sprintf(":%s", httpPort)
	logrus.Infof("start health http listen: %s ", httpPort)
	err := http.ListenAndServe(httpAddr, nil)
	if err != nil {
		logrus.Panicf("health http listen error: port=%d", httpPort)
		return err
	}
	return nil
}

func (h *HTTPServer) httpfunc() {
	http.HandleFunc("status", h.statusHandler)
}

func (h *HTTPServer) statusHandler(w http.ResponseWriter, _ *http.Request) {
	fmt.Fprint(w, "status ok!")
}
