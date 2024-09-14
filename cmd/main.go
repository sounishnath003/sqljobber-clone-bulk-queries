package main

import (
	"log"
	"os"
)

type Config struct {
	PORT     string
	AppName  string
	Hostname string
}

func main() {
	conf := &Config{
		PORT:     getEnv("PORT", "3000"),
		AppName:  getEnv("AppName", "special-appname-001"),
		Hostname: getEnv("Hostname", "apple-macbook-001"),
	}
	log.Printf("config:%+v\n", conf)
	log.Printf("server is up and running on http://localhost:%s\n", conf.PORT)

}

func getEnv(key, fallback string) string {
	val := os.Getenv(key)
	if len(val) == 0 {
		return fallback
	}
	return val
}
