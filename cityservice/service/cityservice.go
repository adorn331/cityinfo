//go:generate protoc -I ../proto --go_out=plugins=grpc:../proto ../proto/cityservice.proto

package service

import (
	pb "cityinfo/cityservice/proto"
	"cityinfo/configs"
	"cityinfo/utils/logger"
	"cityinfo/utils/mysqlutil"
	"context"
	"database/sql"
	"github.com/gomodule/redigo/redis"
	"go.uber.org/zap"
	"strconv"
)

type server struct {
	pb.UnimplementedCityServiceServer
	db *sql.DB
	redisPool *redis.Pool
}

func NewCityServiceServer(db *sql.DB, redisPool *redis.Pool) pb.CityServiceServer {
	return &server{db: db, redisPool: redisPool}
}

func (s *server) RetrieveCities(ctx context.Context, request *pb.RetrieveCitiesRequest) (*pb.RetrieveCitiesReply, error) {
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
			logger.Log.Error("Could not query from mysql", zap.String("reason", err.Error()))
			return nil, err
		}
		for _, row := range rows {
			cityName := (*row)["name"]
			cities = append(cities, &pb.City{Name: cityName})

			// Cache to redis
			_, err = redisConn.Do("zadd", provinceId, 0, cityName)
			if err != nil {
				logger.Log.Error("Could not sync data to redis", zap.String("reason", err.Error()))
			}
		}
	}

	return &pb.RetrieveCitiesReply{Cities: cities}, nil
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
			logger.Log.Error("Could not query from mysql", zap.String("reason", err.Error()))
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
			logger.Log.Error("Could not delete city from mysql", zap.String("reason", err.Error()))
			results = append(results, &pb.OptionResult{Status:  configs.MYSQL_ERR, Msg: err.Error()})
			continue
		}

		// Sync del to redis
		_, err = redisConn.Do("zrem", int32(provinceId), 0, cityName)
		if err != nil {
			logger.Log.Error("Could not sync to redis when deleting cities", zap.String("reason", err.Error()))
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
		logger.Log.Error("Could not begin a tx in mysql", zap.String("reason", err.Error()))
		return &pb.DelProvinceReply{Result: &pb.OptionResult{Status:  configs.MYSQL_ERR, Msg: err.Error()}}, err
	}

	// Del cities of province from mysql
	_, err = mysqlutil.Exec(s.db, "delete from city where province_id = ?", pid)
	if err != nil {
		tx.Rollback()
		logger.Log.Error("Could not delete city from mysql", zap.String("reason", err.Error()))
		return &pb.DelProvinceReply{Result: &pb.OptionResult{Status:  configs.MYSQL_ERR, Msg: err.Error()}}, err
	}

	// Del province from mysql
	rowsAffected, err := mysqlutil.Exec(s.db, "delete from province where id = ?", pid)
	if err != nil {
		tx.Rollback()
		logger.Log.Error("Could not delete province from mysql", zap.String("reason", err.Error()))
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
		logger.Log.Error("Could not delete zset from redis", zap.String("reason", err.Error()))
		return &pb.DelProvinceReply{Result: &pb.OptionResult{Status:  configs.REDIS_ERR, Msg: err.Error()}}, err
	}

	tx.Commit()

	return &pb.DelProvinceReply{Result: &pb.OptionResult{Status: 0, Msg: "ok"}}, err
}
