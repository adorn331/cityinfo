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

	// Logger
	LOG_LEVEL = -1 // debug
	LOG_FILE = "/Users/huangchaogang/cityservice.log"
	LOG_TIME_FORMAT = "2006-01-02T15:04:05.999999999Z07:00"

	// Email
	SMTP_HOST = "smtp.163.com"
	SMTP_PORT = "25"
	SMTP_USER = "adorn331@163.com"
	SMTP_PWD = "Codalab2019"
	EMAIL_FROM = "adorn331@163.com"
	EMAIL_FROM_NICKNAME = "CityServiceErrorNotifier"
	EMAIL_SUBJECT = "Service Internal Error"
	EMAIL_CONTENT_TYPE = "Content-Type: text/plain; charset=UTF-8"

	// Err status code
	CITY_ALREADY_EXIST = -10000
	CITY_NOT_EXIST = -10001
	PROVINCE_NOT_EXIST = -10002
	MYSQL_ERR = -10002
	REDIS_ERR = -10003
)

func GetErrEmailReciver() []string {
	return []string{"756730386@qq.com"}
}