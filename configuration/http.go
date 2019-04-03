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

type httpServer struct {
	router http.Handler
}

func StartHttpServer() {
	router := httprouter.New()
	s := &httpServer{
		router: router,
	}
	router.Handle("GET", "/key", s.GetKey)
	router.Handle("GET", "/keylist", s.GetList)

	router.Handle("POST", "/put", s.SetKey)
	router.Handle("PUT", "/put", s.SetKey)
	router.Handle("DELETE", "/key", s.DeleteKey)
	logrus.Infof("Listen...... %s", ListenAddr)
	err := http.ListenAndServe(":8080", router)
	if err != nil {
		logrus.Error(err)
	}
}

func (s *httpServer) RespondV1(w http.ResponseWriter, code int, data interface{}) {
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

func (s *httpServer) GetKey(w http.ResponseWriter, req *http.Request, ps httprouter.Params) {
	key := req.URL.Query().Get("key")
	logrus.Info(key)
	if key == "" {
		s.RespondV1(w, 500, "parameter error")
	}
	confs := NewConsul(consulAddr)
	var value []byte
	var err error
	value, _, err = confs.Get(key)
	if err != nil {
		logrus.Errorf("consul error [GetKey] %s", err)
		resultsSlice, err := GetKeyDB(key)
		if err != nil {
			logrus.Errorf("mysql error [GetKey] %s", err)
			s.RespondV1(w, 500, fmt.Sprintf("GetKey error %s", err.Error()))
		}
		value = resultsSlice[0]["value"]
	}
	s.RespondV1(w, 200, value)
	//RespondV1(w, 200,"ok")

}

func (s *httpServer) GetList(w http.ResponseWriter, req *http.Request, ps httprouter.Params) {

}

func (s *httpServer) SetKey(w http.ResponseWriter, req *http.Request, ps httprouter.Params) {
	key := req.PostFormValue("key")
	logrus.Info("key: ", key)
	value := req.PostFormValue("value")
	logrus.Info("value: ", value)
	if key == "" {
		s.RespondV1(w, 500, "parameter error")
	}

	confs := NewConsul(consulAddr)
	var err error
	err = confs.Put(key, []byte(value))
	if err != nil {
		s.RespondV1(w, 500, fmt.Sprintf("consul error %s", err.Error()))
	}
	result := map[string][]byte{
		key: []byte(value),
	}

	s.RespondV1(w, 200, result)
}

func (s *httpServer) DeleteKey(w http.ResponseWriter, req *http.Request, ps httprouter.Params) {

}
