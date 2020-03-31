package main

import (
	"fmt"
	"github.com/nagae-memooff/config"
	utils "github.com/nagae-memooff/goutils"
	json "github.com/pquerna/ffjson/ffjson"
	"gopkg.in/Shopify/sarama.v1"
	"io/ioutil"
	"os"
	"sort"
	"strings"
	"sync"
	"time"
)

var (
	KafkaConsumer *KafkaConsumerClient

	kafka_offset_if_notexist int64
)

func init() {
	init_queue = append(init_queue, InitProcess{
		Order:     1,
		InitFunc:  initKafkaConsumer,
		StartFunc: startKafkaConsumer,
		QuitFunc:  closeKafkaConsumer,
	})
}

func initKafkaConsumer() {
	KafkaConsumer = &KafkaConsumerClient{
		brokers:              strings.Split(config.Get("kafka_brokers"), ","),
		topic:                config.Get("kafka_topic"),
		offset_info_file:     fmt.Sprintf("./%s.offset_info", config.Get("kafka_topic")),
		offset_info_file_tmp: fmt.Sprintf("./%s.offset_info.tmp", config.Get("kafka_topic")),
		msg_channel:          make(chan *sarama.ConsumerMessage, 256),
	}

	config.Default("kafka_processor_parrael", "1")

	KafkaConsumer.processor_parrael = config.GetInt("kafka_processor_parrael")

	var err error
	KafkaConsumer.consumer, err = sarama.NewConsumer(KafkaConsumer.brokers, nil)
	Log.Debug("consumer: %v\n", KafkaConsumer.consumer)
	if err != nil {
		shutdown(3, "connect kafka error: %s", err)
	}

	partitionList, err := KafkaConsumer.consumer.Partitions(KafkaConsumer.topic)
	Log.Debug("partitionList: %v\n", partitionList)
	if err != nil {
		shutdown(4, "get kafka partition list error: %s", err)
	}

	if config.Get("kafka_offset_if_notexist") != "oldest" {
		kafka_offset_if_notexist = sarama.OffsetNewest
	} else {
		kafka_offset_if_notexist = sarama.OffsetOldest
	}

	KafkaConsumer.loadOffsetsFromFile()

	KafkaConsumer.partition_consumers = make([]sarama.PartitionConsumer, 0, len(partitionList))

	for partition := range partitionList {
		pc, err := KafkaConsumer.consumer.ConsumePartition(KafkaConsumer.topic, int32(partition), KafkaConsumer.getOffsetForPartition(KafkaConsumer.topic, partition))
		if err != nil {
			shutdown(5, "consume kafka partition %d error: %s", partition, err)
		}

		KafkaConsumer.pending_msgs[partition] = utils.NewSet()
		KafkaConsumer.partition_consumers = append(KafkaConsumer.partition_consumers, pc)
	}
}

func startKafkaConsumer() {
	go KafkaConsumer.bgSaveOffset()

	for _, pc := range KafkaConsumer.partition_consumers {
		KafkaConsumer.pull_wg.Add(1)
		go KafkaConsumer.consumeKafkaPartition(pc)
	}

	for i := 0; i < KafkaConsumer.processor_parrael; i++ {
		KafkaConsumer.processor_wg.Add(1)
		go KafkaConsumer.processKafkaMessages(i)
	}
}

func closeKafkaConsumer() {
	for _, pc := range KafkaConsumer.partition_consumers {
		pc.Close()
	}

	// 在这里等一下，因为close的时点有可能还有飞着的消息，等消费者全部退出再关闭consumer对象
	KafkaConsumer.pull_wg.Wait()

	// Close shuts down the consumer. It must be called after all child
	// PartitionConsumers have already been closed.
	KafkaConsumer.consumer.Close()

	close(KafkaConsumer.msg_channel)

	KafkaConsumer.processor_wg.Wait()

	// 还要再主动持久化一次
	KafkaConsumer.saveOffset()
}

type KafkaConsumerClient struct {
	brokers           []string
	topic             string
	processor_parrael int

	pending_msgs       map[int]*utils.Set
	offsets            map[int]int64
	offsets_mutex      sync.RWMutex
	offset_if_notexist int64

	offset_info_file     string
	offset_info_file_tmp string

	// 拉取者用
	pull_wg sync.WaitGroup

	msg_channel chan *sarama.ConsumerMessage

	consumer            sarama.Consumer
	partition_consumers []sarama.PartitionConsumer

	// 消息处理者用
	processor_wg sync.WaitGroup
}

func (c *KafkaConsumerClient) processKafkaMessages(thd int) {
	defer c.processor_wg.Done()

	for msg := range c.msg_channel {
		// TODO
		Log.Info("receive msg: %s.", msg.Value)
	}
}

