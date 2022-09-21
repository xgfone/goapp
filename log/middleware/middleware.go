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

// Package middleware registers some log config options to configure the log
// middleware, which just needs to be imported to do some initializations.
//
// Example:
//
//   import _ "github.com/xgfone/goapp/log/middleware"
//
package middleware

import (
	"github.com/xgfone/gconf/v6"
	"github.com/xgfone/go-apiserver/middleware/logger"
)

var (
	group = gconf.Group("log")

	logReqHeaders  = group.NewBool("reqheaders", false, "If true, log the request headers.")
	logRespHeaders = group.NewBool("respheaders", false, "If true, log the response headers.")

	logReqBodyLen  = group.NewInt("reqbodylen", 0, "If greater than 0, log the request body only if the body length is not greater than it.")
	logRespBodyLen = group.NewInt("respbodylen", 0, "If greater than 0, log the response body only if the body length is not greater than it.")
)

func init() {
	logger.LogReqHeaders = logReqHeaders.Get
	logger.LogReqBodyLen = logReqBodyLen.Get
	logger.LogRespHeaders = logRespHeaders.Get
	logger.LogRespBodyLen = logRespBodyLen.Get
}
