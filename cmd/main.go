package main

import (
	"context"
	"log"
	"time"

	"github.com/sounishnath003/jobprocessor/internal/core"
	"github.com/sounishnath003/jobprocessor/internal/utils"
)

func main() {

	// Defining the ConfigOpts
	conf := core.ConfigOpts{
		PORT: utils.GetEnv("PORT", 3000).(int),
		BrokerQueue: core.RedisBroker{
			Addrs:        []string{"127.0.0.1:6379"},
			Password:     "",
			DB:           1,
			MinIdleConns: 20,
			DialTimeout:  1 * time.Second,
			ReadTimeout:  1 * time.Second,
			WriteTimeout: 1 * time.Second,
		},
		ResultQueue: core.RedisResult{
			RedisBroker: core.RedisBroker{
				Addrs:        []string{"127.0.0.1:6379"},
				Password:     "",
				DB:           1,
				MinIdleConns: 20,
				DialTimeout:  1 * time.Second,
				ReadTimeout:  1 * time.Second,
				WriteTimeout: 1 * time.Second,
			},
			Expiry:     600 * time.Second, // 10 mins
			MetaExpiry: 600 * time.Second, // 10 mins
		},
	}

	// Initializing the Core
	co, err := core.InitCore(conf)
	if err != nil {
		log.Println("error occured init Core:", err)
		panic(err)
	}

	// Setup an ctx background.
	ctx := context.Background()

	srv := NewServer(co)
	go srv.MustStart()

	if err := co.Start(ctx, "worker-name", 5); err != nil {
		panic(err)
	}
}
