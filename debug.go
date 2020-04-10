package main

import (
	"fmt"
	"github.com/nagae-memooff/config"

	"net/http"
	_ "net/http/pprof"
)

func init() {
	init_queue = append(init_queue, InitProcess{
		Order:    10,
		InitFunc: listenDebug,
	})
}

func listenDebug() {
	if config.Get("debug_port") != "" {
		debug_listen := fmt.Sprintf("0.0.0.0:%s", config.Get("debug_port"))
		fmt.Printf("listening debug port: %s\n", debug_listen)

		go http.ListenAndServe(debug_listen, nil)
	}
}
