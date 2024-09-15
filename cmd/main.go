package main

import (
	"log"

	"github.com/sounishnath003/jobprocessor/internal/core"
	"github.com/sounishnath003/jobprocessor/internal/utils"
)

func main() {

	// Defining the ConfigOpts
	conf := core.ConfigOpts{
		PORT: utils.GetEnv("PORT", 3000).(int),
	}

	// Initializing the Core
	co, err := core.InitCore(conf)
	if err != nil {
		log.Println("error occured init Core:", err)
		panic(err)
	}

	srv := NewServer(co)

	srv.MustStart()
}
