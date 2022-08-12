package js

import (
	"reflect"

	talib "github.com/wilsonwang371/go-talib"
)

var TALibMethods []reflect.Method

func AllTALibMethods() []reflect.Method {
	ta := talib.NewTALib()
	r := reflect.TypeOf(ta)
	methods := make([]reflect.Method, 0)
	for i := 0; i < r.NumMethod(); i++ {
		if r.Method(i).Name != "" {
			methods = append(methods, r.Method(i))
		}
	}
	return methods
}
