// Package main implements a client for Greeter service.
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"cityinfo/configs"
	pb "cityinfo/infoservice/infoservice"
	"google.golang.org/grpc"
)

func main() {
	// Set up a connection to the server.
	conn, err := grpc.Dial(configs.GRPC_SVR_ADDR, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := pb.NewCityManagerClient(conn)

	// Contact the server and print out its response.
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	//// Test fetch city.
	//r, err := c.FetchCities(ctx, &pb.FetchCitiesRequest{Province: &pb.Province{Id: 218}})
	//if err != nil {
	//	log.Fatalf("could not greet: %v", err)
	//}
	//log.Printf("Greeting: %v", r.GetCities())

	// Test add city.
	fmt.Println("@@@@@@@@@@@@@@")
	resp, err := c.AddCities(ctx, &pb.AddCitiesRequest{Cities: []*pb.City{{Name: "test1", Province: &pb.Province{Name: "山东省"}},
		{Name: "test2", Province: &pb.Province{Name: "山东省"}},
	}})
	if err != nil {
		log.Fatalf("could not greet: %v", err)
	}
	log.Printf("Greeting: %v", resp.Result)
	fmt.Println("@@@@@@@@@@@@@@")

	//// Test del city.
	//fmt.Println("@@@@@@@@@@@@@@")
	//resp, err := c.DelCities(ctx, &pb.DelCitiesRequest{CityIds: []int32{2754,2755}})
	//if err != nil {
	//	log.Fatalf("could not greet: %v", err)
	//}
	//log.Printf("Greeting: %v", resp.Result)
	//fmt.Println("@@@@@@@@@@@@@@")



	//r, err = c.SayHelloAgain(ctx, &pb.HelloRequest{Name: name})
	//if err != nil {
	//	log.Fatalf("could not greet: %v", err)
	//}
	//log.Printf("Greeting: %s", r.GetMessage())
}
