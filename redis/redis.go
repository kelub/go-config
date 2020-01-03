package redis

import (
	"github.com/sirupsen/logrus"
	"github.com/go-redis/redis"
)

type redisConfig struct {
	address  string
	password string
}

func NewRedisCli(conf *redisConfig) (*redis.Client, error) {
	opt := &redis.Options{
		Addr:       conf.address,
		Password:   conf.password,
		DB:         0,
		PoolSize:   20000,
		MaxRetries: 1,
	}
	cli := redis.NewClient(opt)
	stringCmd := cli.Ping()
	if stringCmd.Err() != nil {
		logrus.Errorf("get new RedisCli error err:%v", stringCmd.Err())
		return nil, stringCmd.Err()
	}
	return cli, nil
}
