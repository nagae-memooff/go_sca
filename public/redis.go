package public

import (
	"errors"
	"fmt"
	"github.com/FZambia/sentinel"
	"github.com/gomodule/redigo/redis"
	"github.com/nagae-memooff/config"
	"strings"
	"time"
)

var (
	RedisPool   *redis.Pool
	RedisPrefix string

	// 哨兵模式需要
	sntnl             *sentinel.Sentinel
	master_changed_at time.Time
)

func init() {
	InitQueue = append(InitQueue, InitProcess{
		Order:    1,
		InitFunc: initPool,
		QuitFunc: quitPool,
	})
}

func initPool() {
	config.Default("redis_host", "127.0.0.1:6379")
	config.Default("redis_max_idle", "20")
	config.Default("redis_max_active", "20")
	config.Default("redis_prefix", Proname)

	master_name := config.Get("sentinel_master_name")

	if master_name != "" {
		initSentinelPool()
	} else {
		initRedisPool()
	}

}

func initRedisPool() {
	redis_host := config.Get("redis_host")
	redis_passwd := config.Get("redis_passwd")
	max_idle := config.GetInt("redis_max_idle")
	max_active := config.GetInt("redis_max_active")

	var dial func() (c redis.Conn, err error)

	if redis_passwd != "" {
		dial = func() (c redis.Conn, err error) {
			pwd := redis.DialPassword(redis_passwd)
			c, err = redis.Dial("tcp", redis_host, pwd)

			return c, err
		}

	} else {
		dial = func() (c redis.Conn, err error) {
			c, err = redis.Dial("tcp", redis_host)
			return c, err
		}

	}

	RedisPool = &redis.Pool{
		MaxIdle:   max_idle,
		MaxActive: max_active, // max number of connections
		Wait:      true,

		Dial: dial,
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			_, err := c.Do("PING")
			return err
		},
	}
}

func initSentinelPool() {
	addrs := strings.Split(config.Get("sentinel_hosts"), ",")
	master_name := config.Get("sentinel_master_name")

	sntnl = &sentinel.Sentinel{
		Addrs:      addrs,
		MasterName: master_name,
		Pool: func(addr string) *redis.Pool {
			return &redis.Pool{
				MaxIdle:   2,
				MaxActive: 4,
				Wait:      true,
				Dial: func() (redis.Conn, error) {
					return redis.Dial("tcp", addr)
				},
				TestOnBorrow: func(c redis.Conn, t time.Time) error {
					_, err := c.Do("PING")
					return err
				},
			}
		},
		Dial: func(addr string) (redis.Conn, error) {
			timeout := 500 * time.Millisecond
			c, err := redis.DialTimeout("tcp", addr, timeout, timeout, timeout)
			if err != nil {
				return nil, err
			}
			return c, nil
		},
	}

	master_changed_at = time.Now()

	for _, addr := range sntnl.Addrs {
		go func(addr string) {
			pool := sntnl.Pool(addr)

			for {
				c := pool.Get()
				receiver := redis.PubSubConn{c}

				receiver.Subscribe("+switch-master")
				for {
					switch v := receiver.Receive().(type) {
					case redis.Message:
						args := strings.Split(string(v.Data), " ")
						if args[0] != master_name {
							continue
						}

						//logs.Logger.Error("master changed from '%s:%s' to '%s:%s'.", args[1], args[2], args[3], args[4])
						master_changed_at = time.Now()
					case redis.Subscription:
						// do nothing
					case error:
						// logs.Logger.Error("error when watching master: %v. will retry it later.", v)
						c.Close()
						time.Sleep(time.Second * 10)
						break
					}
				}
			}
		}(addr)
	}
	// 订阅master_name
	// 如果变化了，就更新时间

	RedisPool = &redis.Pool{
		MaxIdle:   config.GetInt("redis_max_idle"),
		MaxActive: config.GetInt("redis_max_active"), // max number of connections
		Wait:      true,

		Dial: func() (redis.Conn, error) {
			masterAddr, err := sntnl.MasterAddr()
			if err != nil {
				return nil, err
			}
			var c redis.Conn

			redis_passwd := config.Get("redis_passwd")
			if redis_passwd != "" {
				pwd := redis.DialPassword(redis_passwd)
				c, err = redis.Dial("tcp", masterAddr, pwd)
			} else {
				c, err = redis.Dial("tcp", masterAddr)
			}

			if err != nil {
				return nil, err
			}
			return c, nil
		},

		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			if t.Before(master_changed_at) || !sentinel.TestRole(c, "master") {
				return errors.New("Role check failed")
			} else {
				return nil
			}
		},
	}

}

func quitPool() {
	master_name := config.Get("sentinel_master_name")
	if master_name != "" {
		err := sntnl.Close()

		if err != nil {
			Log.Warn("close redis sentinel failed: '%s'.", err)
		}

		for _, addr := range sntnl.Addrs {
			pool := sntnl.Pool(addr)
			err = pool.Close()

			if err != nil {
				Log.Warn("close redis failed: '%s'.", err)
			}
		}
	} else {
		err := RedisPool.Close()
		if err != nil {
			Log.Warn("close redis conn failed: '%s'.", err)
		}
	}
}

func RedisSet(key, value string, expire int) (data string, err error) {
	c := RedisPool.Get()
	defer func() {
		c.Close()
	}()

	key = fmt.Sprintf("%s/%s", RedisPrefix, key)
	data_origin, err := c.Do("set", key, value)
	if err != nil {
		return
	}

	c.Do("expire", key, expire)

	data = fmt.Sprintf("%s", data_origin)

	return data, nil
}

func RedisGet(key string) (data string, err error) {
	c := RedisPool.Get()
	defer func() {
		c.Close()
	}()

	key = fmt.Sprintf("%s/%s", RedisPrefix, key)
	data, err = redis.String(c.Do("get", key))
	if err != nil {
		return
	}

	return
}
