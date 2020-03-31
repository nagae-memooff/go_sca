package main

import (
	// "encoding/json"
	// "bytes"
	"errors"
	"fmt"
	"sync"

	"github.com/nagae-memooff/config"
	utils "github.com/nagae-memooff/goutils"
	"gopkg.in/Shopify/sarama.v1"
	"strings"
	"sync/atomic"
	"time"
)

var (
	// PendingBuffer      = make(map[string]*bytes.Buffer)
	// PendingBufferMutex sync.RWMutex

	KafkaProducer *KafkaProducerClient
)

func init() {
	init_queue = append(init_queue, InitProcess{
		Order:     1,
		InitFunc:  initKafkaProducer,
		StartFunc: startKafkaProducer,
		QuitFunc:  closeKafkaProducer,
	})
}

func initKafkaProducer() {
	// TODO 解析配置
	_config := utils.NewHashTable()
	_config.Set("brokers", config.Get("kafka_producer_brokers"))
	_config.Set("topic", config.Get("kafka_producer_topic"))
	_config.Set("is_sync", config.Get("kafka_producer_is_sync"))
	_config.Set("use_gzip", config.Get("kafka_producer_use_gzip"))

	KafkaProducer = NewKafkaProducerClient(_config)
}

func startKafkaProducer() {
	_ = KafkaProducer.Start()
}

func closeKafkaProducer() {
	KafkaProducer.Quit()
}

type KafkaProducerClient struct {
	brokers     []string
	config      *sarama.Config
	client      sarama.AsyncProducer
	sync_client sarama.SyncProducer
	topic       string
	running     bool

	total        uint64
	succeed      uint64
	errors_count uint64

	is_sync      bool
	write_method func(body []byte) (err error)

	qmux sync.WaitGroup
}

func (c *KafkaProducerClient) Start() (err error) {
	if c.running {
		return
	}

	if c.is_sync {
		producer, err := sarama.NewSyncProducer(c.brokers, c.config)
		if err != nil {
			c.running = false
			Log.Error("init kafka producer failed: %s", err)
			return err
		}

		c.sync_client = producer
		c.running = true

	} else {

		producer, err := sarama.NewAsyncProducer(c.brokers, c.config)
		if err != nil {
			c.running = false
			Log.Error("init kafka producer failed: %s", err)
			return err
		}

		c.client = producer
		c.running = true

		// 准备success 和 error的处理频道
		c.qmux.Add(2)
		go func() {
			for msg := range producer.Successes() {
				// metadata, ok := msg.Metadata.([]uint8)
				// if ok {
				// 	tail_id := string(metadata)
				// 	PendingBufferMutex.RLock()
				// 	buffer, ok2 := PendingBuffer[tail_id]
				// 	PendingBufferMutex.RUnlock()

				// 	if ok2 {
				// 		buffer.Reset()
				// 		// BytesBufferPool.Put(buffer)

				// 		PendingBufferMutex.Lock()
				// 		delete(PendingBuffer, tail_id)
				// 		PendingBufferMutex.Unlock()

				// 		Log.Debug("tail_id recycle: %s", tail_id)
				// 	}
				// }

				c.succeed = atomic.AddUint64(&c.succeed, 1)

				Log.Debugc(func() string {
					byt, err := msg.Value.Encode()
					return fmt.Sprintf("Send to kafka succeed: %s, %v", byt, err)
				})
			}

			Log.Info("kafka producer closed.")
			c.running = false
			c.qmux.Done()
			return
		}()

		go func() {
			for msg := range producer.Errors() {
				c.errors_count = atomic.AddUint64(&c.errors_count, 1)
				Log.Error("Send to kafka error: %s", msg)
			}

			Log.Info("kafka producer closed.")
			c.running = false
			c.qmux.Done()
			return
		}()
	}

	return
}

func (c *KafkaProducerClient) Write(body []byte) (err error) {
	if !c.running {
		err = errors.New("not running")
		return
	}

	err = c.write_method(body)
	return
}

func (c *KafkaProducerClient) syncWrite(body []byte) (err error) {
	split := len(body) - 35
	real_body := body[:split]

	msg := &sarama.ProducerMessage{Topic: c.topic, Value: sarama.ByteEncoder(real_body)}

	if c.sync_client == nil {
		err = errors.New("client is nil")
		return
	}

	_, _, err = c.sync_client.SendMessage(msg)
	if err != nil {
		time.Sleep(2 * time.Second)
		_, _, err = c.sync_client.SendMessage(msg)

		// TODO 重试一次，如果还失败就扔进error队列里
		if err != nil {
			Log.Error("send to kafka failed: %s", body)
			c.errors_count = atomic.AddUint64(&c.errors_count, 1)
		}
	} else {
		Log.Debug("send to kafka succeed: %s", body)
	}

	if err == nil {
		c.succeed = atomic.AddUint64(&c.succeed, 1)
	}

	c.total = atomic.AddUint64(&c.total, 1)
	return
}

func (c *KafkaProducerClient) asyncWrite(body []byte) (err error) {
	// split := len(body) - 35
	// real_body := body[:split]
	// tail_id := body[split:]
	// msg := &sarama.ProducerMessage{Topic: c.topic, Value: sarama.ByteEncoder(real_body), Metadata: tail_id}

	msg := &sarama.ProducerMessage{Topic: c.topic, Value: sarama.ByteEncoder(body)}

	if c.client == nil {
		err = errors.New("client is nil")
		return
	}

	c.client.Input() <- msg

	c.total = atomic.AddUint64(&c.total, 1)
	return
}

func (c *KafkaProducerClient) Info() (info *utils.HashTable) {
	info = utils.NewHashTable()
	info.Set("total", c.total)
	info.Set("succeed", c.succeed)
	info.Set("errors", c.errors_count)

	return
}

func (c *KafkaProducerClient) Quit() (err error) {
	if !c.running {
		return
	}

	c.running = false
	time.Sleep(time.Second)

	if c.is_sync {
		c.sync_client.Close()
	} else {
		c.client.Close()
	}

	c.qmux.Wait()
	Log.Info("send %d, %d succees, %d failed, %d pending or unknown.", c.total, c.succeed, c.errors_count, c.total-c.succeed-c.errors_count)
	return
}

func NewKafkaProducerClient(config *utils.HashTable) (c *KafkaProducerClient) {
	c = &KafkaProducerClient{}

	conf := sarama.NewConfig()
	conf.Metadata.Retry.Max = 1
	conf.Metadata.Retry.Backoff = 500 * time.Millisecond
	conf.Producer.RequiredAcks = sarama.RequiredAcks(sarama.WaitForLocal)
	conf.Producer.Timeout = 5 * time.Second
	conf.Producer.MaxMessageBytes = 5 << 20 // 5 MB
	conf.Producer.Flush.Bytes = 1 << 20     // 1 MB
	conf.Producer.Flush.Frequency = time.Second
	conf.Producer.Return.Successes = true
	conf.Producer.Return.Errors = true
	conf.Net.DialTimeout = 2 * time.Second

	if config.GetInt("use_gzip") == 1 {
		conf.Producer.Compression = sarama.CompressionGZIP
	} else {
		conf.Producer.Compression = sarama.CompressionNone
	}

	c.is_sync = config.GetBool("is_sync")
	if c.is_sync {
		c.write_method = c.syncWrite
	} else {
		c.write_method = c.asyncWrite
	}

	c.brokers = strings.Split(config.Get("brokers"), ",")
	c.config = conf
	c.topic = config.Get("topic")

	return
}
