package main

import (
	"flag"
	"github.com/garyburd/redigo/redis"
	"github.com/sghaida/redis-migration-tool/pool"
	"github.com/sirupsen/logrus"
)

const numJobs = 10

var (
	srcHost      = flag.String("src", "host:port", "source redis host")
	dstHost      = flag.String("dst", "host:port", "destination redis host")
	keyToMigrate = flag.String("key", "*dafiti*", "redis key pattern to migrate")
	ttl          = flag.Int("ttl", 60*60, "redis key TTL")
)

func main() {
	// parse input
	flag.Parse()
	// worker channel
	jobs := make(chan string, numJobs)

	workerPool := pool.NewWorkerPool(*srcHost, *dstHost, jobs)
	for w := 1; w <= 10; w++ {
		go workerPool.Importer(w, jobs, *ttl)
	}

	keys, err := redis.Strings(workerPool.GetSourceConn().Do("KEYS", *keyToMigrate))
	if err != nil {
		logrus.Error(err)
	}

	var counter int
	for _, key := range keys {
		jobs <- key
		counter++
		logrus.Info(counter)
	}
}
