package api

import (
	//   "encoding/json"
	//   "errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/nagae-memooff/config"
	"go_sca/biz"
	"go_sca/entities"
	"go_sca/models"
	"go_sca/public"
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

func routers() {
	Get("/health_check", health_check)
	Get("/incr/:table", incr)
	Get("/user/:id", GetUser)
}

func ListenHttp() {
	listen := fmt.Sprintf("%s:%s", config.GetMulti("http_listen", "http_port")...)
	config.Default("http_base_url", public.Proname)

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

func BindParams(c *gin.Context, params interface{}) (err error) {
	c.ShouldBindUri(params)
	err = c.ShouldBind(params)

	return
}

func return_ok(c *gin.Context, msg string) {
	//   w.Write([]byte(fmt.Sprintf(`{"status": "ok", "msg": "%s"}`, msg)))
	c.String(http.StatusOK, fmt.Sprintf(`{"status": "ok", "msg": "%s"}`, msg))
}

func return_err(c *gin.Context, code int, msg string) {
	c.String(code, fmt.Sprintf(`{"status": "err", "msg": "%s"}`, msg))
}

func SuccessCode(c *gin.Context, msg string) {
	//   w.Write([]byte(fmt.Sprintf(`{"status": "ok", "msg": "%s"}`, msg)))
	c.JSON(200, gin.H{
		"status": "ok",
		"msg":    msg,
	})
}

/// func return_err(c *gin.Context, code int, msg string) {
/// 	c.String(code, fmt.Sprintf(`{"status": "err", "msg": "%s"}`, msg))
/// }

func ErrorCode(c *gin.Context, code int, msg interface{}, err_code ...interface{}) {
	re, t := c.Get("responseBody")
	if !t {
		re = ""
	}
	hash := gin.H{
		"status": "error",
		"error":  msg,
		"detail": re,
	}

	if len(err_code) > 0 {
		hash["err_code"] = err_code[0]
	}

	c.JSON(code, hash)
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

// / ...
// @Summary 获取一个user
// @Tags users
// @Router /users/id/{id} [get]
// @Param id path int true "id"
// ...
// @Security auth
// // Success 200 {string} json "{"id": 1, "login_name": "xx", name: "xx"}"
// @Success 200 {object} entities.User
func GetUser(c *gin.Context) {
	var params struct {
		ID int `binding:"required" json:"id" form:"id" uri:"id"`
	}

	if err := BindParams(c, &params); err != nil {
		ErrorCode(c, 400, err)

		return
	}

	var user models.User
	user = biz.FindUser(params)

	if user.ID == 0 {
		// ErrorCode(c, 404, "user not found", 404)
		ErrorCode(c, 404, "user not found")
	} else {
		c.JSON(200, entities.NewUser(user))
	}
}
