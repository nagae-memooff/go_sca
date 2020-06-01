package main

import (
	//   utils "github.com/nagae-memooff/goutils"
	"sort"
	//   "time"
)

func init() {
}

type InitProcess struct {
	// 初始化的顺序，从小到大执行
	Order int

	// 初始化方法，一定要写成*阻塞*的（但不要阻塞死，按顺序初始化）
	InitFunc func()

	// 开始执行方法，要写成*非阻塞*的，按顺序执行
	StartFunc func()

	// 关闭方法，关闭时按照倒序依次关闭，一定要写成*阻塞*的
	QuitFunc func()
}

type InitProcessQueue []InitProcess

func (q InitProcessQueue) Len() int {
	return len(q)
}

func (q InitProcessQueue) Less(i, j int) bool {
	return q[i].Order < q[j].Order
}

func (q InitProcessQueue) Swap(i, j int) {
	q[j], q[i] = q[i], q[j]
}

var (
	init_queue InitProcessQueue
)

func main() {
	initConfig()
	initLogger()

	sort.Sort(init_queue)

	// 初始化其他机制
	for _, init_process := range init_queue {
		if init_process.InitFunc != nil {
			init_process.InitFunc()
		}
	}

	printStartMsg()

	for _, init_process := range init_queue {
		if init_process.StartFunc != nil {
			init_process.StartFunc()
		}
	}

	go waitSignal()

	_main()

	<-make(chan bool)
}

func _main() {
	// TODO 主逻辑

	var u []User
	params := map[string]interface{}{
		"name": "admin",
	}

	db.Where(params).Find(&u)

	Log.Info(u)

	// for {
	// 	time.Sleep(time.Second)
	KafkaProducer.Write([]byte("string"))
	// }
}
