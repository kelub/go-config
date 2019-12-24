package server

import (
	"context"
	"fmt"
	"github.com/Sirupsen/logrus"
	"google.golang.org/grpc"
	"kelub/go-config/consul"
	"kelub/go-config/loader"
	serverpb "kelub/go-config/pb/server"
)

/*
RPC服务接口实现
消息队列接口实现
*/

var confPrefix = "conf/"
var topicName = "conf"

type GetConf struct {
}

// GetConfig PRC 实现
func (c *GetConf) GetConfig(ctx context.Context, req *serverpb.GetConfReq) (rsp *serverpb.GetConfRsp, err error) {
	logEntry := logrus.WithFields(logrus.Fields{
		"func_name": "GetConfig",
		"service":   req.GetService(),
		"key":       req.GetKey(),
		"subkey":    req.GetSubkey(),
		"keyType":   serverpb.KeyType_name[req.GetKeyType()],
	})
	//keyType := serverpb.KeyType_name[req.GetKeyType()]
	logEntry.Debugln("GetConfig")
	if len(req.GetService()) == 0 {
		logrus.Errorf("service is nil")
		return nil, fmt.Errorf("service is nil")
	}
	KeyPrefix := req.GetService()
	Key := req.GetService()
	if len(req.GetKey()) != 0 {
		KeyPrefix = KeyPrefix + "/" + req.GetKey()
		Key = Key + "/" + req.GetKey()
		if len(req.GetSubkey()) != 0 {
			KeyPrefix = KeyPrefix + "/" + req.GetSubkey()
		}
	}
	consulMgr := consul.GetConsulMgr()
	list, _, err := consulMgr.List(KeyPrefix)
	if err != nil {
		logEntry.Errorln("获取配置失败", err)
		return nil, err
	}

	confValues := make([]*serverpb.ConfValue, 0, len(list))
	for k, v := range list {
		subkey := []byte(k)[len(Key):]
		confValue := new(serverpb.ConfValue)
		confValue.Key = Key
		confValue.Subkey = string(subkey)
		confValue.Value = string(v)

		confValues = append(confValues, confValue)
	}
	rsp = &serverpb.GetConfRsp{
		List: confValues,
	}
	return
}

func RegisterGetConfig(s *grpc.Server) {
	serverpb.RegisterConfigServer(s, &GetConf{})
}

type PushManager struct {
}

func (p *PushManager) Run() {
	c := consul.GetConsulMgr()
	opts := loader.GetOptions()
	ex := loader.GetGExporter()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	watchValue := c.WatchLoop(ctx, confPrefix, opts.Watch.PollInterval)
	for {
		select {
		case <-ctx.Done():
			return
		case v := <-watchValue:

			ex.Publisher.Publish(topicName)
		}
	}
}
