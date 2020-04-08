package main

import (
	"fmt"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"github.com/nagae-memooff/config"
	"time"
)

var (
	db *gorm.DB
)

func init() {
	init_queue = append(init_queue, InitProcess{
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
	db, err = gorm.Open("mysql", addr)

	if err != nil {
		shutdown(5, "connect mysql server failed: %s", err)
	}

}

func closeMysql() {
	db.Close()
}

type Model struct {
	ID        uint `gorm:"primary_key;AUTO_INCREMENT"`
	CreatedAt time.Time
	UpdatedAt time.Time
	//   DeletedAt *time.Time
}

type User struct {
	Model
	// Email string `gorm:"type:varchar(100);unique_index"`
	// Role         string  `gorm:"size:255"`        // set field size to 255
	// MemberNumber *string `gorm:"unique;not null"` // 指针类型用以规避默认0值陷阱
	// Num          int     `gorm:"AUTO_INCREMENT"`  // set num to auto incrementable
	// Address      string  `gorm:"index:addr"`      // create index with name `addr` for address
	Name      string
	RoleCode  int
	Suspended bool
	Actived   bool
	IgnoreMe  int `gorm:"-"` // ignore this field
}

func (user *User) BeforeCreate(scope *gorm.Scope) error {
	// scope.SetColumn("ID", uuid.New())
	return nil
}
