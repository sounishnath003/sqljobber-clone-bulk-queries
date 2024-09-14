package main

import (
	"log"

	"github.com/sounishnath003/jobprocessor/internal/utils"
)

func main() {

	// Defining the Env Config
	conf := ConfigOpts{
		PORT: utils.GetEnv("PORT", 3000).(int),
	}
	log.Printf("config:%+v\n", conf)

	srv := NewServer(&conf)

	srv.MustStart()
}
