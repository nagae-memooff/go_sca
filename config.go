package main

import (
	"fmt"
	"github.com/nagae-memooff/config"
	"log"
	"runtime"
)

func init() {
}

func initConfig() {
	err := config.Parse(fmt.Sprintf("./%s.conf", Proname))
	if err != nil {
		log.Fatal(err)
	}

	config.Default("listen", "127.0.0.1:8080")

	config.Default("base_url", fmt.Sprintf("/%s", Proname))

	config.Default("log_level", "info")
	config.Default("log_file", fmt.Sprintf("./%s.log", Proname))
	config.Default("enable_log_rotate", "true")
	config.Default("log_rotate_size", "100")
	config.Default("log_rotate_keep", "5")

	config.Default("GOMAXPROCS", "0")

	gomaxprocs := config.GetInt("GOMAXPROCS")
	if gomaxprocs <= 0 {
		runtime.GOMAXPROCS(runtime.NumCPU()/8 + 1)
	} else {
		runtime.GOMAXPROCS(gomaxprocs)
	}
}
