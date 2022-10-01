package cmd

import (
	"os"
	"testing"

	"goat/pkg/db"
	"goat/pkg/logger"

	"go.uber.org/zap"
)

func TestMergeSimple(t *testing.T) {
	var sources []*db.DB

	sourceNames := []string{
		"../samples/data/strategy_data.dumpdb",
		"../samples/data/strategy_data.dumpdb",
	}
	for _, name := range sourceNames {
		tmp, err := db.NewSQLiteDataBase(name, false)
		if err != nil {
			logger.Logger.Error("failed to open database", zap.Error(err))
			t.Fatal("failed to open database")
		}
		sources = append(sources, tmp)
	}

	defer os.Remove("tempoutput.dumpdb")
	output, err := db.NewSQLiteDataBase("tempoutput.dumpdb", true)
	if err != nil {
		logger.Logger.Error("failed to create output database", zap.Error(err))
		t.Fatal("failed to create output database")
	}

	if err := db.MergeDBs(output, sources); err != nil {
		logger.Logger.Error("failed to merge databases", zap.Error(err))
		t.Fatal("failed to merge databases")
	}
}
