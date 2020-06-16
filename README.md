# README
* Original stask:
[![NiTPpQ.png](https://s1.ax1x.com/2020/06/16/NiTPpQ.png)](https://imgchr.com/i/NiTPpQ)

* how to run task1 ( info-collector that crawls city info and store it)?
    * `go run infocollector/main.go`

* how to run task2 (gRPC city service)?
    * `go run cityservice/cmd/main.go`
    
* might need to set up kafka / redis / mysql locally, relevant configs are presented in configs/configs.go
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