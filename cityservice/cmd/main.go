package main

import (
	pb "cityinfo/cityservice/proto"
	"cityinfo/configs"
	"database/sql"
	"fmt"
	"google.golang.org/grpc"
	"log"
	"net"
	"github.com/gomodule/redigo/redis"
	"cityinfo/cityservice/service"
)

func main() {
	lis, err := net.Listen("tcp", configs.GRPC_SVR_ADDR)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	// Connect to redis pool
	redisPool := redis.NewPool(func() (redis.Conn, error) { return redis.Dial(configs.REDIS_NETWORK, configs.REDIS_HOST+":"+configs.REDIS_PORT) }, configs.POOL_MAX_CONN)
	if err != nil {
		fmt.Println("Connect to redis failed:", err)
	}
	defer redisPool.Close()

	// Open mysql DB
	dbConfig := fmt.Sprintf("%s:%s@%s(%s:%d)/%s",
		configs.MYSQL_USERNAME, configs.MYSQL_PASSWORD, configs.MYSQL_NETWORK,
		configs.MYSQL_SERVER, configs.MYSQL_PORT, configs.MYSQL_DB)
	db, err := sql.Open("mysql", dbConfig)
	if err != nil {
		fmt.Println("Connection to mysql failed:", err)
	}
	defer db.Close()

	s := grpc.NewServer()
	pb.RegisterCityServiceServer(s, service.NewCityServiceServer(db, redisPool))
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

