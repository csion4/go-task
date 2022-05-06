package tLogAdapter

import (
	"com.csion/tasks/tLog"
	"fmt"
)

var log = tLog.GetTLog()

// ------- tLog适配gin ------
type TLogGinAdapter struct {

}

func (t *TLogGinAdapter) Write(p []byte) (n int, err error){
	log.Customize(tLog.Info, "%d%t", string(p[: len(p)-1]))
	return n, err

}

// ------- tLog适配gorm ------
type TLogGormAdapter struct {
}

func (t *TLogGormAdapter) Printf(s string, v ...interface{}){
	log.Customize(tLog.Info, "%d%t [%l]", fmt.Sprintf(s, v...))
}