package convert

import (
	"database/sql"
	"fmt"
	"strconv"

	"goat/pkg/logger"

	"go.uber.org/zap"
)

const sqliteAllTablesQuery = `SELECT * FROM sqlite_master WHERE type='table';`

type DBSource interface {
	Open() error
	ReadOneRow() (map[string]string, error)
	TotalCount() int64
	Headers() []string
	Close() error
}

type sqliteDBSource struct {
	dbpath      string
	db          *sql.DB
	table       string
	resultCount int
	resultRows  *sql.Rows
	resultCols  []string
}

// GetTotal implements DBSource
func (s *sqliteDBSource) TotalCount() int64 {
	return int64(s.resultCount)
}

// Headers implements DBSource
func (s *sqliteDBSource) Headers() []string {
	return s.resultCols
}

// Close implements DBSource
func (s *sqliteDBSource) Close() error {
	return s.db.Close()
}

func (s *sqliteDBSource) getTotalRowsCount() error {
	if rows, err := s.db.Query(fmt.Sprintf("SELECT COUNT(*) FROM %s;", s.table)); err != nil {
		logger.Logger.Error("failed to query table", zap.Error(err))
		return err
	} else {
		defer rows.Close()
		cols, err := rows.Columns()
		if err != nil {
			panic(err)
		}

		results := make([]interface{}, len(cols))
		for i := range results {
			results[i] = new(interface{})
		}
		for rows.Next() {
			if err := rows.Scan(results[:]...); err != nil {
				panic(err)
			}
			cur := make([]string, len(cols))
			for i := range results {
				val := *results[i].(*interface{})
				var str string
				if val == nil {
					str = "NULL"
				} else {
					switch v := val.(type) {
					case []byte:
						str = string(v)
					default:
						str = fmt.Sprintf("%v", v)
					}
				}
				cur[i] = str
			}
			if len(cur) != 1 {
				panic("unexpected result")
			} else {
				if res, err := strconv.Atoi(cur[0]); err != nil {
					panic(err)
				} else {
					s.resultCount = res
				}
			}
		}
	}
	return nil
}

func (s *sqliteDBSource) Open() error {
	if db, err := sql.Open("sqlite3", s.dbpath); err != nil {
		panic(err)
	} else {
		s.db = db
	}

	if rows, err := s.db.Query(sqliteAllTablesQuery); err != nil {
		logger.Logger.Error("failed to query sqlite tables", zap.Error(err))
		return err
	} else {
		defer rows.Close()
		foundTable := false
		cols, err := rows.Columns()
		if err != nil {
			panic(err)
		}
		logger.Logger.Debug("result columns", zap.Strings("cols", cols))
		results := make([]interface{}, len(cols))
		for i := range results {
			results[i] = new(interface{})
		}
		for rows.Next() {
			if err := rows.Scan(results[:]...); err != nil {
				panic(err)
			}
			cur := make([]string, len(cols))
			for i := range results {
				val := *results[i].(*interface{})
				var str string
				if val == nil {
					str = "NULL"
				} else {
					switch v := val.(type) {
					case []byte:
						str = string(v)
					default:
						str = fmt.Sprintf("%v", v)
					}
				}
				if cols[i] == "tbl_name" {
					s.table = str
					foundTable = true
				}
				cur[i] = str
			}
			logger.Logger.Debug("found one table details", zap.Any("table", cur))
		}
		if !foundTable {
			return fmt.Errorf("table not found")
		}
	}

	s.getTotalRowsCount()

	if s.resultCount == 0 {
		return fmt.Errorf("no rows in table")
	}

	if rows2, err := s.db.Query(fmt.Sprintf("SELECT * FROM %s;", s.table)); err != nil {
		logger.Logger.Error("failed to query table", zap.Error(err))
		return err
	} else {
		s.resultRows = rows2
	}

	return nil
}

// ReadOneRow implements DBSource
func (s *sqliteDBSource) ReadOneRow() (map[string]string, error) {
	if s.resultRows == nil {
		return nil, fmt.Errorf("no results")
	}
	if s.resultCols == nil {
		cols, err := s.resultRows.Columns()
		if err != nil {
			panic(err)
		}
		s.resultCols = cols
	}
	results := make([]interface{}, len(s.resultCols))
	for i := range results {
		results[i] = new(interface{})
	}

	if s.resultRows.Next() {
		if err := s.resultRows.Scan(results[:]...); err != nil {
			panic(err)
		}
		cur := make([]string, len(s.resultCols))
		for i := range results {
			val := *results[i].(*interface{})
			var str string
			if val == nil {
				str = "NULL"
			} else {
				switch v := val.(type) {
				case []byte:
					str = string(v)
				default:
					str = fmt.Sprintf("%v", v)
				}
			}
			cur[i] = str
		}
		// logger.Logger.Debug("found one row", zap.Any("row", cur))

		row := make(map[string]string)
		for i := range cur {
			row[s.resultCols[i]] = cur[i]
		}
		return row, nil
	} else {
		s.resultRows.Close()
		return nil, nil
	}
}

func NewDBSource(dbpath string, dbtype string) DBSource {
	if dbtype == "sqlite" {
		return &sqliteDBSource{
			dbpath:      dbpath,
			resultCount: 0,
			resultRows:  nil,
			resultCols:  nil,
		}
	} else {
		panic("unimplemented")
	}
}
