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

import "github.com/xgfone/gconf/v6"

// SQLDBOpts collects the options of the SQL database.
var SQLDBOpts = []gconf.Opt{
	gconf.StrOpt("connection", "The URL connection to the alarm database, user:password@tcp(127.0.0.1:3306)/db").C(false),
	gconf.IntOpt("maxconnnum", "The maximum number of the connections.").C(false).D(0),
	gconf.BoolOpt("logsqlstmt", "Log the sql statement when executing it.").C(false),
	gconf.BoolOpt("logsqlargs", "Log the arguments of sql statement when executing it.").C(false),
}

// SQLDBOptGroup is the group of the sql database config options.
var SQLDBOptGroup = gconf.NewGroup("database.sql")

// RegisterSQLDBOpts registers the default options of the sql database.
func RegisterSQLDBOpts() {
	SQLDBOptGroup.RegisterOpts(SQLDBOpts...)
}
