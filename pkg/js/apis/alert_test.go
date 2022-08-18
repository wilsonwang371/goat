package apis

import (
	"testing"

	"goat/pkg/config"

	"github.com/robertkrimen/otto"
)

func TestAlertSim(t *testing.T) {
	cfg := &config.Config{}
	cli, err := NewAlertObject(nil, nil)
	if cli != nil || err == nil {
		t.Error("expected nil, got", cli)
	}

	cli, err = NewAlertObject(cfg, nil)
	if cli != nil || err == nil {
		t.Error("expected nil, got", cli)
	}

	cli, err = NewAlertObject(cfg, otto.New())
	if err != nil {
		t.Error(err)
	}
}
