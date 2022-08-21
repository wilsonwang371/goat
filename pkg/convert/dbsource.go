package convert

import "database/sql"

type DBSource interface {
	ReadOneRow() (map[string]string, error)
	Eof() bool
}

type sqliteDBSource struct {
	db *sql.DB
}

// Eof implements DBSource
func (*sqliteDBSource) Eof() bool {
	panic("unimplemented")
}

// ReadOneRow implements DBSource
func (*sqliteDBSource) ReadOneRow() (map[string]string, error) {
	panic("unimplemented")
}

func NewDBSource(dbfile string, dbtype string) DBSource {
	if dbtype == "sqlite3" {
		if db, err := sql.Open(dbtype, dbfile); err != nil {
			panic(err)
		} else {
			return &sqliteDBSource{
				db: db,
			}
		}
	} else {
		panic("unimplemented")
	}
}
