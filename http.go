package main

import (
	//   "encoding/json"
	//   "errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/nagae-memooff/config"
	"net/http"

	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	//   "os"
	//   "strings"
	//   "sync"
	//   "time"
)

var (
	router       *gin.Engine
	router_group *gin.RouterGroup

	abc int
)

func init() {
	init_queue = append(init_queue, InitProcess{
		Order:    5,
		InitFunc: listenHttp,
	})
}

func routers() {
	Get("/health_check", health_check)
	Get("/incr/:table", incr)
}

func listenHttp() {
	listen := fmt.Sprintf("%s:%s", config.GetMulti("http_listen", "http_port")...)
	config.Default("http_base_url", Proname)

	base_url := config.Get("http_base_url")

	if config.Get("log_file") != "stdout" {
		gin.SetMode(gin.ReleaseMode)
	}

	router = gin.Default()
	router_group = router.Group(base_url)

	routers()

	srv := &http.Server{
		Addr:    listen,
		Handler: h2c.NewHandler(router, &http2.Server{}),
	}
	go srv.ListenAndServe()
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

func incr(c *gin.Context) {
	// Parameters in path
	//   _ = c.Param("table")
	// action := c.Param("action")
	// message := name + " is " + action
	// c.String(http.StatusOK, message)

	// Querystring parameters
	// firstname := c.DefaultQuery("firstname", "Guest")
	// lastname := c.Query("lastname") // shortcut for c.Request.URL.Query().Get("lastname")

	// POST multipart
	// message := c.PostForm("message")
	// nick := c.DefaultPostForm("nick", "anonymous")
	abc += 1
	c.Writer.Write([]byte(fmt.Sprintf("%d\n", abc)))
}
