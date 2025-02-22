// SPDX-FileCopyrightText: 2021 Lightmeter <hello@lightmeter.io>
//
// SPDX-License-Identifier: AGPL-3.0-only

package dbconn

import (
	"context"
	"database/sql"

	"github.com/rs/zerolog/log"
	_ "gitlab.com/lightmeter/controlcenter/lmsqlite3"
	"gitlab.com/lightmeter/controlcenter/pkg/closers"
	"gitlab.com/lightmeter/controlcenter/util/errorutil"
)

type RoConn struct {
	*sql.DB
}

type RwConn struct {
	*sql.DB
}

// Execute some code in a transaction
func (conn *RwConn) Tx(ctx context.Context, f func(context.Context, *sql.Tx) error) error {
	tx, err := conn.BeginTx(ctx, nil)

	if err != nil {
		return errorutil.Wrap(err)
	}

	if err := f(ctx, tx); err != nil {
		if err := tx.Rollback(); err != nil {
			return errorutil.Wrap(err)
		}

		return errorutil.Wrap(err)
	}

	if err := tx.Commit(); err != nil {
		return errorutil.Wrap(err)
	}

	return nil
}

type RoPooledConn struct {
	closers.Closers
	RoConn

	LocalId int
	stmts   map[interface{}]*sql.Stmt
}

func (c *RoPooledConn) PrepareStmt(query string, key interface{}) error {
	if _, ok := c.stmts[key]; ok {
		log.Panic().Msgf("A prepared statement for %v already exists!", key)
	}

	stmt, err := c.Prepare(query)
	if err != nil {
		return errorutil.Wrap(err)
	}

	c.stmts[key] = stmt
	c.Closers.Add(stmt)

	return nil
}

// GetStmt gets an prepared statement by a key, where the calles does **NOT** own the returned value
func (c *RoPooledConn) GetStmt(key interface{}) *sql.Stmt {
	stmt, ok := c.stmts[key]
	if !ok {
		log.Panic().Msgf("Sql stmt with key %v not implemented!!!!", key)
	}

	return stmt
}

type RoPool struct {
	closers.Closers

	conns []*RoPooledConn
	pool  chan *RoPooledConn
}

func (p *RoPool) ForEach(f func(*RoPooledConn) error) error {
	for _, v := range p.conns {
		if err := f(v); err != nil {
			return errorutil.Wrap(err)
		}
	}

	return nil
}

func (p *RoPool) Acquire() (*RoPooledConn, func()) {
	conn, release, _ := p.AcquireContext(context.Background())
	return conn, release
}

func (p *RoPool) AcquireContext(ctx context.Context) (*RoPooledConn, func(), error) {
	select {
	case c := <-p.pool:
		return c, func() { p.pool <- c }, nil
	case <-ctx.Done():
		return nil, func() {}, errorutil.Wrap(ctx.Err())
	}
}

type PooledPair struct {
	closers.Closers

	RwConn     RwConn
	RoConnPool *RoPool
	Filename   string
}

func OpenRO(filename string, poolSize int) (pool *RoPool, err error) {
	poolChan := make(chan *RoPooledConn, poolSize)
	poolClosers := closers.New()
	conns := []*RoPooledConn{}

	defer func() {
		if err != nil {
			errorutil.UpdateErrorFromCloser(poolClosers, &err)
		}
	}()

	for i := 0; i < poolSize; i++ {
		reader, err := sql.Open("lm_sqlite3", `file:`+filename+`?mode=ro&cache=private&_query_only=true&_loc=auto&_journal=WAL&_sync=OFF&_mutex=no`)
		if err != nil {
			return nil, errorutil.Wrap(err)
		}

		conn := &RoPooledConn{
			RoConn:  RoConn{reader},
			LocalId: i,
			stmts:   map[interface{}]*sql.Stmt{},
			Closers: closers.New(newConnCloser(filename, ROMode, reader)),
		}

		conns = append(conns, conn)
		poolClosers.Add(conn)

		poolChan <- conn
	}

	return &RoPool{
		pool:    poolChan,
		conns:   conns,
		Closers: poolClosers,
	}, nil
}

func OpenRW(filename string) (conn RwConn, err error) {
	// use 5 seconds of busy timeout,
	// allowing multiple writers to act without "database is busy" errors,
	// forcing them to wait for 5seconds before giving up
	writer, err := sql.Open("lm_sqlite3", `file:`+filename+`?mode=rwc&cache=private&_loc=auto&_journal=WAL&_sync=OFF&_mutex=no&_busy_timeout=5000`)
	if err != nil {
		return RwConn{}, errorutil.Wrap(err)
	}

	return RwConn{writer}, nil
}

func Open(filename string, poolSize int) (pair *PooledPair, err error) {
	writer, err := OpenRW(filename)
	if err != nil {
		return nil, errorutil.Wrap(err)
	}

	defer func() {
		if err != nil {
			errorutil.UpdateErrorFromCloser(writer, &err)
		}
	}()

	pool, err := OpenRO(filename, poolSize)
	if err != nil {
		return nil, errorutil.Wrap(err)
	}

	return &PooledPair{RwConn: writer, RoConnPool: pool, Closers: closers.New(newConnCloser(filename, RWMode, writer.DB), pool), Filename: filename}, nil
}
