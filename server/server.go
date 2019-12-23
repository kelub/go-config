package server

import (
	"context"
	"fmt"
	"github.com/Sirupsen/logrus"
	"kelub/go-config/consul"
	serverpb "kelub/go-config/pb/server"
)

/*
RPC服务接口实现
消息队列接口实现
*/

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
