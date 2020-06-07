package configs

const (
	// Envirment param
	KAFKA_BROKER = "localhost:9092"
	KAFKA_TOPIC = "test6"

	CITYINFO_URL = "http://www.hotelaah.com/dijishi.html"

	MYSQL_USERNAME = "root"
	MYSQL_PASSWORD = "raycghuang"
	MYSQL_NETWORK = "tcp"
	MYSQL_SERVER = "127.0.0.1"
	MYSQL_PORT = 3306
	MYSQL_DB = "city_and_province"

	REDIS_HOST = "127.0.0.1"
	REDIS_PORT = "6379"
	REDIS_NETWORK = "tcp"
	POOL_MAX_CONN = 100

	GRPC_SVR_ADDR = "localhost:50051"

	// Err status code
	CITY_ALREADY_EXIST = -10000
	CITY_NOT_EXIST = -10001
	PROVINCE_NOT_EXIST = -10002
	MYSQL_ERR = -10002
	REDIS_ERR = -10003
)