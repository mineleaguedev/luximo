package main

import (
	"github.com/fasthttp/router"
	"github.com/valyala/fasthttp"
	"log"
)

func main() {
	r := router.New()
	log.Fatal(fasthttp.ListenAndServe(":8080", r.Handler))
}
