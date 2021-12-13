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
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/xgfone/go-log"
)

// DB is an interface to stands for the general sql.DB.
type DB interface {
	Begin() (*sql.Tx, error)
	BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error)
	Close() error
	Conn(ctx context.Context) (*sql.Conn, error)
	Driver() driver.Driver
	Exec(query string, args ...interface{}) (sql.Result, error)
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	Ping() error
	PingContext(ctx context.Context) error
	Prepare(query string) (*sql.Stmt, error)
	PrepareContext(ctx context.Context, query string) (*sql.Stmt, error)
	Query(query string, args ...interface{}) (*sql.Rows, error)
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	QueryRow(query string, args ...interface{}) *sql.Row
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
	SetConnMaxLifetime(d time.Duration)
	SetMaxIdleConns(n int)
	SetMaxOpenConns(n int)
	Stats() sql.DBStats
}

// Wrapper is a wrapper of the read and write db.
type Wrapper struct {
	Index  int
	Writer DB
	Reader DB
}

type dbPool []Wrapper

func (p dbPool) Len() int {
	return len(p)
}

func (p dbPool) Less(i, j int) bool {
	return p[i].Index < p[j].Index
}

func (p dbPool) Swap(i, j int) {
	db := p[i]
	p[i] = p[j]
	p[j] = db
}

// Pool is a read-write DB pool.
type Pool struct {
	dbs   dbPool
	index func(string) int
}

// NewPool returns a new DB pool.
func NewPool() *Pool {
	return &Pool{dbs: make(dbPool, 0, 32), index: getDBIndexByDefault}
}

func getDBIndexByDefault(key string) (index int) {
	if key == "" {
		panic(errors.New("the key of the db index is empty"))
	}

	switch v := key[len(key)-1]; v {
	case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
		index = int(v - '0')
	case 'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j', 'k', 'l', 'm', 'n',
		'o', 'p', 'q', 'r', 's', 't', 'u', 'v', 'w', 'x', 'y', 'z':
		index = int(v-'a') + 10
	case 'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M', 'N',
		'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z':
		index = int(v-'A') + 36
	}

	return
}

// SetIndexer sets the indexer to get the corresponding DB by the key, that's,
// it will convert the same key to a constant index forever.
//
// The default indexer only converts the key ending with any character of
// "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ".
func (p *Pool) SetIndexer(index func(key string) int) {
	if index == nil {
		panic(errors.New("the index is nil"))
	}
	p.index = index
}

func (p *Pool) addDB(index int, writer, reader DB) {
	if writer == nil && reader == nil {
		panic(fmt.Errorf("the writer and reader are nil with the index '%d'", index))
	}

	if writer == nil {
		writer = reader
	}
	if reader == nil {
		reader = writer
	}

	for i := range p.dbs {
		if p.dbs[i].Index == index {
			p.dbs[i].Writer = writer
			p.dbs[i].Reader = reader
			return
		}
	}

	p.dbs = append(p.dbs, Wrapper{Index: index, Writer: writer, Reader: reader})
	sort.Sort(p.dbs)
}

// AddDB adds a db into the pool.
//
// If reader is nil, the reader DB is the same as writer by default.
//
// Noitce: the index is only used to sort and identify whether two DBs are equal.
func (p *Pool) AddDB(index int, writer DB, reader ...DB) {
	if len(reader) > 0 {
		p.addDB(index, writer, reader[0])
	} else {
		p.addDB(index, writer, nil)
	}
}

// AddReaderDB adds a reader db into the pool.
//
// Noitce: the index is only used to sort and identify whether two DBs are equal.
func (p *Pool) AddReaderDB(index int, reader DB) {
	p.addDB(index, nil, reader)
}

func (p *Pool) getDB(key string) (db Wrapper) {
	if key == "" {
		db = p.dbs[0]
	} else {
		db = p.dbs[p.index(key)%len(p.dbs)]
	}
	log.Debug().Kv("key", key).Kv("index", db.Index).
		Printf("calculating the index by key")
	return
}

// GetDB is short for GetWriterDB.
func (p *Pool) GetDB(key string) DB {
	return p.GetWriterDB(key)
}

// GetWriterDB returns the writer DB by the key.
//
// If no DB, return nil.
func (p *Pool) GetWriterDB(key string) DB {
	return p.getDB(key).Writer
}

// GetReaderDB returns the reader DB by the key.
//
// If no reader, return the writer.
func (p *Pool) GetReaderDB(key string) DB {
	db := p.getDB(key)
	if db.Reader != nil {
		return db.Reader
	}
	return db.Writer
}

// GetAllDBs returns all the dbs.
func (p *Pool) GetAllDBs() []Wrapper {
	return []Wrapper(p.dbs)
}
