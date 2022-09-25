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

	db := NewSQLiteDataBase(file.Name(), true)
	assert.NotNil(t, db)

	db2 := NewSQLiteDataBase("/tmp/test.999.db", false)
	assert.NotNil(t, db2)
}
