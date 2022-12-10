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
	"github.com/xgfone/gconf/v6"
	"github.com/xgfone/go-atexit"
	"github.com/xgfone/go-log"
	"github.com/xgfone/go-log/logf"
	"github.com/xgfone/sqlx"
)

// Connection is the configuration option to connect to the sql database.
var Connection = gconf.StrOpt("connection", "The URL connection to the sql database, user:password@tcp(ip:port)/db")

func init() {
	sqlx.DefaultConfigs = append(sqlx.DefaultConfigs,
		LogInterceptor(true, true),
		OnExit())
}

// LogInterceptor returns a Config to set the log interceptor for sqlx.DB.
func LogInterceptor(debug, logArgs bool) sqlx.Config {
	return func(db *sqlx.DB) {
		if debug {
			db.Interceptor = sqlx.LogInterceptor(logf.Tracef, logArgs)
		} else {
			db.Interceptor = sqlx.LogInterceptor(logf.Infof, logArgs)
		}
	}
}

// OnExit returns a Config to register a close callback which will be called
// when the program exits.
func OnExit() sqlx.Config {
	return func(db *sqlx.DB) { atexit.OnExit(func() { db.Close() }) }
}

// InitMysqlDB initializes the mysql connection.
func InitMysqlDB(connURL string, configs ...sqlx.Config) *sqlx.DB {
	if configs == nil {
		configs = sqlx.DefaultConfigs
	}

	connURL = sqlx.SetConnURLLocation(connURL, sqlx.Location)
	db, err := sqlx.Open("mysql", connURL, configs...)
	if err != nil {
		log.Fatal().Str("conn", connURL).Err(err).
			Printf("fail to open the mysql connection")
	}
	return db
}
