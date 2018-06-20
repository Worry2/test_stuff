package main

import (
	"fmt"
	"log"
	"os"

	"github.com/go-redis/redis"
)

func main() {

	r := os.Getenv("REDIS_IP")
	if len(r) == 0 {
		r = "localhost:6379"
	}

	client := redis.NewClient(&redis.Options{
		Addr: r,
	})

	pong, err := client.Ping().Result()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(pong)

	if err := client.Set("k1", "testi", 0).Err(); err != nil {
		log.Fatal(err)
	}

	s, err := client.Get("k1").Result()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(s)
}
