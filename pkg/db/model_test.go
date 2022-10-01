package db

import (
	"io/ioutil"
	"log"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDBOpen(t *testing.T) {
	file, err := ioutil.TempFile("/tmp", "test.*.db")
	if err != nil {
		log.Fatal(err)
	}
	defer os.Remove(file.Name())
	file.Close()

	db, err := NewSQLiteDataBase(file.Name(), true)
	if err != nil {
		t.Fatal(err)
	}
	assert.NotNil(t, db)

	os.Remove("/tmp/test.999.db")
	db2, err := NewSQLiteDataBase("/tmp/test.999.db", false)
	if err != nil {
		t.Fatal(err)
	}
	assert.NotNil(t, db2)
}
