package main

import (
	"cityinfo/configs"
	pb "cityinfo/infoservice/infoservice"
	"cityinfo/utils/mysqlutil"
	"context"

	"github.com/gomodule/redigo/redis"
	"github.com/rafaeljusto/redigomock"
	"gopkg.in/DATA-DOG/go-sqlmock.v1"
	"reflect"
	"testing"
)

func TestServer_AddCities(t *testing.T) {
	ctx := context.Background()
	db, dbMock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	redisMock := redigomock.NewConn()
	poolMock := &redis.Pool{
		// Return the same connection mock for each Get() call.
		Dial:    func() (redis.Conn, error) { return redisMock, nil },
		MaxIdle: configs.POOL_MAX_CONN,
	}

	s := NewCityManagerServer(db, poolMock)

	type args struct {
		ctx context.Context
		req *pb.AddCitiesRequest
	}

	// Prepare test case table
	tests := []struct {
		name    string
		s       pb.CityManagerServer
		args    args
		mock    func()
		want    *pb.AddCitiesReply
		wantErr bool
	}{
		{
			name: "Add OK",
			s:    s,
			args: args{
				ctx: ctx,
				req: &pb.AddCitiesRequest{
					Cities: []*pb.City{
						{Name: "城市1", Province: &pb.Province{Name: "山东省"}},
						{Name: "城市2", Province: &pb.Province{Name: "山东省"}},
					},
				},
			},
			mock: func() {
				// Mock Mysql
				dbMock.ExpectQuery("select .* from province").WithArgs("山东省").
					WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow(1, "山东省"))
				dbMock.ExpectQuery("select .* from city").WithArgs("城市1", 1).
					WillReturnRows(sqlmock.NewRows([]string{"id", "name", "province_id"}))
				dbMock.ExpectExec("insert into city").WithArgs("城市1", 1).
					WillReturnResult(sqlmock.NewResult(1, 1))

				dbMock.ExpectQuery("select .* from province").WithArgs("山东省").
					WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow(1, "山东省"))
				dbMock.ExpectQuery("select .* from city").WithArgs("城市2", 1).
					WillReturnRows(sqlmock.NewRows([]string{"id", "name", "province_id"}))
				dbMock.ExpectExec("insert into city").WithArgs("城市2", 1).
					WillReturnResult(sqlmock.NewResult(2, 1))

				// Mock redis
				redisMock.Command("zadd", int32(1), 0, "城市1").Expect("OK")
				redisMock.Command("zadd", int32(1), 0, "城市2").Expect("OK")
			},
			want: &pb.AddCitiesReply{
				Result: []*pb.OptionResult{
					{Status: 0, Msg: "ok"},
					{Status: 0, Msg: "ok"},
				},
			},
		},
		{
			name: "Duplicated add",
			s:    s,
			args: args{
				ctx: ctx,
				req: &pb.AddCitiesRequest{
					Cities: []*pb.City{
						{Name: "城市1", Province: &pb.Province{Name: "山东省"}},
						{Name: "城市1", Province: &pb.Province{Name: "山东省"}},
					},
				},
			},
			mock: func() {
				// Mock Mysql
				dbMock.ExpectQuery("select .* from province").WithArgs("山东省").
					WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow(1, "山东省"))
				dbMock.ExpectQuery("select .* from city").WithArgs("城市1", 1).
					WillReturnRows(sqlmock.NewRows([]string{"id", "name", "province_id"}))
				dbMock.ExpectExec("insert into city").WithArgs("城市1", 1).
					WillReturnResult(sqlmock.NewResult(1, 1))

				dbMock.ExpectQuery("select .* from province").WithArgs("山东省").
					WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow(1, "山东省"))
				dbMock.ExpectQuery("select .* from city").WithArgs("城市1", 1).
					WillReturnRows(sqlmock.NewRows([]string{"id", "name", "province_id"}).AddRow(1, "城市1", 1))
				dbMock.ExpectExec("insert into city").WithArgs("城市1", 1).
					WillReturnError(&mysqlutil.CityProvinceExistError{})

				// Mock redis
				redisMock.Command("zadd", int32(1), 0, "城市1").Expect("OK")
			},
			want: &pb.AddCitiesReply{
				Result: []*pb.OptionResult{
					{Status: 0, Msg: "ok"},
					{Status: configs.CITY_ALREADY_EXIST, Msg: "such city and province already exist!"},
				},
			},
		},
	}

	// Start testing
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mock()
			got, err := tt.s.AddCities(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("CityManagerServer.AddCities() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err == nil && !reflect.DeepEqual(got, tt.want) {
				t.Errorf("CityManagerServer.AddCities() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestServer_FetchCities(t *testing.T) {
	ctx := context.Background()
	db, dbMock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	redisMock := redigomock.NewConn()
	poolMock := &redis.Pool{
		// Return the same connection mock for each Get() call.
		Dial:    func() (redis.Conn, error) { return redisMock, nil },
		MaxIdle: configs.POOL_MAX_CONN,
	}
	defer poolMock.Close()

	s := NewCityManagerServer(db, poolMock)

	type args struct {
		ctx context.Context
		req *pb.FetchCitiesRequest
	}

	// Prepare test case table
	tests := []struct {
		name    string
		s       pb.CityManagerServer
		args    args
		mock    func()
		want    *pb.FetchCitiesReply
		wantErr bool
	}{
		{
			name: "OK: Get from Redis",
			s:    s,
			args: args{
				ctx: ctx,
				req: &pb.FetchCitiesRequest{
					ProvinceId: int32(1),
				},
			},
			mock: func() {
				//Mock redis
				redisMock.Command("zrange", int32(1), "0", "-1").ExpectStringSlice("城市1", "城市2", "城市3")

			},
			want: &pb.FetchCitiesReply{
				Cities: []*pb.City{
					{Name: "城市1"},
					{Name: "城市2"},
					{Name: "城市3"},
				},
			},
		},
		{
			name: "OK: Get from MySQL",
			s:    s,
			args: args{
				ctx: ctx,
				req: &pb.FetchCitiesRequest{
					ProvinceId: int32(1),
				},
			},
			mock: func() {
				// Mock redis
				redisMock.Command("zrange", int32(1), "0", "-1")

				// Mock mysql
				rows := sqlmock.NewRows([]string{"id", "name"}).
					AddRow(1, "城市1").
					AddRow(2, "城市2").
					AddRow(3, "城市3")
				dbMock.ExpectQuery("select .* from city").WithArgs(int32(1)).
					WillReturnRows(rows)

				// Mock sync to redis
				redisMock.Command("zadd", int32(1), 0, "城市1").Expect("OK")
				redisMock.Command("zadd", int32(1), 0, "城市2").Expect("OK")
				redisMock.Command("zadd", int32(1), 0, "城市3").Expect("OK")
			},
			want: &pb.FetchCitiesReply{
				Cities: []*pb.City{
					{Name: "城市1"},
					{Name: "城市2"},
					{Name: "城市3"},
				},
			},
		},
		{
			name: "Not Exist",
			s:    s,
			args: args{
				ctx: ctx,
				req: &pb.FetchCitiesRequest{
					ProvinceId: int32(666),
				},
			},
			mock: func() {
				// Mock redis
				redisMock.Command("zrange", int32(1), "0", "-1")

				// Mock mysql
				dbMock.ExpectQuery("select .* from city").WithArgs(int32(666)).
					WillReturnRows(sqlmock.NewRows([]string{"id", "name", "province_id"}))
			},
			want: &pb.FetchCitiesReply{
				Cities: nil,
			},
		},
	}

	// Start testing
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mock()
			got, err := tt.s.FetchCities(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("CityManagerServer.FetchCities() error = %v, wantErr %v(testcase name: %v)", err, tt.wantErr, tt.name)
				return
			}
			if err == nil && !reflect.DeepEqual(got, tt.want) {
				t.Errorf("CityManagerServer.FetchCities() = %v, want %v (testcase name: %v)", got, tt.want, tt.name)
			}
		})
	}
}

