package main

import (
	"fmt"
	"github.com/buaazp/fasthttprouter"
	"github.com/valyala/fasthttp"
	"log"
)

func main() {
	router := fasthttprouter.New()

	router.POST("/order", Order)

	if err := fasthttp.ListenAndServe(":80", router.Handler); err != nil {
		log.Println(err)
	}
}

func Order(ctx *fasthttp.RequestCtx) {
	fmt.Println(ctx.PostBody())
}
