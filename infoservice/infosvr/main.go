//go:generate protoc -I ../infoservice --go_out=plugins=grpc:../infoservice ../infoservice/infoservice.proto

package main

import (
	"cityinfo/configs"
	pb "cityinfo/infoservice/infoservice"
	"cityinfo/utils/mysqlutil"
	"context"
	"database/sql"
	"fmt"
	"github.com/gomodule/redigo/redis"
	"google.golang.org/grpc"
	"log"
	"net"
	"strconv"
)

type server struct {
	pb.UnimplementedCityManagerServer
	db *sql.DB
	redisPool *redis.Pool
}

// NewToDoServiceServer creates
func NewCityManagerServer(db *sql.DB, redisPool *redis.Pool) pb.CityManagerServer {
	return &server{db: db, redisPool: redisPool}
}

func (s *server) FetchCities(ctx context.Context, request *pb.FetchCitiesRequest) (*pb.FetchCitiesReply, error) {
	provinceId := request.GetProvinceId()

	var cities []*pb.City

	redisConn := s.redisPool.Get()
	defer redisConn.Close()

	// Query from redis.
	values, err := redis.Values(redisConn.Do("zrange", provinceId, "0", "-1"))

	if err == nil && len(values) > 0 {
		for _, v := range values {
			cityName := string(v.([]byte))
			cities = append(cities, &pb.City{Name: cityName})
		}
	} else {
		// Could not query from redis, then query from mysql.
		rows, err := mysqlutil.FetchRows(s.db,"select name from city where province_id = ?", provinceId)
		if err != nil {
			fmt.Println("Could not query from mysql:", err)
			// todo handle err
		}
		for _, row := range rows {
			cityName := (*row)["name"]
			cities = append(cities, &pb.City{Name: cityName})

			// Sync to redis
			_, err = redisConn.Do("zadd", provinceId, 0, cityName)
			if err != nil {
				fmt.Println("Error when store in Redis:", err )
				// todo handle err
			}
		}
	}

	return &pb.FetchCitiesReply{Cities: cities}, nil
}

func (s *server) AddCities(ctx context.Context, request *pb.AddCitiesRequest) (*pb.AddCitiesReply, error) {
	cities := request.Cities
	var results []*pb.OptionResult

	redisConn := s.redisPool.Get()
	defer redisConn.Close()

	for _, city := range cities {
		result := new(pb.OptionResult)

		cityName := city.Name
		provinceName := city.Province.Name

		// Insert to mysql
		_, provinceId, err := mysqlutil.InsertCityProvince(s.db, cityName, provinceName)
		if err != nil {
			if _, ok := err.(*mysqlutil.CityProvinceExistError); ok {
				result.Status = configs.CITY_ALREADY_EXIST
			} else {
				result.Status = configs.MYSQL_ERR
			}
			result.Msg = err.Error()

			results = append(results, result)
			continue
		}

		// Sync to redis
		_, err = redisConn.Do("zadd", int32(provinceId), 0, cityName)
		if err != nil {
			result.Status = configs.REDIS_ERR
			result.Msg = err.Error()
		}

		result.Status = 0
		result.Msg = "ok"
		results = append(results, result)
	}

	return &pb.AddCitiesReply{Result: results}, nil
}

func (s *server) DelCities(ctx context.Context, request *pb.DelCitiesRequest) (*pb.DelCitiesReply, error) {
	cityIds := request.CityIds
	var results []*pb.OptionResult

	redisConn := s.redisPool.Get()
	defer redisConn.Close()

	for _, cid := range cityIds {
		// Query the existence of city
		rows, err := mysqlutil.FetchRows(s.db,"select name, province_id from city where id = ?", cid)
		if err != nil {
			fmt.Println("Could not query from mysql:", err)
		}
		if len(rows) == 0 {
			// This city do not exist.
			results = append(results, &pb.OptionResult{Status: configs.CITY_NOT_EXIST, Msg: "city not exist!"})
			continue
		}
		cityName := (*rows[0])["name"]
		provinceId, _ := strconv.Atoi((*rows[0])["province_id"])

		// Del from mysql
		_, err = mysqlutil.Exec(s.db, "delete from city where id = ?", cid)
		if err != nil {
			results = append(results, &pb.OptionResult{Status:  configs.MYSQL_ERR, Msg: err.Error()})
			continue
		}

		// Sync del to redis
		_, err = redisConn.Do("zrem", int32(provinceId), 0, cityName)
		if err != nil {
			results = append(results, &pb.OptionResult{Status:  configs.REDIS_ERR, Msg: err.Error()})
			continue
		}

		results = append(results,  &pb.OptionResult{Status:  0, Msg: "ok"})
	}

	return &pb.DelCitiesReply{Result: results}, nil
}

func (s *server) DelProvince(ctx context.Context, request *pb.DelProvinceRequest) (*pb.DelProvinceReply, error) {
	pid := request.ProvinceId

	redisConn := s.redisPool.Get()
	defer redisConn.Close()

	tx, err := s.db.Begin()
	if err != nil {
		return &pb.DelProvinceReply{Result: &pb.OptionResult{Status:  configs.MYSQL_ERR, Msg: err.Error()}}, err
	}

	// Del cities of province from mysql
	_, err = mysqlutil.Exec(s.db, "delete from city where province_id = ?", pid)
	if err != nil {
		tx.Rollback()
		return &pb.DelProvinceReply{Result: &pb.OptionResult{Status:  configs.MYSQL_ERR, Msg: err.Error()}}, err
	}

	// Del province from mysql
	rowsAffected, err := mysqlutil.Exec(s.db, "delete from province where id = ?", pid)
	if err != nil {
		tx.Rollback()
		return &pb.DelProvinceReply{Result: &pb.OptionResult{Status:  configs.MYSQL_ERR, Msg: err.Error()}}, err
	}
	if rowsAffected == 0 {
		tx.Rollback()
		return &pb.DelProvinceReply{Result: &pb.OptionResult{Status:  configs.PROVINCE_NOT_EXIST, Msg: "province not exist!"}}, err
	}

	// Sync del zset in redis
	_, err = redisConn.Do("zremrangebyrank", pid, 0, -1)
	if err != nil {
		tx.Rollback()
		return &pb.DelProvinceReply{Result: &pb.OptionResult{Status:  configs.REDIS_ERR, Msg: err.Error()}}, err
	}

	tx.Commit()

	return &pb.DelProvinceReply{Result: &pb.OptionResult{Status: 0, Msg: "ok"}}, err
}

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
	pb.RegisterCityManagerServer(s, &server{db: db, redisPool: redisPool})
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
