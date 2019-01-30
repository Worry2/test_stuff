package main

import (
	"github.com/micro/go-log"
	"github.com/micro/go-micro"
	"github.com/tahkapaa/test_stuff/micro/example/handler"
	"github.com/tahkapaa/test_stuff/micro/example/subscriber"

	example "github.com/tahkapaa/test_stuff/micro/example/proto/example"
)

func main() {
	// New Service
	service := micro.NewService(
		micro.Name("go.micro.srv.example"),
		micro.Version("latest"),
	)

	// Initialise service
	service.Init()

	// Register Handler
	example.RegisterExampleHandler(service.Server(), new(handler.Example))

	// Register Struct as Subscriber
	micro.RegisterSubscriber("go.micro.srv.example", service.Server(), new(subscriber.Example))

	// Register Function as Subscriber
	micro.RegisterSubscriber("go.micro.srv.example", service.Server(), subscriber.Handler)

	// Run service
	if err := service.Run(); err != nil {
		log.Fatal(err)
	}
}
