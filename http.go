package main

import (
	//   "encoding/json"
	//   "errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/nagae-memooff/config"
	"net/http"
	//   "os"
	//   "strings"
	//   "sync"
	//   "time"
)

var (
	router       *gin.Engine
	router_group *gin.RouterGroup
)

func init() {
	init_queue = append(init_queue, InitProcess{
		Order:    5,
		InitFunc: listenHttp,
	})
}

func routers() {
	Get("/health_check", health_check)
}

func listenHttp() {
	listen := fmt.Sprintf("%s:%s", config.GetMulti("http_listen", "http_port")...)
	config.Default("http_base_url", Proname)

	base_url := config.Get("http_base_url")

	router = gin.Default()
	router_group = router.Group(base_url)

	routers()
	if config.Get("log_file") != "stdout" {
		gin.SetMode(gin.ReleaseMode)
	}

	go router.Run(listen)
	// err = router.Run(listen)

	// if err != nil {
	// 	Log.Critical("failed to listen http port: %s", err)

	// }
	// return
}

func Get(path string, f func(c *gin.Context)) {
	router_group.GET(path, f)
}

func Post(path string, f func(c *gin.Context)) {
	router_group.POST(path, f)
}

func Put(path string, f func(c *gin.Context)) {
	router_group.PUT(path, f)
}

func Delete(path string, f func(c *gin.Context)) {
	router_group.DELETE(path, f)
}

func Patch(path string, f func(c *gin.Context)) {
	router_group.PATCH(path, f)
}

func Head(path string, f func(c *gin.Context)) {
	router_group.HEAD(path, f)
}

func Options(path string, f func(c *gin.Context)) {
	router_group.OPTIONS(path, f)
}

func return_ok(c *gin.Context, msg string) {
	//   w.Write([]byte(fmt.Sprintf(`{"status": "ok", "msg": "%s"}`, msg)))
	c.String(http.StatusOK, fmt.Sprintf(`{"status": "ok", "msg": "%s"}`, msg))
}

func return_err(c *gin.Context, code int, msg string) {
	c.String(code, fmt.Sprintf(`{"status": "err", "msg": "%s"}`, msg))
}

func health_check(c *gin.Context) {
	// Parameters in path
	// name := c.Param("name")
	// action := c.Param("action")
	// message := name + " is " + action
	// c.String(http.StatusOK, message)

	// Querystring parameters
	// firstname := c.DefaultQuery("firstname", "Guest")
	// lastname := c.Query("lastname") // shortcut for c.Request.URL.Query().Get("lastname")

	// POST multipart
	// message := c.PostForm("message")
	// nick := c.DefaultPostForm("nick", "anonymous")
	return_ok(c, "running")
}
