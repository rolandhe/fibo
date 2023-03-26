package main

import (
	"github.com/julienschmidt/httprouter"
	"github.com/rolandhe/fibo/core"
	"github.com/rolandhe/fibo/logger"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"net/http"
	"os"
	"strconv"
	"time"
)

func main() {
	g := core.NewGenerator()
	err := core.InitWorkerId(g)
	if err != nil {
		logger.GLogger.Infoln(err)
		return
	}

	h2s := &http2.Server{}

	router := httprouter.New()

	router.GET("/core/one/*nameSpace", core.HttpService(g))
	router.GET("/core/one", core.HttpService(g))
	router.GET("/core/batch/*nameSpace", core.HttpBatchService(g))
	router.GET("/core/batch", core.HttpBatchService(g))
	router.GET("/core/heath", func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
		writer.WriteHeader(200)
	})
	appInfo := core.GetAppInfo()
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
