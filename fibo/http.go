package fibo

import (
	"encoding/json"
	"github.com/julienschmidt/httprouter"
	"github.com/rolandhe/fibo/logger"
	"net/http"
	"reflect"
	"strconv"
	"time"
	"unsafe"
)

const (
	okStatus  = 200
	errStatus = 500
)

const (
	httpLogNone        = 0
	httpLogWithoutBody = 1
	httpLogWithBody    = 2
)

type ResultOne struct {
	Code       int    `json:"code"`
	ErrMessage string `json:"errMessage"`
	Id         int64  `json:"id"`
}

type ResultBatch struct {
	Code       int    `json:"code"`
	ErrMessage string `json:"errMessage"`
	BatchIds   []*BatchIds
}

func HttpService(gen *Generator) func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
	return func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
		logLevel := gen.conf.logLevel
		var startTime int64
		url := request.URL.String()
		if logLevel > httpLogNone {
			startTime = time.Now().UnixNano()
			logger.GLogger.Infof("enter %s", url)
		}
		ns := params.ByName("nameSpace")
		if ns == "" || ns == "/" {
			ns = DefaultNamespace
		}
		var ret ResultOne
		id, err := gen.GenOneId(ns)
		if err != nil {
			ret.Code = errStatus
			ret.ErrMessage = err.Error()
		} else {
			ret.Id = id
			ret.Code = okStatus
		}
		jsonValue, _ := json.Marshal(&ret)

		outputResult(writer, jsonValue)
		loggerFinal(logLevel, url, startTime, jsonValue)
	}
}

func HttpBatchService(gen *Generator) func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
	return func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
		logLevel := gen.conf.logLevel
		var startTime int64
		url := request.URL.String()
		if logLevel > httpLogNone {
			startTime = time.Now().UnixNano()
			logger.GLogger.Infof("enter %s", url)
		}

		ns := params.ByName("nameSpace")
		if ns == "" || ns == "/" {
			ns = DefaultNamespace
		}
		queryValues := request.URL.Query()
		batchVale := queryValues["batch"]
		if len(batchVale) == 0 {
			loggerFinal(logLevel, url, startTime, outputBatchResult(writer, buildBatchErr("no batch query value")))
			return
		}

		batch, err := strconv.Atoi(batchVale[0])
		if err != nil {
			loggerFinal(logLevel, url, startTime, outputBatchResult(writer, buildBatchErr("invalid batch query value")))
			return
		}
		batchIds, err := gen.GenBatchId(ns, int64(batch))
		if err != nil {
			loggerFinal(logLevel, url, startTime, outputBatchResult(writer, buildBatchErr(err.Error())))
			return
		}
		r := &ResultBatch{
			Code:     okStatus,
			BatchIds: batchIds,
		}
		loggerFinal(logLevel, url, startTime, outputBatchResult(writer, r))
	}
}

func loggerFinal(logLevel int, url string, startTime int64, jsonValue []byte) {
	if logLevel == httpLogNone {
		return
	}
	cost := time.Now().UnixNano() - startTime
	body := ""
	if logLevel == httpLogWithBody {
		body = attachBytesString(jsonValue)
	}
	logger.GLogger.Infof("leave %s,cost:%d ns,body:%s", url, cost, body)
}

func outputBatchResult(writer http.ResponseWriter, r *ResultBatch) []byte {
	jsonValue, _ := json.Marshal(r)
	outputResult(writer, jsonValue)
	return jsonValue
}

func outputResult(writer http.ResponseWriter, jsonValue []byte) {
	writer.Header().Set("Content-Type", "application/jsonValue")
	writer.Write(jsonValue)
}

func buildBatchErr(errMsg string) *ResultBatch {
	return &ResultBatch{
		Code:       errStatus,
		ErrMessage: errMsg,
	}
}

func attachBytesString(b []byte) string {
	sl := (*reflect.SliceHeader)(unsafe.Pointer(&b))
	sh := &reflect.StringHeader{Data: sl.Data, Len: sl.Len}

	return *(*string)(unsafe.Pointer(sh))
}
