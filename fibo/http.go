package fibo

import (
	"encoding/json"
	"github.com/julienschmidt/httprouter"
	"net/http"
	"strconv"
)

const (
	okStatus  = 200
	errStatus = 500
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
		json, _ := json.Marshal(&ret)
		outputResult(writer, json)
	}
}

func HttpBatchService(gen *Generator) func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
	return func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
		ns := params.ByName("nameSpace")
		if ns == "" || ns == "/" {
			ns = DefaultNamespace
		}
		queryValues := request.URL.Query()
		batchVale := queryValues["batch"]
		if len(batchVale) == 0 {
			outputBatchResult(writer, buildBatchErr("no batch query value"))
			return
		}

		batch, err := strconv.Atoi(batchVale[0])
		if err != nil {
			outputBatchResult(writer, buildBatchErr("invalid batch query value"))
			return
		}
		batchIds, err := gen.GenBatchId(ns, int64(batch))
		if err != nil {
			outputBatchResult(writer, buildBatchErr(err.Error()))
			return
		}
		r := &ResultBatch{
			Code:     okStatus,
			BatchIds: batchIds,
		}
		outputBatchResult(writer, r)
	}
}

func outputBatchResult(writer http.ResponseWriter, r *ResultBatch) {
	json, _ := json.Marshal(r)
	outputResult(writer, json)
}

func outputResult(writer http.ResponseWriter, json []byte) {
	writer.Header().Set("Content-Type", "application/json")
	writer.Write(json)
}

func buildBatchErr(errMsg string) *ResultBatch {
	return &ResultBatch{
		Code:       errStatus,
		ErrMessage: errMsg,
	}
}
