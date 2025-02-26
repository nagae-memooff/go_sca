package public

import (
	"os"
	"os/signal"
	"syscall"

	"time"

	"fmt"
	"runtime/debug"
	"sort"
	// "github.com/nagae-memooff/config"
)

func WaitSignal() {
	c := make(chan os.Signal)
	signal.Notify(c, syscall.SIGUSR1, syscall.SIGUSR2, syscall.SIGINT, syscall.SIGTERM)
	t := 0

	//阻塞直至有信号传入
	for s := range c {
		switch s.String() {
		case "terminated", "interrupt":
			if t > 0 {
				// 连续两下，强制退出
				os.Exit(0)
			}
			t++

			go Shutdown(0, "")
			time.Sleep(time.Second)
		case "user defined signal 1":
			// ReloadGlobalConfig()
			// runtime.GC()

		case "user defined signal 2":
			// ReloadPluginsConfig()
			debug.FreeOSMemory()
		default:
			Log.Warningc(func() string {
				return fmt.Sprintf("receive system signal: %s", s.String())
			})
		}
	}
}

func PrintStartMsg() {
	Log.Info("start %s", Proname)
}

func Shutdown(code int, message string, params ...interface{}) {
	sort.Reverse(InitQueue)

	for _, init_process := range InitQueue {
		if init_process.QuitFunc != nil {
			init_process.QuitFunc()
		}
	}

	if message != "" {
		Log.Error(message, params...)
	}

	Log.Info("shut down %s", Proname)

	closeLogger()

	os.Exit(code)
}
