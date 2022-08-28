package profile

import (
	"fmt"
	"net/http"
	_ "net/http/pprof"

	"goat/pkg/consts"
)

func init() {
	go func() {
		http.ListenAndServe(fmt.Sprintf(":%d", consts.ProfilePort), nil)
	}()
}
