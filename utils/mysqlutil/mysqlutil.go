package mysqlutil

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"strconv"
)

// Insert data
func Insert(db *sql.DB, sqlstr string, args ...interface{}) (int64, error) {
	result, err := db.Exec(sqlstr, args...)
	if err != nil {
		return 0, err
	}

	return result.LastInsertId()
}

// Update or delete
func Exec(db *sql.DB, sqlstr string, args ...interface{}) (int64, error) {
	result, err := db.Exec(sqlstr, args...)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

func FetchRows(db *sql.DB, sqlstr string, args ...interface{}) ([]*map[string]string, error) {
	rows, err := db.Query(sqlstr, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return nil, err
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
			return nil, err
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

// fixme: might need to add mutex when high concurrency occur.
func InsertCityProvince(dbConn *sql.DB, city string, province string) (cityId int64, provinceId int64, err error){
	// query the province
	provinceRows, err := FetchRows(dbConn, "select id from province where name = ?", province)
	if err != nil {
		return 0, 0, err
	}
	if len(provinceRows) == 0 {
		// If province not exist, insert the province
		provinceId, err = Insert(dbConn, "insert into province(name) values(?)", province)
		if err != nil {
			return 0, 0, err
		}
	} else {
		// province already exist, get its id
		provinceId, _ = strconv.ParseInt((*provinceRows[0])["id"], 10, 64)
	}

	// query the city
	cityRows, err := FetchRows(dbConn, "select id from city where name = ? and province_id = ?", city, provinceId)
	if err != nil {
		return 0, 0, err
	}
	if len(cityRows) > 0 {
		// If the record already exist, report err
		return cityId, provinceId, &CityProvinceExistError{}
	} else {
		// Insert the city
		cityId, err = Insert(dbConn, "insert into city(name, province_id) values(?, ?)", city, provinceId)
		if err != nil {
			return 0, 0, err
		}
	}

	return cityId, provinceId, nil
}