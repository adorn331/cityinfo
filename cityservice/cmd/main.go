package main

import (
	pb "cityinfo/cityservice/proto"
	"cityinfo/cityservice/service"
	"cityinfo/utils/logger"
	"cityinfo/configs"
	"database/sql"
	"fmt"
	"github.com/gomodule/redigo/redis"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"net"
)

// Run the city service.
func main() {
	lis, err := net.Listen("tcp", configs.GRPC_SVR_ADDR)
	if err != nil {
		logger.Log.Fatal("Fail to listen port", zap.String("reason", err.Error()))
	}

	// Connect to redis pool
	redisPool := redis.NewPool(func() (redis.Conn, error) { return redis.Dial(configs.REDIS_NETWORK, configs.REDIS_HOST+":"+configs.REDIS_PORT) }, configs.POOL_MAX_CONN)
	if err != nil {
		logger.Log.Fatal("Fail to connect redis", zap.String("reason", err.Error()))
	}
	defer redisPool.Close()

	// Open mysql DB
	dbConfig := fmt.Sprintf("%s:%s@%s(%s:%d)/%s",
		configs.MYSQL_USERNAME, configs.MYSQL_PASSWORD, configs.MYSQL_NETWORK,
		configs.MYSQL_SERVER, configs.MYSQL_PORT, configs.MYSQL_DB)
	db, err := sql.Open("mysql", dbConfig)
	if err != nil {
		logger.Log.Fatal("Fail to conect mysql", zap.String("reason", err.Error()))
	}
	defer db.Close()

	s := grpc.NewServer()
	pb.RegisterCityServiceServer(s, service.NewCityServiceServer(db, redisPool))
	if err := s.Serve(lis); err != nil {
		logger.Log.Fatal("Fail to serve", zap.String("reason", err.Error()))
	}
}

