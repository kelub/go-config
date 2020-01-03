package pubsub

import (
	"context"
	"fmt"
	"github.com/Shopify/sarama"
	"github.com/sirupsen/logrus"
	"time"
)

type KKSubscriber struct {
	addrs []string
}

func CreateKKSubscriber(addrs []string) *KKSubscriber {
	return &KKSubscriber{addrs: addrs}
}

type GroupHandler struct {
}

// Setup is run at the beginning of a new session, before ConsumeClaim.
func (h *GroupHandler) Setup(_ sarama.ConsumerGroupSession) error {
	return nil
}

func (h *GroupHandler) Cleanup(_ sarama.ConsumerGroupSession) error {
	return nil
}

// Messages returns the read channel for the messages that are returned by
// the broker. The messages channel will be closed when a new rebalance cycle
// is due. You must finish processing and mark offsets within
// Config.Consumer.Group.Session.Timeout before the topic/partition is eventually
// re-assigned to another group member.
func (h *GroupHandler) ConsumeClaim(sess sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		fmt.Printf("Message topic:%q partition:%d offset:%d\n", msg.Topic, msg.Partition, msg.Offset)
		sess.MarkMessage(msg, "")
	}
	return nil
}

func (sub *KKSubscriber) Subscribe(topic string, groupID string, handler sarama.ConsumerGroupHandler) error {
	logEntry := logrus.WithFields(logrus.Fields{
		"func_name": "Subscribe",
		"topic":     topic,
		"groupID":   groupID,
	})
	if handler == nil {
		return fmt.Errorf("handler is nil")
	}
	config := &sarama.Config{}
	config.Consumer.Return.Errors = true
	config.Consumer.Offsets.Initial = sarama.OffsetNewest
	config.Consumer.Group.Session.Timeout = 10 * time.Second
	ConsumerGroup, err := sarama.NewConsumerGroup(sub.addrs, groupID, config)
	if err != nil {
		logEntry.Errorf("Subscribe Failed", err)
	}
	defer ConsumerGroup.Close()
	// Track errors
	go func() {
		for err := range ConsumerGroup.Errors() {
			fmt.Println("ERROR", err)
			logEntry.Errorf("message error", err)
		}
	}()

	ctx := context.Background()
	for {
		err := ConsumerGroup.Consume(ctx, []string{topic}, handler)
		if err != nil {
			logEntry.Errorf("Group Consume error", err)
			return err
		}
	}
}
