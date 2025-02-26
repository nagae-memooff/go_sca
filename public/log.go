package public

import (
	"fmt"
	"github.com/nagae-memooff/config"
	// "gitlab.10101111.com/oped/goconfig.git"
	log "github.com/nagae-memooff/log4go"
	"time"
)

var Log log.Logger

func InitLogger() {
	LogLevel := log.LevelByString(config.Get("log_level"))

	Log = log.NewDefaultLogger(LogLevel)

	switch config.Get("log_file") {
	case "", "stdout":
		fmt.Printf("print log to stdout.\n")
	default:
		var filter *log.FileLogWriter

		if config.GetBool("enable_log_rotate") {
			filter = log.NewFileLogWriter(config.Get("log_file"), true)
			filter.SetRotateSize(config.GetInt("log_rotate_size") << 20)
			// filter.SetRotateSize(4096)
			filter.SetRotateKeep(config.GetInt("log_rotate_keep"))
		} else {
			filter = log.NewFileLogWriter(config.Get("log_file"), false)
		}
		Log.AddFilter("file", LogLevel, filter)

		fmt.Printf("print log to %s.\n", config.Get("log_file"))
	}

}

func closeLogger() {
	//   Log.Close()
	log.Close()
	time.Sleep(50 * time.Millisecond)
}
