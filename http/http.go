package http

import (
	"encoding/json"
	"log"
	"net/http"
	_ "net/http/pprof"

	"github.com/ZeaLoVe/query/g"
)

type Dto struct {
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

func Start() {
	if !g.Config().Http.Enabled {
		log.Println("http.Start warning, not enabled")
		return
	}

	// config http routes
	configCommonRoutes()
	configProcHttpRoutes()
	configGraphRoutes()
	configApiRoutes()
	configGrafanaRoutes()

	// start http server
	addr := g.Config().Http.Listen
	s := &http.Server{
		Addr:           addr,
		MaxHeaderBytes: 1 << 30,
	}

	log.Println("http.Start ok, listening on", addr)
	log.Fatalln(s.ListenAndServe())
}

func RenderJson(w http.ResponseWriter, v interface{}) {
	bs, err := json.Marshal(v)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.Header().Set("Access-Control-Allow-Headers", "Authorization,DNT,X-Mx-ReqToken,Keep-Alive,User-Agen,x-requested-with,Content-Type,Content-Length")
	w.Header().Set("Access-Control-Allow-Methods", "POST,GET,OPTIONS,PUT")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	w.Write(bs)
}

func RenderDataJson(w http.ResponseWriter, data interface{}) {
	RenderJson(w, Dto{Msg: "success", Data: data})
}

func RenderMsgJson(w http.ResponseWriter, msg string) {
	RenderJson(w, map[string]string{"msg": msg})
}

func AutoRender(w http.ResponseWriter, data interface{}, err error) {
	if err != nil {
		RenderMsgJson(w, err.Error())
		return
	}
	RenderDataJson(w, data)
}

func StdRender(w http.ResponseWriter, data interface{}, err error) {
	if err != nil {
		w.WriteHeader(400)
		RenderMsgJson(w, err.Error())
		return
	}
	RenderJson(w, data)
}
