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

	db := NewSQLiteDataBase(file.Name(), false)
	assert.NotNil(t, db)
}
