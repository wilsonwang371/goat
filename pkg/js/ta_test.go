package js

import (
	"fmt"
	"reflect"
	"testing"

	talib "github.com/wilsonwang371/go-talib"
)

func TestTALibMethods(t *testing.T) {
	ta := talib.NewTALib()
	r := reflect.TypeOf(ta)
	for i := 0; i < r.NumMethod(); i++ {
		m := r.Method(i)
		fmt.Println(m.Name)
	}
}
