package db

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/k1LoW/harvest/config"
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
func NewDB(ctx context.Context, l *zap.Logger, c *config.Config, dbPath string) (*DB, error) {
	fullPath, err := filepath.Abs(dbPath)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	l.Info(fmt.Sprintf("Create %s", fullPath))
	db, err := sqlx.Connect("sqlite3", dbPath)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	db.MustExec(
		`
CREATE TABLE targets (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  source TEXT NOT NULL,
  description TEXT,
  type TEXT NOT NULL,
  regexp TEXT,
  multi_line INTEGER,
  time_format TEXT,
  time_zone TEXT,
  scheme TEXT NOT NULL,
  host TEXT,
  user TEXT,
  port INTEGER,
  path TEXT NOT NULL
);
CREATE TABLE tags (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  name TEXT NOT NULL,
  UNIQUE(name)
);
CREATE TABLE targets_tags (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  target_id INTEGER NOT NULL,
  tag_id INTEGER NOT NULL,
  UNIQUE(target_id, tag_id)
);
CREATE VIRTUAL TABLE logs USING FTS4(host, path, target_id INTEGER, ts INTEGER, filled_by_prev_ts INTEGER, content);
`,
	)

	tags := map[string]int64{}
	for tag, _ := range c.Tags() {
		res, err := db.Exec(`INSERT INTO tags (name) VALUES ($1);`, tag)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		id, err := res.LastInsertId()
		tags[tag] = id
	}

	for _, t := range c.Targets {
		res, err := db.NamedExec(`
INSERT INTO targets (
  source,
  description,
  type,
  regexp,
  multi_line,
  time_format,
  time_zone,
  scheme,
  host,
  user,
  port,
  path
) VALUES (
  :source,
  :description,
  :type,
  :regexp,
  :multi_line,
  :time_format,
  :time_zone,
  :scheme,
  :host,
  :user,
  :port,
  :path
);`, t)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		id, err := res.LastInsertId()
		if err != nil {
			return nil, errors.WithStack(err)
		}
		t.Id = id
		for _, tag := range t.Tags {
			_, err := db.Exec(`INSERT INTO targets_tags (target_id, tag_id) VALUES ($1, $2);`, id, tags[tag])
			if err != nil {
				return nil, errors.WithStack(err)
			}
		}
	}
	l.Info("DB initialized")

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
			<-ticker.C
			d.logger.Info(fmt.Sprintf("%d log data are fetched", count))
		}
	}()

L:
	for log := range d.logChan {
		_, err := d.db.Exec(`
INSERT INTO logs (
  host,
  path,
  ts,
  target_id,
  filled_by_prev_ts,
  content
) VALUES ($1, $2, $3, $4, $5, $6);`,
			log.Host,
			log.Path,
			log.Timestamp,
			log.Target.Id,
			log.FilledByPrevTs,
			log.Content,
		)
		if err != nil {
			d.logger.Error("DB error", zap.String("error", err.Error()))
			break L
		}
		count++
		select {
		case <-d.ctx.Done():
			break L
		default:
		}
	}
	d.logger.Info(fmt.Sprintf("%d log data are fetched", count))
}

// Cat ...
func (d *DB) Cat(cond string) chan parser.Log {
	tt, err := d.GetTargetIdAndTags()
	if err != nil {
		d.logger.Error("DB error", zap.String("error", err.Error()))
		close(d.logChan)
		return d.logChan
	}
	go func() {
		defer close(d.logChan)
		log := parser.Log{}
		rows, err := d.db.Queryx(fmt.Sprintf(`
SELECT
  logs.host,
  logs.path,
  logs.ts,
  logs.filled_by_prev_ts,
  logs.content,
  targets.id AS "target.id",
  targets.source AS "target.source",
	targets.description AS "target.description",
	targets.type AS "target.type",
	targets.regexp AS "target.regexp",
	targets.multi_line AS "target.multi_line",
	targets.time_format AS "target.time_format",
	targets.time_zone AS "target.time_zone",
	targets.scheme AS "target.scheme",
	targets.host AS "target.host",
	targets.user AS "target.user",
	targets.port AS "target.port",
	targets.path AS "target.path"
FROM logs LEFT JOIN targets ON logs.target_id = targets.id
%s
ORDER BY logs.ts, logs.rowid ASC;`, cond))
		if err != nil {
			d.logger.Error("DB error", zap.String("error", err.Error()))
			return
		}
		for rows.Next() {
			err := rows.StructScan(&log)
			if err != nil {
				d.logger.Error("DB error", zap.String("error", err.Error()))
				break
			}
			log.Target.Tags = tt[log.Target.Id]
			d.logChan <- log
		}
	}()

	return d.logChan
}

// resultHost ...
type resultHost struct {
	Host string `db:"host"`
}

// GetHosts ...
func (d *DB) GetHosts() ([]string, error) {
	query := "SELECT host FROM logs GROUP BY host ORDER BY host;"
	hosts := []string{}
	r := []resultHost{}
	err := d.db.Select(&r, query)
	if err != nil {
		return []string{}, err
	}
	for _, h := range r {
		hosts = append(hosts, h.Host)
	}
	return hosts, nil
}

// resultTag ...
type resultTag struct {
	Tag string `db:"name"`
}

// GetTags ...
func (d *DB) GetTags() ([]string, error) {
	query := "SELECT name FROM tags GROUP BY name;"
	tags := []string{}
	r := []resultTag{}
	err := d.db.Select(&r, query)
	if err != nil {
		return []string{}, err
	}
	for _, t := range r {
		tags = append(tags, t.Tag)
	}
	return tags, nil
}

type resultLength struct {
	Length int `db:"length"`
}

// GetColumnMaxLength ...
func (d *DB) GetColumnMaxLength(colName ...string) (int, error) {
	query := fmt.Sprintf("SELECT (length(%s)) AS length from logs GROUP BY %s ORDER by length DESC LIMIT 1;", strings.Join(colName, ")+length("), strings.Join(colName, ","))
	l := resultLength{}
	err := d.db.Get(&l, query)
	if err != nil {
		return 0, err
	}
	return l.Length, nil
}

func (d *DB) GetTagMaxLength() (int, error) {
	query := "SELECT length(GROUP_CONCAT(tags.name, ' ')) AS length FROM targets_tags LEFT JOIN tags ON targets_tags.tag_id = tags.id WHERE target_id IN (SELECT target_id from logs GROUP BY target_id) GROUP BY targets_tags.target_id ORDER BY length DESC LIMIT 1;"
	l := resultLength{}
	err := d.db.Get(&l, query)
	if err != nil {
		return 0, err
	}
	return l.Length + 2, nil // add `[` and `]`
}

type resultTargetIdAndTags struct {
	TargetId int64  `db:"target_id"`
	Tags     string `db:"tags"`
}

func (d *DB) GetTargetIdAndTags() (map[int64][]string, error) {
	tt := []resultTargetIdAndTags{}
	query := "SELECT targets_tags.target_id AS target_id, GROUP_CONCAT(tags.name,', ') AS tags FROM tags LEFT JOIN targets_tags ON tags.id = targets_tags.tag_id GROUP BY targets_tags.target_id;"
	err := d.db.Select(&tt, query)
	if err != nil {
		return map[int64][]string{}, err
	}
	targets := map[int64][]string{}
	for _, t := range tt {
		targets[t.TargetId] = strings.Split(t.Tags, ", ")
	}
	return targets, nil
}
