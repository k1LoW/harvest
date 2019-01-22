package db

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/jmoiron/sqlx"
	"github.com/k1LoW/harvest/parser"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

// DB ...
type DB struct {
	ctx     context.Context
	db      *sqlx.DB
	logChan chan parser.Log
	logger  *zap.Logger
}

// NewDB ...
func NewDB(ctx context.Context, l *zap.Logger, dbPath string) (*DB, error) {
	fullPath, err := filepath.Abs(dbPath)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	l.Info(fmt.Sprintf("Create %s.", fullPath))
	db, err := sqlx.Connect("sqlite3", dbPath)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	db.MustExec(
		`CREATE VIRTUAL TABLE log USING FTS4(host, path, ts INTEGER, content);`,
	)
	l.Info("DB initialized.")

	return &DB{
		ctx:     ctx,
		db:      db,
		logger:  l,
		logChan: make(chan parser.Log),
	}, nil
}

// AttachDB ...
func AttachDB(ctx context.Context, l *zap.Logger, dbPath string) (*DB, error) {
	fullPath, err := filepath.Abs(dbPath)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	db, err := sqlx.Connect("sqlite3", fullPath)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return &DB{
		ctx:     ctx,
		db:      db,
		logger:  l,
		logChan: make(chan parser.Log),
	}, nil
}

// In ...
func (d *DB) In() chan parser.Log {
	return d.logChan
}

// StartInsert ...
func (d *DB) StartInsert() {
	defer close(d.logChan)
L:
	for log := range d.logChan {
		_, err := d.db.NamedExec("INSERT INTO log (host, path, ts, content) VALUES (:host, :path, :ts, :content)", &log)
		if err != nil {
			d.logger.Error("DB error", zap.Error(err))
			break L
		}
		select {
		case <-d.ctx.Done():
			break L
		default:
		}
	}
}

// Cat ...
func (d *DB) Cat() chan parser.Log {
	go func() {
		defer close(d.logChan)
		log := parser.Log{}
		rows, err := d.db.Queryx("SELECT * FROM log ORDER BY ts ASC;")
		if err != nil {
			d.logger.Error("DB error", zap.Error(err))
			return
		}
		for rows.Next() {
			err := rows.StructScan(&log)
			if err != nil {
				d.logger.Error("DB error", zap.Error(err))
				break
			}
			d.logChan <- log
		}
	}()

	return d.logChan
}
