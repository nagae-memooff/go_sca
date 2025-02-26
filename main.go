package main

import (
	//   utils "github.com/nagae-memooff/goutils"
	"go_sca/api"
	"go_sca/models"
	"go_sca/public"
	//   "time"
)

func init() {
}

func main() {
	initConfig()
	public.InitLogger()

	public.InitQueueFunc1()

	go public.WaitSignal()

	api.ListenHttp()

	public.PrintStartMsg()
	_main()

	<-make(chan bool)
}

func _main() {
	// TODO 主逻辑

	var u []models.User
	params := map[string]interface{}{
		"name": "admin",
	}

	public.Db.Where(params).Find(&u)

	public.Log.Info(u)

	// for {
	// 	time.Sleep(time.Second)
	// public.KafkaProducer.Write([]byte("string"))
	// }
}
