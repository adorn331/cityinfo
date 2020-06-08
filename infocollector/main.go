package main

import (
	"cityinfo/configs"
	"cityinfo/utils/crawler"
	"cityinfo/utils/kafkautil"
	"cityinfo/utils/mysqlutil"
	"database/sql"
	"fmt"
	"github.com/garyburd/redigo/redis"
	"github.com/segmentio/kafka-go"
	"time"
)


func main() {
	// Init kafka writer
	kafkaW := kafka.NewWriter(kafka.WriterConfig{
		Brokers: []string{configs.KAFKA_BROKER},
		Topic:   configs.KAFKA_TOPIC,
		Balancer: &kafka.LeastBytes{},
	})
	defer kafkaW.Close()

	// Init kafka reader
	kafkaR := kafka.NewReader(kafka.ReaderConfig{
		Brokers:   []string{configs.KAFKA_BROKER},
		Topic:     configs.KAFKA_TOPIC,
		Partition: 0,
		MinBytes:  10e3, // 10KB
		MaxBytes:  10e6, // 10MB
	})
	defer kafkaR.Close()

	go produceData(kafkaW)

	consumeData(kafkaR)
}

// Fetch data from Website, then write to Kafka.
func produceData(w *kafka.Writer) {
	cityMap, err := crawler.FetchCities(configs.CITYINFO_URL)
	if err != nil {
		fmt.Println("Could not fetch data from Website:", err)
	}

	// Batch send msg to Kafka
	if err := kafkautil.BatchSendMsg(w, cityMap); err != nil {
		fmt.Println("Could not write to Kafka:", err)
	}

	fmt.Println("Write to Kafka Done.")
}

// Keep reading msg from kafka then store to mysqlutil and redis.
func consumeData(r *kafka.Reader) {
	// Connect to mysql
	dbConfig := fmt.Sprintf("%s:%s@%s(%s:%d)/%s",
		configs.MYSQL_USERNAME, configs.MYSQL_PASSWORD, configs.MYSQL_NETWORK,
		configs.MYSQL_SERVER, configs.MYSQL_PORT, configs.MYSQL_DB)
	db, err := sql.Open("mysql", dbConfig)
	if err != nil {
		fmt.Println("Connection to mysql failed:", err)
		return
	}
	defer db.Close()

	// Connect to redis
	redisConn, err := redis.Dial(configs.REDIS_NETWORK, configs.REDIS_HOST+":"+configs.REDIS_PORT)
	if err != nil {
		fmt.Println("Connect to redis failed:", err)
		return
	}
	defer redisConn.Close()

	var offset int64 = 0 // msg offset

	// Keep reading msg from kafka, then write data to mysql and redis.
	for {
		city, province, err := kafkautil.ReadMsg(r, offset)
		if err != nil {
			time.Sleep(time.Millisecond * 100)
			continue
		}

		fmt.Println(offset, city, province)

		// Insert to mysql
		_, provinceId, err := mysqlutil.InsertCityProvince(db, city, province)
		if err != nil {
			fmt.Println("err when inserting to mysql", err)
		}

		// Insert to city and province to redis
		_, err = redisConn.Do("zadd", provinceId, 0, city)
		if err != nil {
			fmt.Println("Error when insert city to Redis:", err )
		}

		offset++
	}
}

//// Insert city - province to mysql and redis.
//func syncToStorage(city string, province string, db *sql.DB, redisConn redis.Conn) error {
//	// Insert to mysql
//	var provinceId int64
//	provinceRows, err := mysqlutil.FetchRows(db, "select * from province where name = ?", province)
//	if err != nil {
//		fmt.Println("Error when during query:", err )
//	}
//	if len(provinceRows) == 0 {
//		// If province not exist, insert the province
//		provinceId, err = mysqlutil.Insert(db, "insert into province(name) values(?)", province)
//		if err != nil {
//			fmt.Println("Error when Insert province to mysqlutil:", err )
//		}
//	} else {
//		// If province already exist, get its id
//		provinceId, _ = strconv.ParseInt((*provinceRows[0])["id"], 10, 64)
//	}
//
//	// Insert the city
//	_, err = mysqlutil.Insert(db, "insert into city(name, province_id) values(?, ?)", city, provinceId)
//	if err != nil {
//		fmt.Println("Error when Insert city to mysql:", err )
//	}
//
//	// Insert to city and province to redis
//	_, err = redisConn.Do("zadd", provinceId, 0, city)
//	if err != nil {
//		fmt.Println("Error when insert city to Redis:", err )
//	}
//
//	return err
//}

