package mysqlutil

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"strconv"
)

// Insert data
func Insert(db *sql.DB, sqlstr string, args ...interface{}) (int64, error) {
	result, err := db.Exec(sqlstr, args...)
	if err != nil {
		fmt.Println("Fail to exec stmt:", err)
	}

	return result.LastInsertId()
}

// Update or delete
func Exec(db *sql.DB, sqlstr string, args ...interface{}) (int64, error) {
	result, err := db.Exec(sqlstr, args...)
	if err != nil {
		fmt.Println("Fail to exec stmt:", err)
	}
	return result.RowsAffected()
}


func FetchRows(db *sql.DB, sqlstr string, args ...interface{}) ([]*map[string]string, error) {
	rows, err := db.Query(sqlstr, args...)
	if err != nil {
		fmt.Println("Fail to query stmt:", err)
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		fmt.Println("Fail to get columns of row:", err)
	}

	values := make([]sql.RawBytes, len(columns))
	scanArgs := make([]interface{}, len(values))
	var ret []*map[string]string

	for i := range values {
		scanArgs[i] = &values[i]
	}
	for rows.Next() {
		rowMap := make(map[string]string)
		err = rows.Scan(scanArgs...)
		if err != nil {
			fmt.Println("Fail to scan rows:", err)
		}
		var value string

		for i, col := range values {
			if col == nil {
				value = "NULL"
			} else {
				value = string(col)
			}
			rowMap[columns[i]] = value
		}
		ret = append(ret, &rowMap)
	}
	return ret, nil
}


type CityProvinceExistError struct {}

func (e *CityProvinceExistError) Error() string {
	return "such city and province already exist!"
}

func InsertCityProvince(dbConn *sql.DB, city string, province string) (int64, int64, error){
	var provinceId, cityId int64

	// query the province
	provinceRows, err := FetchRows(dbConn, "select * from province where name = ?", province)
	if err != nil {
		fmt.Println("Error when during query:", err )
	}
	if len(provinceRows) == 0 {
		// If province not exist, insert the province
		provinceId, err = Insert(dbConn, "insert into province(name) values(?)", province)
		if err != nil {
			fmt.Println("Error when Insert province to mysql:", err )
		}
	} else {
		// province already exist, get its id
		provinceId, _ = strconv.ParseInt((*provinceRows[0])["id"], 10, 64)
	}

	// query the city
	cityRows, err := FetchRows(dbConn, "select * from city where name = ? and province_id = ?", city, provinceId)
	if err != nil {
		fmt.Println("Error when during query cities:", err )
	}
	if len(cityRows) > 0 {
		// If the record already exist, report err
		return cityId, provinceId, &CityProvinceExistError{}
	} else {
		// Insert the city
		cityId, err = Insert(dbConn, "insert into city(name, province_id) values(?, ?)", city, provinceId)
		if err != nil {
			fmt.Println("Error when Insert city to mysql:", err )
		}
	}

	return cityId, provinceId, nil
}