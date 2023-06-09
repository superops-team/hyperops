package http

import (
	"fmt"
	"os"

	"github.com/buaazp/fasthttprouter"
	"github.com/valyala/fasthttp"
)

func Index(ctx *fasthttp.RequestCtx) {
	fmt.Fprint(ctx, "OK!\n")
}

func APIregist(r *fasthttprouter.Router) {
	r.GET("/", Index)
}

func Serve(addr string) {
	router := fasthttprouter.New()
	APIregist(router)

	p := NewPrometheus("hyperops")
	fastpHandler := p.WrapHandler(router)

	if err := fasthttp.ListenAndServe(addr, fastpHandler); err != nil {
		fasthttp.ListenAndServeUNIX("/var/run/hyperops.sock", os.FileMode(int(0755)), fastpHandler)
	}
}
