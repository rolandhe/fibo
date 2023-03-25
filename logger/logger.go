package logger

import (
	"go.uber.org/zap"
	"log"
)

var GLogger *zap.SugaredLogger

func init() {
	logger, err := zap.NewDevelopment()
	if err != nil {
		log.Println(err)
	}
	GLogger = logger.Sugar()
}