func TestServer_DelCities(t *testing.T) {
	ctx := context.Background()
	db, dbMock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	redisMock := redigomock.NewConn()
	poolMock := &redis.Pool{
		// Return the same connection mock for each Get() call.
		Dial:    func() (redis.Conn, error) { return redisMock, nil },
		MaxIdle: configs.POOL_MAX_CONN,
	}
	defer poolMock.Close()

	s := NewCityManagerServer(db, poolMock)

	type args struct {
		ctx context.Context
		req *pb.DelCitiesRequest
	}

	// Prepare test case table
	tests := []struct {
		name    string
		s       pb.CityManagerServer
		args    args
		mock    func()
		want    *pb.DelCitiesReply
		wantErr bool
	}{
		{
			name: "OK",
			s:    s,
			args: args{
				ctx: ctx,
				req: &pb.DelCitiesRequest{
					CityIds: []int32{1, 2, 3},
				},
			},
			mock: func() {
				dbMock.ExpectQuery("select .* from city").WithArgs(int32(1)).
					WillReturnRows(sqlmock.NewRows([]string{"id", "name", "province_id"}).AddRow(1, "城市1", 1))
				dbMock.ExpectExec("delete from city").WithArgs(int32(1)).
					WillReturnResult(sqlmock.NewResult(1, 1))
				redisMock.Command("zrem", int32(1), 0, "城市1").Expect("OK")

				dbMock.ExpectQuery("select .* from city").WithArgs(int32(2)).
					WillReturnRows(sqlmock.NewRows([]string{"id", "name", "province_id"}).AddRow(2, "城市2", 1))
				dbMock.ExpectExec("delete from city").WithArgs(int32(2)).
					WillReturnResult(sqlmock.NewResult(2, 1))
				redisMock.Command("zrem", int32(1), 0, "城市2").Expect("OK")

				dbMock.ExpectQuery("select .* from city").WithArgs(int32(3)).
					WillReturnRows(sqlmock.NewRows([]string{"id", "name", "province_id"}).AddRow(3, "城市3", 1))
				dbMock.ExpectExec("delete from city").WithArgs(int32(3)).
					WillReturnResult(sqlmock.NewResult(3, 1))
				redisMock.Command("zrem", int32(1), 0, "城市3").Expect("OK")
			},
			want: &pb.DelCitiesReply{
				Result: []*pb.OptionResult{
					{Status: 0, Msg: "ok"},
					{Status: 0, Msg: "ok"},
					{Status: 0, Msg: "ok"},
				},
			},
		},
		{
			name: "Not Exist",
			s:    s,
			args: args{
				ctx: ctx,
				req: &pb.DelCitiesRequest{
					CityIds: []int32{1, 778},
				},
			},
			mock: func() {
				dbMock.ExpectQuery("select .* from city").WithArgs(int32(1)).
					WillReturnRows(sqlmock.NewRows([]string{"id", "name", "province_id"}).AddRow(1, "城市1", 1))
				dbMock.ExpectExec("delete from city").WithArgs(int32(1)).
					WillReturnResult(sqlmock.NewResult(1, 1))
				redisMock.Command("zrem", int32(1), 0, "城市1").Expect("OK")

				dbMock.ExpectQuery("select .* from city").WithArgs(int32(778)).
					WillReturnRows(sqlmock.NewRows([]string{"id", "name", "province_id"}))
				dbMock.ExpectExec("delete from city").WithArgs(int32(2)).
					WillReturnResult(sqlmock.NewResult(-1, 0))
			},
			want: &pb.DelCitiesReply{
				Result: []*pb.OptionResult{
					{Status: 0, Msg: "ok"},
					{Status: configs.CITY_NOT_EXIST, Msg: "city not exist!"},
				},
			},
		},
	}

	// Start testing
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mock()
			got, err := tt.s.DelCities(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("CityManagerServer.DelCities() error = %v, wantErr %v(testcase name: %v)", err, tt.wantErr, tt.name)
				return
			}
			if err == nil && !reflect.DeepEqual(got, tt.want) {
				t.Errorf("CityManagerServer.DelCities() = %v, want %v (testcase name: %v)", got, tt.want, tt.name)
			}
		})
	}
}

