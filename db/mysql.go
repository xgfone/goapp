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

	"github.com/xgfone/go-log"
	"github.com/xgfone/sqlx"
)

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
