package cmd

import (
	"io/ioutil"
	"testing"

	"goat/pkg/convert"
	"goat/pkg/db"
	"goat/pkg/js"
)

func TestConvertSimple(t *testing.T) {
	rt := js.NewDBConvertRuntime(&cfg)
	script, err := ioutil.ReadFile("../samples/convert/mappings.js")
	if err != nil {
		t.Fatal("failed to read script file")
	}
	compiledScript, err := rt.Compile(string(script))
	if err != nil {
		t.Fatal("failed to compile script")
	}

	if _, err := rt.Execute(compiledScript); err != nil {
		t.Fatal("failed to execute script")
	}

	dbsource := convert.NewDBSource("../samples/data/strategy_data.sqlite", "sqlite")
	dboutput := db.NewSQLiteDataBase("../stategy_data.db", true)
	if err := rt.Convert(dbsource, dboutput); err != nil {
		t.Fatal("failed to convert data")
	}
}
