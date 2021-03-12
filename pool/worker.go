package pool

import (
	"github.com/garyburd/redigo/redis"
	"github.com/sirupsen/logrus"
	"time"
)

type WorkerPool struct {
	srcPool     *redis.Pool
	dstPool     *redis.Pool
	commandChan <-chan string
}

func NewWorkerPool(src, dst string, commands <-chan string) *WorkerPool {

	dial := func(host string) func() (redis.Conn, error) {
		return func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", host)
			if err != nil {
				return nil, err
			}
			return c, err
		}
	}

	barrow := func(c redis.Conn, t time.Time) error {
		_, err := c.Do("PING")
		return err
	}
	// source redis pool
	srcPool := &redis.Pool{
		Dial:         dial(src),
		TestOnBorrow: barrow,
		MaxIdle:      10,
		MaxActive:    60,
		IdleTimeout:  240 * time.Second,
	}
	// destination redis pool
	dstPool := &redis.Pool{
		Dial:         dial(dst),
		TestOnBorrow: barrow,
		MaxIdle:      10,
		MaxActive:    60,
		IdleTimeout:  240 * time.Second,
	}

	return &WorkerPool{
		srcPool:     srcPool,
		dstPool:     dstPool,
		commandChan: commands,
	}
}

func (w *WorkerPool) Importer(id int, keys <-chan string, ttl int) {
	src := w.srcPool.Get()
	dst := w.dstPool.Get()

	for key := range keys {
		keysToMigrate, _ := redis.Strings(src.Do("KEYS", key))
		if len(keysToMigrate) == 0 {
			logrus.Infof("worker: %d started", id)
			values, err := redis.Strings(src.Do("LRANGE", key, 0, 100))
			if err != nil {
				logrus.Errorf("unable to get the values for key %v: %v", key, err)
			}
			// push right
			for _, value := range values {
				_, err := dst.Do("RPUSH", key, value)
				if err != nil {
					logrus.Error(err)
				}
			}
			// set TTL
			_, err = dst.Do("EXPIRE", key, ttl)
			if err != nil {
				logrus.Error(err)
			}
			logrus.Infof("worker: %d done", id)
		}
	}
}

func (w *WorkerPool) GetSourceConn() redis.Conn {
	return w.srcPool.Get()
}
