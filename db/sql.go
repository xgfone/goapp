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
	"runtime"
	"time"

	"github.com/xgfone/go-atexit"
	"github.com/xgfone/go-log"
	"github.com/xgfone/sqlx"
)

// Location sets the location in sql connection url if missing.
var Location = time.UTC

// DefaultConfig is the default config.
var DefaultConfig = []Config{OnExit(), Ping(), MaxOpenConns(0)}

// Config is used to set the sqlx.DB.
type Config func(*sqlx.DB)

// MaxIdleConns returns a Config to set the maximum number of the idle connection.
func MaxIdleConns(n int) Config {
	return func(db *sqlx.DB) { db.SetMaxIdleConns(n) }
}

// MaxOpenConns returns a Config to set the maximum number of the open connection.
//
// If maxnum is equal to or less than 0, it is runtime.NumCPU()*2 by default.
func MaxOpenConns(maxnum int) Config {
	if maxnum <= 0 {
		maxnum = runtime.NumCPU() * 2
	}
	return func(db *sqlx.DB) { db.SetMaxOpenConns(maxnum) }
}

// ConnMaxLifetime returns a Config to set the maximum lifetime of the connection.
func ConnMaxLifetime(d time.Duration) Config {
	return func(db *sqlx.DB) { db.SetConnMaxLifetime(d) }
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
			log.Fatal().Kv("err", err).Printf("failed to ping mysql")
		}
	}
}

// OnExit returns a Config to register a close callback which will be called
// when the program exits.
func OnExit() Config {
	return func(db *sqlx.DB) { atexit.Register(func() { db.Close() }) }
}
