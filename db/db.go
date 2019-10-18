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
	"github.com/k1LoW/harvest/version"
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
CREATE VIRTUAL TABLE logs USING FTS4(
  host,
  path,
  target_id INTEGER,
  ts,
  ts_unixnano INTEGER,
  ts_year INTEGER,
  ts_month INTEGER,
  ts_day INTEGER,
  ts_hour INTEGER,
  ts_minute INTEGER,
  ts_second INTEGER,
  ts_time_zone,
  filled_by_prev_ts INTEGER,
  content
);
CREATE TABLE metas (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  key TEXT NOT NULL,
  value TEXT NOT NULL,
  UNIQUE(key)
);
`,
	)

	tags := map[string]int64{}
	for tag := range c.Tags() {
		res, err := db.Exec(`INSERT INTO tags (name) VALUES ($1);`, tag)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		id, err := res.LastInsertId()
		if err != nil {
			return nil, errors.WithStack(err)
		}
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

	d := &DB{
		ctx:     ctx,
		db:      db,
		logger:  l,
		logChan: make(chan parser.Log),
	}

	err = d.SetMeta("harvest.version", version.Version)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	err = d.SetMeta("db.initialized_at", time.Now().Format(time.RFC3339))
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return d, nil
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
		ts := log.Timestamp
		if ts == nil {
			ts = &time.Time{}
		}

		_, err := d.db.Exec(`
INSERT INTO logs (
  host,
  path,
  ts,
  ts_unixnano,
  ts_year,
  ts_month,
  ts_day,
  ts_hour,
  ts_minute,
  ts_second,
  ts_time_zone,
  target_id,
  filled_by_prev_ts,
  content
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14);`,
			log.Host,
			log.Path,
			ts,
			ts.UnixNano(),
			ts.Year(),
			ts.Month(),
			ts.Day(),
			ts.Hour(),
			ts.Minute(),
			ts.Second(),
			ts.Format("-0700"),
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
		/* #nosec */
		rows, err := d.db.Queryx(fmt.Sprintf(`
