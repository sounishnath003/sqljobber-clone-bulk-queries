package utils

import (
	"log"
	"os"
)

func GetEnv(key string, fallback interface{}) any {
	if ok := os.Getenv(key); len(ok) > 0 {
		log.Printf("finding KEY=%s value from OS environment\n", key)
		return ok
	}
	log.Printf("KEY=%s not found. setting the fallback value\n", key)
	return fallback
}
