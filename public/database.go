package public

import (
	"fmt"
	"github.com/nagae-memooff/config"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var (
	Db *gorm.DB
)

func init() {
	InitQueue = append(InitQueue, InitProcess{
		Order:    5,
		InitFunc: initMysql,
		QuitFunc: closeMysql,
	})
}

func initMysql() {
	config.Default("mysql_host", "localhost")
	config.Default("mysql_port", "3306")
	config.Default("mysql_charset", "utf8mb4")

	user := config.Get("mysql_user")
	password := config.Get("mysql_pwd")
	host := config.Get("mysql_host")
	port := config.Get("mysql_port")
	dbname := config.Get("mysql_dbname")

	charset := config.Get("mysql_charset")

	var err error
	addr := fmt.Sprintf("%s:%s@(%s:%s)/%s?charset=%s&parseTime=True&loc=Local", user, password, host, port, dbname, charset)
	Db, err = gorm.Open(mysql.Open(addr), &gorm.Config{})

	if err != nil {
		Shutdown(5, "connect mysql server failed: %s", err)
	}

}

func closeMysql() {
	// gorm v2版本支持连接池，官方明确表明不需要手动关闭数据库连接
}