SELECT
  logs.host,
  logs.path,
  logs.ts_unixnano,
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
ORDER BY logs.ts_unixnano, logs.rowid ASC;`, cond))
		if err != nil {
			d.logger.Error("DB error", zap.String("error", err.Error()))
			return
		}
		for rows.Next() {
			err := rows.StructScan(&log)
			// restore log.Timestamp from log.TimestampUnixNano
			if log.TimestampUnixNano < 0 {
				log.Timestamp = nil
			} else {
				ts := time.Unix(0, log.TimestampUnixNano)
				log.Timestamp = &ts
			}
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
	query := fmt.Sprintf("SELECT (length(%s)) AS length from logs GROUP BY %s ORDER by length DESC LIMIT 1;", strings.Join(colName, ")+length("), strings.Join(colName, ",")) // #nosec
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
	return l.Length, nil
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

type resultTargetName struct {
	Target string `db:"target"`
}

func (d *DB) Count(groups []string, matches []string) ([][]string, error) {
	targetGroup := []string{}
	tagGroup := []string{}
	tsGroupBy := []string{}
	tsColmun := "ts"
	for _, g := range groups {
		switch g {
		case "year":
			tsColmun = `strftime("%Y-01-01 00:00:00", datetime(ts), "localtime")`
			tsGroupBy = []string{"year"}
		case "month":
			tsColmun = `strftime("%Y-%m-01 00:00:00", datetime(ts), "localtime")`
			tsGroupBy = []string{"ts_year", "ts_month"}
		case "day":
			tsColmun = `strftime("%Y-%m-%d 00:00:00", datetime(ts), "localtime")`
			tsGroupBy = []string{"ts_year", "ts_month", "ts_day"}
		case "hour":
			tsColmun = `strftime("%Y-%m-%d %H:00:00", datetime(ts), "localtime")`
			tsGroupBy = []string{"ts_year", "ts_month", "ts_day", "ts_hour"}
		case "minute":
			tsColmun = `strftime("%Y-%m-%d %H:%M:00", datetime(ts), "localtime")`
			tsGroupBy = []string{"ts_year", "ts_month", "ts_day", "ts_hour", "ts_minute"}
		case "second":
			tsColmun = `strftime("%Y-%m-%d %H:%M:%S", datetime(ts), "localtime")`
			tsGroupBy = []string{"ts_year", "ts_month", "ts_day", "ts_hour", "ts_minute", "ts_second"}
		case "description":
			targetGroup = append(targetGroup, "t.description")
		case "host":
			targetGroup = append(targetGroup, "t.host")
		case "target":
			targetGroup = append(targetGroup, "t.source")
		default:
			tagGroup = append(tagGroup, g)
		}
	}

	header := []string{}
	columns := []string{}
	if len(tsGroupBy) > 0 {
		header = []string{"ts"}
		columns = []string{tsColmun}
	}

	switch {
	case len(targetGroup) > 0:
		columnNameQuery := strings.Join(targetGroup, `||"/"||`)
		groupByQuery := strings.Join(targetGroup, ",")
		query := fmt.Sprintf("SELECT %s AS target FROM logs AS l LEFT JOIN targets AS t ON l.target_id = t.id GROUP BY %s;", columnNameQuery, groupByQuery) // #nosec
		tn := []resultTargetName{}
		err := d.db.Select(&tn, query)
		if err != nil {
			return nil, err
		}
		for _, n := range tn {
			if len(tagGroup) > 0 {
				for _, tag := range tagGroup {
					if len(matches) > 0 {
						for _, m := range matches {
							columnName := strings.Join([]string{n.Target, tag, m}, "/")
							header = append(header, columnName)
							columns = append(columns, fmt.Sprintf(`SUM(CASE WHEN %s = "%s" AND l.target_id IN (SELECT tt.target_id FROM tags AS t LEFT JOIN targets_tags AS tt ON t.id = tt.tag_id WHERE t.name = "%s") AND l.content LIKE "%%%s%%" THEN 1 ELSE 0 END) AS "%s"`, columnNameQuery, n.Target, tag, m, columnName)) // #nosec
						}
					} else {
						columnName := strings.Join([]string{n.Target, tag}, "/")
						header = append(header, columnName)
						columns = append(columns, fmt.Sprintf(`SUM(CASE WHEN %s = "%s" AND l.target_id IN (SELECT tt.target_id FROM tags AS t LEFT JOIN targets_tags AS tt ON t.id = tt.tag_id WHERE t.name = "%s") THEN 1 ELSE 0 END) AS "%s"`, columnNameQuery, n.Target, tag, columnName)) // #nosec
					}
				}
			} else {
				if len(matches) > 0 {
					for _, m := range matches {
						columnName := strings.Join([]string{n.Target, m}, "/")
						header = append(header, columnName)
						columns = append(columns, fmt.Sprintf(`SUM(CASE WHEN %s = "%s" AND l.content LIKE "%%%s%%" THEN 1 ELSE 0 END) AS "%s"`, columnNameQuery, n.Target, m, columnName)) // #nosec
					}
				} else {
					header = append(header, n.Target)
					columns = append(columns, fmt.Sprintf(`SUM(CASE WHEN %s = "%s" THEN 1 ELSE 0 END) AS "%s"`, columnNameQuery, n.Target, n.Target)) // #nosec
				}
			}
		}
	case len(targetGroup) == 0 && len(tagGroup) > 0:
		for _, tag := range tagGroup {
			if len(matches) > 0 {
				for _, m := range matches {
					columnName := strings.Join([]string{tag, m}, "/")
					header = append(header, columnName)
					columns = append(columns, fmt.Sprintf(`SUM(CASE WHEN l.target_id IN (SELECT tt.target_id FROM tags AS t LEFT JOIN targets_tags AS tt ON t.id = tt.tag_id WHERE t.name = "%s") AND l.content LIKE "%%%s%%" THEN 1 ELSE 0 END) AS "%s"`, tag, m, columnName)) // #nosec
				}
			} else {
				header = append(header, tag)
				columns = append(columns, fmt.Sprintf(`SUM(CASE WHEN l.target_id IN (SELECT tt.target_id FROM tags AS t LEFT JOIN targets_tags AS tt ON t.id = tt.tag_id WHERE t.name = "%s") THEN 1 ELSE 0 END) AS "%s"`, tag, tag)) // #nosec
			}
		}
	case len(targetGroup) == 0 && len(tagGroup) == 0 && len(matches) > 0:
		for _, m := range matches {
			header = append(header, m)
			columns = append(columns, fmt.Sprintf(`SUM(CASE WHEN l.content LIKE "%%%s%%" THEN 1 ELSE 0 END) AS "%s"`, m, m)) // #nosec
		}
	case len(targetGroup) == 0 && len(tagGroup) == 0 && len(matches) == 0:
		header = append(header, "count")
		columns = append(columns, "COUNT(*)")
	}

	var query string
	if len(tsGroupBy) > 0 {
		query = fmt.Sprintf(`SELECT %s FROM logs AS l LEFT JOIN targets AS t ON l.target_id = t.id GROUP BY %s ORDER BY ts;`, strings.Join(columns, ", "), strings.Join(tsGroupBy, ", ")) // #nosec
	} else {
		query = fmt.Sprintf(`SELECT %s FROM logs AS l LEFT JOIN targets AS t ON l.target_id = t.id;`, strings.Join(columns, ", ")) // #nosec
	}

	d.logger.Debug(fmt.Sprintf("Execute query: %s", query))

	rows, err := d.db.Query(query)
	if err != nil {
		return nil, err
	}

	results := [][]string{header}

	for rows.Next() {
		rowColumns := make([]string, len(columns))
		columnPointers := make([]interface{}, len(columns))
		for i := range rowColumns {
			columnPointers[i] = &rowColumns[i]
		}
		err := rows.Scan(columnPointers...)
		if err != nil {
			return nil, err
		}
		results = append(results, rowColumns)
	}

	return results, nil
}

func (d *DB) SetMeta(key string, value string) error {
	_, err := d.db.Exec(`INSERT INTO metas (key, value) VALUES ($1, $2);`, key, value)
	if err != nil {
		return errors.WithStack(err)
	}
	return nil
}
