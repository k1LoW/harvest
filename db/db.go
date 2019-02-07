package db

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"time"

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
		`CREATE VIRTUAL TABLE log USING FTS4(host, path, tag, ts INTEGER, filled_by_prev INTEGER, content);`,
	)
	l.Info("DB initialized.")

	db.MustExec("PRAGMA journal_mode = MEMORY")
	db.MustExec("PRAGMA synchronous = NORMAL")

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

	db.MustExec("PRAGMA journal_mode = MEMORY")
	db.MustExec("PRAGMA synchronous = NORMAL")

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
	count := 0

	ticker := time.NewTicker(time.Duration(10) * time.Second)
	go func() {
		for {
			select {
			case <-ticker.C:
				d.logger.Info(fmt.Sprintf("%d logs are fetched. ", count))
			}
		}
	}()

L:
	for log := range d.logChan {
		_, err := d.db.NamedExec("INSERT INTO log (host, path, tag, ts, filled_by_prev, content) VALUES (:host, :path, :tag, :ts, :filled_by_prev, :content)", &log)
		if err != nil {
			d.logger.Error("DB error", zap.Error(err))
			break L
		}
		count++
		select {
		case <-d.ctx.Done():
			d.logger.Info(fmt.Sprintf("%d logs are fetched.", count))
			break L
		default:
		}
	}
}

// Cat ...
func (d *DB) Cat(cond string) chan parser.Log {
	go func() {
		defer close(d.logChan)
		log := parser.Log{}
		rows, err := d.db.Queryx(fmt.Sprintf("SELECT * FROM log %s ORDER BY ts, rowid ASC;", cond))
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

// ResultLength ...
type ResultLength struct {
	Length int `db:"length"`
}

// GetColumnMaxLength ...
func (d *DB) GetColumnMaxLength(colName ...string) (int, error) {
	query := fmt.Sprintf("SELECT (length(%s)) AS length from log GROUP BY %s ORDER by length DESC LIMIT 1;", strings.Join(colName, ")+length("), strings.Join(colName, ","))
	l := ResultLength{}
	err := d.db.Get(&l, query)
	if err != nil {
		return 0, err
	}
	return l.Length, nil
}
