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

// Package redis provides some assistant functions about redis.
package redis

import (
	"github.com/go-redis/redis/v8"
	"github.com/xgfone/go-atexit"
	"github.com/xgfone/go-log"
)

// InitRedis initializes the redis client.
func InitRedis(redisConnURL string) *redis.Client {
	options, err := redis.ParseURL(redisConnURL)
	if err != nil {
		log.Fatal().Str("conn", redisConnURL).Err(err).
			Printf("can't parse redis URL connection")
	}

	client := redis.NewClient(options)
	atexit.Register(func() { client.Close() })
	return client
}
