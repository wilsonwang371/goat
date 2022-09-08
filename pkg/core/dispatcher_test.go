package core

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func TestSimpleDispatcher(t *testing.T) {
	fmt.Println("start dispatcher")
	d := NewDispatcher(context.TODO())
	go d.Run()
	time.Sleep(time.Second * 2)
	fmt.Println("stop dispatcher")
	d.Stop()
}
