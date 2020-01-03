package pubsub

import (
	"fmt"
	"github.com/Shopify/sarama"
	"github.com/sirupsen/logrus"
	"time"
)

type KKAsyncPublisher struct {
	producer       sarama.AsyncProducer
	addr           []string
	pubWaitTimeout time.Duration
}

func CreateKKAsyncPublisher(addr []string, pubWaitTimeout time.Duration) *KKAsyncPublisher {
	pub := &KKAsyncPublisher{
		addr:           addr,
		pubWaitTimeout: pubWaitTimeout,
	}
	err := pub.Create()
	if err != nil {
		pub.producer.AsyncClose()
		return nil
	}
	return pub
}

func (pub *KKAsyncPublisher) config() *sarama.Config {
	config := &sarama.Config{}
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Partitioner = sarama.NewRandomPartitioner
	config.Producer.Return.Successes = true
	config.Producer.Return.Errors = true
	return config

}

func (pub *KKAsyncPublisher) producerLoop() error {
	for {
		select {
		case <-pub.producer.Successes():
		case err := <-pub.producer.Errors():
			logrus.Errorln("Failed to produce message", err)
		}
	}
}

func (pub *KKAsyncPublisher) Create() error {
	config := pub.config()
	producer, err := sarama.NewAsyncProducer(pub.addr, config)
	if err != nil {
		logrus.Errorln("Failed to create producer")
		return err
	}
	pub.producer = producer
	go pub.producerLoop()
	return nil
}

func (pub *KKAsyncPublisher) Close() {
	if pub.producer != nil {
		pub.producer.AsyncClose()
	}
}

func (pub *KKAsyncPublisher) Publish(topic string, data []byte) error {
	msg := &sarama.ProducerMessage{
		Topic: topic,
		Value: sarama.ByteEncoder(data),
	}
	select {
	case <-time.After(pub.pubWaitTimeout):
		return fmt.Errorf("publish wait timeout")
	case pub.producer.Input() <- msg:
		logrus.Debugf("publish Successes: topic=%s, data=%s \n", topic, string(data))

	}
	return nil
}