func TestServer_DelProvince(t *testing.T) {
	ctx := context.Background()
	db, dbMock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	redisMock := redigomock.NewConn()
	poolMock := &redis.Pool{
		// Return the same connection mock for each Get() call.
		Dial:    func() (redis.Conn, error) { return redisMock, nil },
		MaxIdle: configs.POOL_MAX_CONN,
	}
	defer poolMock.Close()

	s := NewCityManagerServer(db, poolMock)

	type args struct {
		ctx context.Context
		req *pb.DelProvinceRequest
	}

	// Prepare test case table
	tests := []struct {
		name    string
		s       pb.CityManagerServer
		args    args
		mock    func()
		want    *pb.DelProvinceReply
		wantErr bool
	}{
		{
			name: "OK",
			s:    s,
			args: args{
				ctx: ctx,
				req: &pb.DelProvinceRequest{
					ProvinceId: int32(1),
				},
			},
			mock: func() {
				dbMock.ExpectBegin()
				dbMock.ExpectExec("delete from city").WithArgs(int32(1)).
					WillReturnResult(sqlmock.NewResult(1, 3))
				dbMock.ExpectExec("delete from province").WithArgs(int32(1)).
					WillReturnResult(sqlmock.NewResult(1, 1))
				dbMock.ExpectCommit()
				redisMock.Command("zremrangebyrank", int32(1), 0, -1).Expect("OK")
			},
			want: &pb.DelProvinceReply{
				Result: &pb.OptionResult{Status: 0, Msg: "ok"},
			},
		},
		{
			name: "Not exist",
			s:    s,
			args: args{
				ctx: ctx,
				req: &pb.DelProvinceRequest{
					ProvinceId: int32(666),
				},
			},
			mock: func() {
				dbMock.ExpectBegin()
				dbMock.ExpectExec("delete from city").WithArgs(int32(666)).
					WillReturnResult(sqlmock.NewResult(-1, 0))
				dbMock.ExpectExec("delete from province").WithArgs(int32(666)).
					WillReturnResult(sqlmock.NewResult(-1, 0))
				dbMock.ExpectRollback()
				//redisMock.Command("zremrangebyrank", int32(1), 0, -1).Expect("OK")
			},
			want: &pb.DelProvinceReply{
				Result: &pb.OptionResult{Status: configs.PROVINCE_NOT_EXIST, Msg: "province not exist!"},
			},
		},
	}

	// Start testing
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mock()
			got, err := tt.s.DelProvince(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("CityManagerServer.DelProvince() error = %v, wantErr %v(testcase name: %v)", err, tt.wantErr, tt.name)
				return
			}
			if err == nil && !reflect.DeepEqual(got, tt.want) {
				t.Errorf("CityManagerServer.DelProvince() = %v, want %v (testcase name: %v)", got, tt.want, tt.name)
			}
		})
	}
}