// 记录offset
func (c *KafkaConsumerClient) getOffsetForPartition(topic string, partition int) (offset int64) {
	// pending_msgs := pending_set.GetAllInt()

	c.offsets_mutex.RLock()
	wl_offset, ok1 := c.offsets[partition]
	c.offsets_mutex.RUnlock()

	if !ok1 {
		Log.Info("start consume %s-%d from newest.", topic, partition)

		c.offsets_mutex.Lock()
		c.offsets[partition] = kafka_offset_if_notexist
		c.offsets_mutex.Unlock()

		return kafka_offset_if_notexist
	}

	if wl_offset >= 0 {
		wl_offset += 1
	}

	// pdmin_msgs 里取最小的
	var pending_msgs []int
	_pending_msgs, ok2 := c.pending_msgs[partition]

	if ok2 {
		pending_msgs = _pending_msgs.GetAllInt()
	}

	// FIXME 如果有一个很久之前的pending msg，那会造成凭空大量消费
	if len(pending_msgs) > 0 {
		// sort排序后，最小值在前
		sort.Ints(pending_msgs)
		pdmin_offset := int64(pending_msgs[0])

		if wl_offset < pdmin_offset {
			offset = wl_offset
		} else {
			offset = pdmin_offset
		}
	} else {
		offset = wl_offset
	}

	// offset = sarama.OffsetNewest
	Log.Info("start consume %s-%d from %d.", topic, partition, offset)

	return
}

func (c *KafkaConsumerClient) consumeKafkaPartition(pc sarama.PartitionConsumer) {
	defer c.pull_wg.Done()

	go func(pc sarama.PartitionConsumer) {
		for errmsg := range pc.Errors() {
			Log.Error("consume failed: %v", errmsg)
		}
	}(pc)

	for msg := range pc.Messages() {
		c.pending_msgs[int(msg.Partition)].Add(msg.Offset)

		c.offsets_mutex.Lock()
		c.offsets[int(msg.Partition)] = msg.Offset
		c.offsets_mutex.Unlock()

		c.msg_channel <- msg
	}
}

func (c *KafkaConsumerClient) bgSaveOffset() {
	for {
		time.Sleep(time.Second)

		c.saveOffset()
	}
}

func (c *KafkaConsumerClient) saveOffset() {
	_offsets := make([]*OffsetInfo, 0, len(c.pending_msgs))

	for partition, pending_set := range c.pending_msgs {
		pending_msgs := pending_set.GetAllInt()

		c.offsets_mutex.RLock()
		offset := c.offsets[partition]
		c.offsets_mutex.RUnlock()

		Log.Debug("topic: %s, partition: %d, offset: %d, pending_set: %v.", c.topic, partition, offset, pending_msgs)
		offset_info := NewOffsetInfo(c.topic, partition, offset, pending_msgs)

		_offsets = append(_offsets, offset_info)
	}

	json_bytes, err := json.Marshal(_offsets)
	if err != nil {
		Log.Error("json marshal failed: %s", err)
	}

	c.saveOffsetToFile(json_bytes)
}

type OffsetInfo struct {
	Topic      string `json:"topic"`
	Partition  int    `json:"partition"`
	Offset     int64  `json:"offset"`
	PendingSet []int  `json:"pending_set"`
}

func NewOffsetInfo(topic string, partition int, offset int64, pending_set []int) *OffsetInfo {
	return &OffsetInfo{
		Topic:      topic,
		Partition:  partition,
		Offset:     offset,
		PendingSet: pending_set,
	}
}

// 落盘+mv
func (c *KafkaConsumerClient) saveOffsetToFile(info []byte) (err error) {
	err = ioutil.WriteFile(c.offset_info_file_tmp, info, 0644)
	if err != nil {
		return
	}

	err = os.Rename(c.offset_info_file_tmp, c.offset_info_file)
	return err
}

// 读取一下上次持久化的信息
func (c *KafkaConsumerClient) loadOffsetsFromFile() {
	c.pending_msgs = make(map[int]*utils.Set)
	c.offsets = make(map[int]int64)

	// 从文件中读取记录。若不存在则说明是第一次启动; 若存在但数据不对，则需要怎么打警告
	if !utils.IsFileExist(c.offset_info_file) {
		Log.Info("offset file not exist.will consume from newest.")
		return
	}

	_json_bytes, err := ioutil.ReadFile(c.offset_info_file)
	if err != nil {
		Log.Error("read offset file failed: %s.", err)
		return
	}

	Log.Info("load offset from file %s.:\n%s.", c.offset_info_file, _json_bytes)

	var _offsets []*OffsetInfo
	err = json.Unmarshal(_json_bytes, &_offsets)
	if err != nil {
		Log.Error("parse offset file failed: %s.", err)
		return
	}

	for _, offset_info := range _offsets {
		partition := offset_info.Partition
		offset := offset_info.Offset
		pending_set := offset_info.PendingSet

		c.offsets_mutex.Lock()
		c.offsets[partition] = offset
		c.offsets_mutex.Unlock()

		s := utils.NewSet()
		for _, pending := range pending_set {
			s.Add(pending)
		}

		c.pending_msgs[partition] = s

	}

	return
}
