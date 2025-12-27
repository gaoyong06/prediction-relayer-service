package data

import (
	"context"
	"encoding/json"

	"xinyuan_tech/relayer-service/internal/conf"

	"github.com/apache/rocketmq-client-go/v2"
	"github.com/apache/rocketmq-client-go/v2/primitive"
	"github.com/apache/rocketmq-client-go/v2/producer"
	"github.com/go-kratos/kratos/v2/log"
)

// RocketMQProducer RocketMQ 生产者接口
type RocketMQProducer interface {
	// SendMessage 发送消息
	SendMessage(ctx context.Context, topic, tag string, body interface{}) error
	// Close 关闭生产者
	Close() error
}

// rocketmqProducer RocketMQ 生产者实现
type rocketmqProducer struct {
	producer rocketmq.Producer
	topic    string
	log      *log.Helper
}

// NewRocketMQ 创建 RocketMQ 生产者实例
func NewRocketMQ(c *conf.Data, logger log.Logger) (RocketMQProducer, func(), error) {
	logger = log.With(logger, "module", "data/rocketmq")
	logHelper := log.NewHelper(logger)

	var nameServer string
	var producerGroup string
	var topic string
	var retryTimes int

	if c.GetRocketmq() != nil {
		rmqConf := c.GetRocketmq()
		nameServer = rmqConf.GetNameServer()
		producerGroup = rmqConf.GetProducerGroup()
		topic = rmqConf.GetTopic()
		retryTimes = int(rmqConf.GetRetryTimes())
	}

	if nameServer == "" {
		nameServer = "127.0.0.1:9876" // 默认值
	}
	if producerGroup == "" {
		producerGroup = "relayer_service_producer_group" // 默认值
	}
	if topic == "" {
		topic = "relayer_events" // 默认值
	}
	if retryTimes == 0 {
		retryTimes = 2 // 默认值
	}

	// 创建生产者
	p, err := rocketmq.NewProducer(
		producer.WithNsResolver(primitive.NewPassthroughResolver([]string{nameServer})),
		producer.WithGroupName(producerGroup),
		producer.WithRetry(retryTimes),
	)
	if err != nil {
		logHelper.Errorf("Failed to create RocketMQ producer: %v", err)
		return nil, nil, err
	}

	// 启动生产者
	if err := p.Start(); err != nil {
		logHelper.Errorf("Failed to start RocketMQ producer: %v", err)
		return nil, nil, err
	}

	logHelper.Infof("RocketMQ producer started: nameServer=%s, producerGroup=%s, topic=%s", nameServer, producerGroup, topic)

	cleanup := func() {
		if p != nil {
			p.Shutdown()
		}
	}

	return &rocketmqProducer{
		producer: p,
		topic:    topic,
		log:      logHelper,
	}, cleanup, nil
}

// SendMessage 发送消息
func (r *rocketmqProducer) SendMessage(ctx context.Context, topic, tag string, body interface{}) error {
	// 序列化消息体
	data, err := json.Marshal(body)
	if err != nil {
		r.log.Errorf("Failed to marshal message body: %v", err)
		return err
	}

	// 如果未指定 topic，使用默认 topic
	if topic == "" {
		topic = r.topic
	}

	// 创建消息
	msg := primitive.NewMessage(topic, data)
	if tag != "" {
		msg.WithTag(tag)
	}

	// 发送消息（同步发送）
	result, err := r.producer.SendSync(ctx, msg)
	if err != nil {
		r.log.Errorf("Failed to send message: %v", err)
		return err
	}

	r.log.Debugf("Message sent successfully: topic=%s, tag=%s, msgId=%s", topic, tag, result.MsgID)
	return nil
}

// Close 关闭生产者
func (r *rocketmqProducer) Close() error {
	if r.producer != nil {
		return r.producer.Shutdown()
	}
	return nil
}


