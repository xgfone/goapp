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
	"sync/atomic"

	"github.com/xgfone/gconf/v6"
	"github.com/xgfone/go-defaults"
	"github.com/xgfone/go-sqlx"
	"github.com/xgfone/goapp/log"
)

// Connection is the configuration option to connect to the sql database.
var Connection = gconf.StrOpt("connection", "The URL connection to the sql database, user:password@tcp(ip:port)/db.")

var (
	// LogLevel is the level to log the sql statement and args.
	LogLevel = new(slog.LevelVar)

	// LogArgs is used to decide whether log args when logging the sql statement.
	LogArgs = new(atomic.Bool)
)

func init() { LogLevel.Set(log.LevelTrace) }

func _logsql(msg string, attrs ...slog.Attr) {
	slog.LogAttrs(context.Background(), LogLevel.Level(), msg, attrs...)
}

func logsql(sql string, args []interface{}) (string, []interface{}, error) {
	if LogArgs.Load() {
		_logsql("log sql statement", slog.String("sql", sql), slog.Any("args", args))
	} else {
		_logsql("log sql statement", slog.String("sql", sql))
	}
	return sql, args, nil
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

	db.Interceptor = sqlx.Interceptors{
		sqlx.InterceptorFunc(logsql),
		sqlx.DefaultSqlCollector,
	}
	return db
}
