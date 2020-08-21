// Copyright 2020 xgfone
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

package db

import (
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/xgfone/go-tools/v7/lifecycle"
	"github.com/xgfone/goapp/log"
	"github.com/xgfone/sqlx"
)

// Location sets the location in sql connection url if missing.
var Location = time.UTC

// SetMySQLLocation sets the loc argument in the mysql connection url if missing.
func SetMySQLLocation(mysqlConnURL string, loc *time.Location) string {
	if loc == nil {
		return mysqlConnURL
	}

	if index := strings.IndexByte(mysqlConnURL, '?') + 1; index > 0 {
		query, err := url.ParseQuery(mysqlConnURL[index:])
		if err == nil && query.Get("loc") == "" {
			query.Set("loc", loc.String())
			return mysqlConnURL[:index] + query.Encode()
		}
		return mysqlConnURL
	}

	return fmt.Sprintf("%s?loc=%s", mysqlConnURL, loc.String())
}

// InitMysqlDB initializes the MySQL DB.
//
// If configs is nil, it will use DefaultConfig as the default.
func InitMysqlDB(connURL string, configs ...Config) *sqlx.DB {
	connURL = SetMySQLLocation(connURL, Location)
	db, err := sqlx.Open("mysql", connURL)
	if err != nil {
		log.Fatal("failed to conenct to mysql", log.E(err))
	}

	if configs == nil {
		configs = DefaultConfig
	}

	for _, c := range configs {
		c(db)
	}

	return db
}

// DefaultConfig is the default config.
var DefaultConfig = []Config{Ping(), OnExit(), MaxOpenConns(100)}

// Config is used to set the sqlx.DB.
type Config func(*sqlx.DB)

// MaxIdleConns returns a Config to set the maximum number of the idle connection.
func MaxIdleConns(n int) Config {
	return func(db *sqlx.DB) { db.SetMaxIdleConns(n) }
}

// MaxOpenConns returns a Config to set the maximum number of the open connection.
func MaxOpenConns(maxnum int) Config {
	return func(db *sqlx.DB) { db.SetMaxOpenConns(maxnum) }
}

// ConnMaxLifetime returns a Config to set the maximum lifetime of the connection.
func ConnMaxLifetime(d time.Duration) Config {
	return func(db *sqlx.DB) { db.SetConnMaxLifetime(d) }
}

// ConnMaxIdleTime returns a Config to set the maximum idle time of the connection.
func ConnMaxIdleTime(d time.Duration) Config {
	return func(db *sqlx.DB) { db.SetConnMaxIdleTime(d) }
}

// LogInterceptor returns a Config to set the log interceptor for sqlx.DB.
func LogInterceptor(debug, logArgs bool) Config {
	return func(db *sqlx.DB) {
		if debug {
			db.Interceptor = sqlx.LogInterceptor(log.Debugf, logArgs)
		} else {
			db.Interceptor = sqlx.LogInterceptor(log.Infof, logArgs)
		}
	}
}

// Ping returns a Config to ping the db server, which exits the program
// when fails.
func Ping() Config {
	return func(db *sqlx.DB) {
		if err := db.Ping(); err != nil {
			log.Fatal("failed to ping mysql", log.E(err))
		}
	}
}

// OnExit returns a Config to register a close callback into lifecycle.Manager.
func OnExit() Config {
	return func(db *sqlx.DB) { lifecycle.Register(func() { db.Close() }) }
}
