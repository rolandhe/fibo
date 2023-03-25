package main

import (
	"fmt"
	"github.com/julienschmidt/httprouter"
	"github.com/rolandhe/fibo/fibo"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"log"
	"net/http"
	"os"
	"time"
)

func main() {
	g := fibo.NewGenerator()
	err := fibo.InitWorkerId(g)
	if err != nil {
		log.Println(err)
		return
	}

	h2s := &http2.Server{}

	router := httprouter.New()

	router.GET("/fibo/one/*nameSpace", fibo.HttpService(g))
	router.GET("/fibo/batch/*nameSpace", fibo.HttpBatchService(g))
	server := &http.Server{
		Addr:        "0.0.0.0:8080",
		Handler:     h2c.NewHandler(router, h2s),
		IdleTimeout: time.Minute * 30,
	}

	fmt.Printf("Listening [0.0.0.0:8080]...\n")
	checkErr(server.ListenAndServe(), "while listening")
}

func checkErr(err error, msg string) {
	if err == nil {
		return
	}
	fmt.Printf("ERROR: %s: %s\n", msg, err)
	os.Exit(1)
}
