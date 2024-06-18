// Copyright 2022 xgfone
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package db provides some assistant functions about the database.
package db

import (
	"context"
	"log/slog"

	"github.com/xgfone/gconf/v6"
	"github.com/xgfone/go-defaults"
	"github.com/xgfone/go-defaults/assists"
	"github.com/xgfone/go-sqlx"
	"github.com/xgfone/goapp/log"
)

// Connection is the configuration option to connect to the sql database.
var Connection = gconf.StrOpt("connection", "The URL connection to the sql database, user:password@tcp(ip:port)/db.")

// LogLevel is the level to log the sql statement and args.
var LogLevel = new(slog.LevelVar)

func init() {
	LogLevel.Set(log.LevelTrace)
	sqlx.DefaultConfigs = append(sqlx.DefaultConfigs, LogInterceptor(true), OnExit())
}

// LogInterceptor returns a Config to set the log interceptor for sqlx.DB.
func LogInterceptor(logargs bool) sqlx.Config {
	return func(db *sqlx.DB) { db.Interceptor = logsql(logargs) }
}

func _logsql(msg string, attrs ...slog.Attr) {
	slog.LogAttrs(context.Background(), LogLevel.Level(), msg, attrs...)
}

func logsql(logargs bool) sqlx.InterceptorFunc {
	return func(sql string, args []interface{}) (string, []interface{}, error) {
		if logargs {
			_logsql("log sql statement", slog.String("sql", sql), slog.Any("args", args))
		} else {
			_logsql("log sql statement", slog.String("sql", sql))
		}
		return sql, args, nil
	}
}

// OnExit returns a Config to register a close callback which will be called
// when the program exits.
func OnExit() sqlx.Config {
	return func(db *sqlx.DB) { assists.OnClean(func() { db.Close() }) }
}

// InitMysqlDB initializes the mysql connection.
func InitMysqlDB(connURL string, configs ...sqlx.Config) *sqlx.DB {
	if configs == nil {
		configs = sqlx.DefaultConfigs
	}

	connURL = sqlx.SetConnURLLocation(connURL, defaults.TimeLocation.Get())
	db, err := sqlx.Open("mysql", connURL, configs...)
	if err != nil {
		slog.Error("fail to open the mysql connection", "conn", connURL, "err", err)
		defaults.Exit(1)
	}
	return db
}
