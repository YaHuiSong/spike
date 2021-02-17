package spikeutil

import (
	"encoding/json"
	"log"
	"net/http"
)

//ResponseData the struct of the response data
type ResponseData struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data,omitempty"`
}

// RespJson returns the json data to client
func RespJson(w http.ResponseWriter, code int, msg string, data interface{}) {
	Resp(w, code, msg, data)
}

//Resp returns the json data from seperate information
func Resp(w http.ResponseWriter, code int, msg string, data interface{}) {
	//设置header 为JSON 默认是text/html,所以特别指出返回的数据类型为application/json
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	resp := ResponseData{
		Code: code,
		Msg:  msg,
		Data: data,
	}
	ret, err := json.Marshal(resp)
	if err != nil {
		log.Panicln(err.Error())
	}

	w.Write(ret)
}
