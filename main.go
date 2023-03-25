package main

import (
	"github.com/julienschmidt/httprouter"
	"github.com/rolandhe/fibo/fibo"
	"github.com/rolandhe/fibo/logger"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"net/http"
	"os"
	"strconv"
	"time"
)

func main() {
	g := fibo.NewGenerator()
	err := fibo.InitWorkerId(g)
	if err != nil {
		logger.GLogger.Infoln(err)
		return
	}

	h2s := &http2.Server{}

	router := httprouter.New()

	router.GET("/fibo/one/*nameSpace", fibo.HttpService(g))
	router.GET("/fibo/one", fibo.HttpService(g))
	router.GET("/fibo/batch/*nameSpace", fibo.HttpBatchService(g))
	router.GET("/fibo/batch", fibo.HttpBatchService(g))
	router.GET("/fibo/heath", func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
		writer.WriteHeader(200)
	})
	appInfo := fibo.GetAppInfo()
	server := &http.Server{
		Addr:        "0.0.0.0:" + strconv.Itoa(appInfo.Port),
		Handler:     h2c.NewHandler(router, h2s),
		IdleTimeout: time.Minute * 30,
	}

	logger.GLogger.Infof("App %s Listening [0.0.0.0:%d]...\n", appInfo.Name, appInfo.Port)
	checkErr(server.ListenAndServe(), "while listening")
}

func checkErr(err error, msg string) {
	if err == nil {
		return
	}
	logger.GLogger.Errorf("ERROR: %s: %s\n", msg, err)
	os.Exit(1)
}
