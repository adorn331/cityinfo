# README

* task如下图:
[![NiTPpQ.png](https://s1.ax1x.com/2020/06/16/NiTPpQ.png)](https://imgchr.com/i/NiTPpQ)

* 目录结构
```
cityinfo
├── echosvr          task0：echo svr
├── infocollector    task1：城市信息收集及存储
├── cityservice      task2：城市信息gRPC服务
│   ├── cmd            运行svr
│   ├── loadtesting    压力测试脚本&结果
│   ├── proto          pb文件
│   └── service        gRPC svr
├── configs          相关配置
└── utils            相关utils
    ├── crawler        城市信息网页爬虫
    ├── emailutil      发送邮件相关
    ├── kafkautil      kafka相关
    ├── logger         日志相关
    └── mysqlutil      mysql相关
```

* 运行 task1
    * `go run infocollector/main.go`

* 运行 task2
    * `go run cityservice/cmd/main.go`
    
* 需要 kafka / redis / mysql 依赖, 相关配置在 configs/configs.go
    * mysql schema
    ```sql
  CREATE TABLE province(
     id INT UNSIGNED AUTO_INCREMENT,
     name VARCHAR(40) NOT NULL,
     PRIMARY KEY (id)
  )ENGINE=InnoDB DEFAULT CHARSET=utf8;
  
  CREATE TABLE city(
     id INT UNSIGNED AUTO_INCREMENT,
     name VARCHAR(40) NOT NULL,
     province_id INT UNSIGNED,
     PRIMARY KEY (id),
     foreign key(province_id) references province(id)
  )ENGINE=InnoDB DEFAULT CHARSET=utf8;