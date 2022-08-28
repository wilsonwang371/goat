package profile

import (
	"fmt"
	"net/http"
	_ "net/http/pprof"

	"goat/pkg/consts"
)

func StartProfilingServer() {
	go func() {
		http.ListenAndServe(fmt.Sprintf("127.0.0.1:%d", consts.ProfilePort), nil)
	}()
}
