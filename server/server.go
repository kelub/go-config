package server

import (
	"bytes"
	"context"
	"fmt"
	"github.com/Sirupsen/logrus"
	"github.com/golang/protobuf/proto"
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
var fastWatchTopic = "fastWatch_"

type GetConf struct {
}

// GetConfig PRC 实现
func (c *GetConf) GetConfig(ctx context.Context, req *serverpb.GetConfReq) (rsp *serverpb.GetConfRsp, err error) {
	if !req.IsFastWatch {
		return c.getConfig(ctx, req)
	} else {
		return c.fastWatch(ctx, req)
	}
}

func (c *GetConf) getConfig(ctx context.Context, req *serverpb.GetConfReq) (rsp *serverpb.GetConfRsp, err error) {
	logEntry := logrus.WithFields(logrus.Fields{
		"func_name": "getCOnfig",
		"service":   req.GetService(),
		"key":       req.GetKey(),
		"subkey":    req.GetSubkey(),
		"keyType":   serverpb.KeyType_name[req.GetKeyType()],
	})
	//keyType := serverpb.KeyType_name[req.GetKeyType()]
	logEntry.Debugln("GetConfig")
	KeyPrefix, Key, err := c.keyPrefix(req)
	if err != nil {
		return nil, err
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
		Service: req.GetService(),
		List:    confValues,
	}
	return
}

func (c *GetConf) fastWatch(ctx context.Context, req *serverpb.GetConfReq) (rsp *serverpb.GetConfRsp, err error) {
	logEntry := logrus.WithFields(logrus.Fields{
		"func_name": "fastWatch",
		"service":   req.GetService(),
		"key":       req.GetKey(),
		"subkey":    req.GetSubkey(),
		"keyType":   serverpb.KeyType_name[req.GetKeyType()],
	})
	KeyPrefix, Key, err := c.keyPrefix(req)
	if err != nil {
		return nil, err
	}
	fw := consul.GetFastWatch()
	done := make(chan struct{})
	if req.KeyType == 0 {
		valueCh, err := fw.KeyWatch(done, KeyPrefix)
		if err != nil {
			logEntry.Errorf("KeyWatch error:", err)
			return nil, err
		}
		go func() {
			pub := loader.GetGExporter().Publisher
			topic := fastWatchTopic + KeyPrefix
			for {
				select {
				case <-done:
					logEntry.Errorf("KeyWatch Loop Exit")
					return
				case list := <-valueCh:
					confValues := make([]*serverpb.ConfValue, 1)
					confValue := new(serverpb.ConfValue)
					confValue.Key = req.GetKey()
					confValue.Subkey = req.GetSubkey()
					confValue.Value = string(list)
					confValues[0] = confValue
					confs := serverpb.GetConfRsp{
						Service: req.GetService(),
						List:    confValues,
					}
					v, err := proto.Marshal(&confs)
					if err != nil {
						logEntry.Errorf("消息序列化失败")
						continue
					}
					err = pub.Publish(topic, v)
					if err != nil {
						return
					}
				}
			}
		}()
	} else if req.KeyType == 1 {
		valueCh, err := fw.KeyPrefixWatch(done, KeyPrefix)
		if err != nil {
			logEntry.Errorf("KeyPrefixWatch error:", err)
			return nil, err
		}
		go func() {
			pub := loader.GetGExporter().Publisher
			topic := fastWatchTopic + KeyPrefix
			for {
				select {
				case <-done:
					logEntry.Errorf("KeyPrefixWatch Loop Exit")
					return
				case list := <-valueCh:
					confValues := make([]*serverpb.ConfValue, 0, len(list))
					for k, v := range list {
						subkey := []byte(k)[len(Key):]
						confValue := new(serverpb.ConfValue)
						confValue.Key = Key
						confValue.Subkey = string(subkey)
						confValue.Value = string(v)

						confValues = append(confValues, confValue)
					}
					confs := serverpb.GetConfRsp{
						Service: req.GetService(),
						List:    confValues,
					}
					v, err := proto.Marshal(&confs)
					if err != nil {
						logEntry.Errorf("消息序列化失败")
						continue
					}
					err = pub.Publish(topic, v)
					if err != nil {
						return
					}
				}
			}
		}()
	} else {
		return nil, fmt.Errorf("KeyType Parameter error")
	}
	return nil, nil
}

func (c *GetConf) keyPrefix(req *serverpb.GetConfReq) (KeyPrefix, Key string, err error) {
	if len(req.GetService()) == 0 {
		logrus.Errorf("service is nil")
		return "", "", fmt.Errorf("service is nil")
	}
	KeyPrefix = req.GetService()
	Key = req.GetService()
	if len(req.GetKey()) != 0 {
		KeyPrefix = KeyPrefix + "/" + req.GetKey()
		Key = Key + "/" + req.GetKey()
		if len(req.GetSubkey()) != 0 {
			KeyPrefix = KeyPrefix + "/" + req.GetSubkey()
		}
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
			if v.Err != nil {
				logrus.Errorf("WatchLoop Error", v.Err)
				return
			}
			message, err := p.parese(v.Value)
			if err != nil {
				logrus.Errorf("proto Marshal error", err)
				return
			}
			err = ex.Publisher.Publish(topicName, message)
			if err != nil {
				return
			}
		}
	}
}

func (p *PushManager) parese(watchValue map[string][]byte) (data []byte, err error) {
	var value []byte
	for k, _ := range watchValue {
		value = []byte(k)
		break
	}
	values := bytes.Split(value, []byte("/"))

	service := bytes.Join(values[:2], []byte("/"))
	key := bytes.Join(values[:3], []byte("/"))

	//[]byte(service)
	confValues := make([]*serverpb.ConfValue, 0, len(watchValue))

	for k, v := range watchValue {
		subkey := []byte(k)[len(key):]
		confValue := new(serverpb.ConfValue)
		confValue.Key = string(key)
		confValue.Subkey = string(subkey)
		confValue.Value = string(v)
		confValues = append(confValues, confValue)
	}
	message := &serverpb.GetConfRsp{
		Service: string(service),
		List:    confValues,
	}

	data, err = proto.Marshal(message)
	if err != nil {
		logrus.Errorf("proto Marshal error", err)
		return nil, fmt.Errorf("proto Marshal error %s", err.Error())
	}
	return data, nil
}
