package configuration

import (
	"encoding/json"
	"fmt"
	"github.com/Sirupsen/logrus"
	"github.com/julienschmidt/httprouter"
	"net/http"
)

var (
	ListenAddr = ":8080"
	consulAddr = "127.0.0.1:8500"
)

func StartHttpServer(){
	router := httprouter.New()
	router.Handle("GET", "/key", GetKey)
	router.Handle("GET", "/keylist", GetList)

	router.Handle("POST", "/put", SetKey)
	router.Handle("PUT", "/put", SetKey)
	router.Handle("DELETE", "/key", DeleteKey)
	logrus.Infof("Listen...... %s",ListenAddr)
	err := http.ListenAndServe(":8080", router)
	if err != nil{
		logrus.Error(err)
	}
}


func RespondV1(w http.ResponseWriter, code int, data interface{}) {
	var response []byte
	var err error
	var isJSON bool

	if code == 200 {
		switch data.(type) {
		case string:
			response = []byte(data.(string))
		case []byte:
			response = data.([]byte)
		case nil:
			response = []byte{}
		default:
			isJSON = true
			response, err = json.Marshal(data)
			if err != nil {
				code = 500
				data = err
			}
		}
	}

	if code != 200 {
		isJSON = true
		response = []byte(fmt.Sprintf(`{"message":"%s"}`, data))
	}

	if isJSON {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
	}
	w.Header().Set("X-GOCONFIG-Content-Type", "go-config; version=1.0")
	w.WriteHeader(code)
	w.Write(response)
}

func GetKey(w http.ResponseWriter, req *http.Request, ps httprouter.Params) {
	key := req.URL.Query().Get("key")
	logrus.Info(key)
	if key == "" {
		RespondV1(w, 500, "parameter error")
	}
	confs := NewConsul(consulAddr)
	value, _, err := confs.Get(key)
	if err != nil{
		RespondV1(w, 500, fmt.Sprintf("consul error %s",err.Error()))
	}
	RespondV1(w, 200,value)
	//RespondV1(w, 200,"ok")

}

func GetList(w http.ResponseWriter, req *http.Request, ps httprouter.Params) {

}

func SetKey(w http.ResponseWriter, req *http.Request, ps httprouter.Params) {
	key := req.PostFormValue("key")
	logrus.Info("key: ",key)
	value := req.PostFormValue("value")
	logrus.Info("value: ",value)
	if key == "" {
		RespondV1(w, 500, "parameter error")
	}

	confs := NewConsul(consulAddr)
	err := confs.Put(key, []byte(value))
	if err != nil{
		RespondV1(w, 500, fmt.Sprintf("consul error %s",err.Error()))
	}
	result := map[string][]byte{
		key: []byte(value),
	}

	RespondV1(w, 200,result)
}

func DeleteKey(w http.ResponseWriter, req *http.Request, ps httprouter.Params) {

}
