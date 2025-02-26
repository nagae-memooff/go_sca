package public

import (
	// "fmt"
	"sort"
)

var (
	Proname string
)

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
	InitQueue InitProcessQueue
)

func InitQueueFunc1() {
	sort.Sort(InitQueue)

	// 初始化其他机制
	for _, init_process := range InitQueue {
		if init_process.InitFunc != nil {
			init_process.InitFunc()
		}
	}

	for _, init_process := range InitQueue {
		if init_process.StartFunc != nil {
			init_process.StartFunc()
		}
	}
}
