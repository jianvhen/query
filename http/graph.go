package http

import (
	"encoding/json"
	"errors"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/open-falcon/common/model"
	"github.com/open-falcon/query/graph"
	"github.com/toolkits/logger"
)

type EChartsData struct {
	Timestamp []interface{}              `json:"timestamp"`
	Data      map[string]([]interface{}) `json:"data"`
}

func ParseDuration(param string) (start, end int64) {
	var now, before time.Time
	dur, err := strconv.Atoi(param[0 : len(param)-1])
	if err != nil { //can't get number
		dur = 1
	}
	now = time.Now()
	if strings.HasSuffix(param, "h") {
		//1e9 means 1 seconds in go
		before = now.Add(-time.Duration(1e9 * 60 * 60 * dur))
	} else if strings.HasSuffix(param, "d") {
		before = now.Add(-time.Duration(1e9 * 60 * 60 * 24 * dur))
	} else { //param error
		return 0, 0
	}
	return before.Unix(), now.Unix()
}

//如果请求的不同counter的采集周期不一致，或者不同的采集频率，可能出现不准确
func (this *EChartsData) GetEchartsData(datas []*model.GraphQueryResponse) {
	if this.Data == nil {
		this.Data = make(map[string]([]interface{}))
	}
	var max int
	var index int
	for i, _ := range datas {
		if max < len(datas[i].Values) {
			max = len(datas[i].Values)
			index = i
		}
	}
	counter := datas[index].Counter
	for _, val := range datas[index].Values {
		this.Timestamp = append(this.Timestamp, val.Timestamp)
	}
	for i, _ := range datas {
		counter = datas[i].Counter
		for j, val := range datas[i].Values {
			if val.Timestamp == this.Timestamp[j] {
				this.Data[counter] = append(this.Data[counter], val.Value)
			} else {
				this.Data[counter] = append(this.Data[counter], model.JsonFloat(math.NaN())) //not data
			}
		}
	}
}

type GraphAliveParam struct {
	Endpoint string `json:"endpoint"`
}

type GraphAliveResponse struct {
	Endpoint string `json:"endpoint"`
	Status   int    `json:"status"`
}

type GraphHistoryParam struct {
	Start            int                    `json:"start"`
	End              int                    `json:"end"`
	CF               string                 `json:"cf"`
	EndpointCounters []model.GraphInfoParam `json:"endpoint_counters"`
}

func configGraphRoutes() {

	// method:post
	http.HandleFunc("/graph/history", func(w http.ResponseWriter, r *http.Request) {
		var body GraphHistoryParam
		decoder := json.NewDecoder(r.Body)
		err := decoder.Decode(&body)
		if err != nil {
			StdRender(w, "", err)
			return
		}

		if len(body.EndpointCounters) == 0 {
			StdRender(w, "", errors.New("empty_payload"))
			return
		}

		data := []*model.GraphQueryResponse{}
		for _, ec := range body.EndpointCounters {
			result, err := graph.QueryOne(int64(body.Start), int64(body.End), body.CF, ec.Endpoint, ec.Counter)
			if err != nil {
				logger.Error("query one fail: %v", err)
			}
			data = append(data, result)
		}

		StdRender(w, data, nil)
	})

	// method:get
	http.HandleFunc("/graph/history/one", func(w http.ResponseWriter, r *http.Request) {
		start := r.FormValue("start")
		end := r.FormValue("end")
		cf := r.FormValue("cf")
		endpoint := r.FormValue("endpoint")
		counter := r.FormValue("counter")

		if endpoint == "" || counter == "" {
			StdRender(w, "", errors.New("empty_endpoint_counter"))
			return
		}

		if cf != "AVERAGE" && cf != "MAX" && cf != "MIN" {
			StdRender(w, "", errors.New("invalid_cf"))
			return
		}

		now := time.Now()
		start_i64, err := strconv.ParseInt(start, 10, 64)
		if err != nil {
			start_i64 = now.Unix() - 3600
		}
		end_i64, err := strconv.ParseInt(end, 10, 64)
		if err != nil {
			end_i64 = now.Unix()
		}

		result, err := graph.QueryOne(start_i64, end_i64, cf, endpoint, counter)
		logger.Trace("query one result: %v, err: %v", result, err)
		if err != nil {
			StdRender(w, "", err)
			return
		}

		StdRender(w, result, nil)
	})

	// get, info
	http.HandleFunc("/graph/info/one", func(w http.ResponseWriter, r *http.Request) {
		endpoint := r.FormValue("endpoint")
		counter := r.FormValue("counter")

		if endpoint == "" || counter == "" {
			StdRender(w, "", errors.New("empty_endpoint_counter"))
			return
		}

		result, err := graph.Info(endpoint, counter)
		logger.Trace("graph.info result: %v, err: %v", result, err)
		if err != nil {
			StdRender(w, "", err)
			return
		}

		StdRender(w, result, nil)
	})

	// post, info
	http.HandleFunc("/graph/info", func(w http.ResponseWriter, r *http.Request) {
		var body []*model.GraphInfoParam
		decoder := json.NewDecoder(r.Body)
		err := decoder.Decode(&body)
		if err != nil {
			StdRender(w, "", err)
			return
		}

		if len(body) == 0 {
			StdRender(w, "", errors.New("empty_payload"))
			return
		}

		data := []*model.GraphFullyInfo{}
		for _, param := range body {
			info, err := graph.Info(param.Endpoint, param.Counter)
			if err != nil {
				logger.Trace("graph.info fail, resp: %v, err: %v", info, err)
			} else {
				logger.Trace("graph.info result: %v, err: %v", info, err)
			}
			data = append(data, info)
		}

		StdRender(w, data, nil)
	})

	// post, last
	http.HandleFunc("/graph/last", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "OPTIONS" {
			w.Header().Set("Content-Type", "application/json; charset=UTF-8")
			w.Header().Set("Access-Control-Allow-Headers", "Authorization,DNT,X-Mx-ReqToken,Keep-Alive,User-Agen")
			w.Header().Set("Access-Control-Allow-Methods", "POST,GET,OPTIONS,PUT")
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.Header().Set("Content-Length", "0")
			w.WriteHeader(200)
			return
		}
		var body []*model.GraphLastParam
		decoder := json.NewDecoder(r.Body)
		err := decoder.Decode(&body)
		if err != nil {
			StdRender(w, "", err)
			return
		}

		if len(body) == 0 {
			StdRender(w, "", errors.New("empty_payload"))
			return
		}

		data := []*model.GraphLastResp{}
		for _, param := range body {
			last, err := graph.Last(param.Endpoint, param.Counter)
			if err != nil {
				logger.Trace("graph.last fail, resp: %v, err: %v", last, err)
				continue
			}
			data = append(data, last)
		}
		StdRender(w, data, nil)
	})

	//method:get
	http.HandleFunc("/graph/sdp/one", func(w http.ResponseWriter, r *http.Request) {
		var duration, cf, endpoint string
		var counters []string
		var echarts EChartsData
		r.ParseForm()
		for key, value := range r.Form {
			switch key {
			case "duration":
				duration = value[0]
			case "cf":
				cf = value[0]
			case "endpoint":
				endpoint = value[0]
			case "counter":
				counters = value
			}
		}

		if endpoint == "" || len(counters) == 0 {
			StdRender(w, "", errors.New("empty_endpoint_counter"))
			return
		}
		//set default cf
		if cf == "" {
			cf = "AVERAGE"
		}

		if cf != "AVERAGE" && cf != "MAX" && cf != "MIN" {
			StdRender(w, "", errors.New("invalid_cf"))
			return
		}

		start, end := ParseDuration(duration)

		data := []*model.GraphQueryResponse{}
		for _, counter := range counters {
			result, err := graph.QueryOne(start, end, cf, endpoint, string(counter))
			if err != nil {
				logger.Error("query one fail: %v", err)
			}
			data = append(data, result)
		}
		echarts.GetEchartsData(data)

		StdRender(w, echarts, nil)
	})

	// post, last
	http.HandleFunc("/graph/sdp/alive", func(w http.ResponseWriter, r *http.Request) {
		var body []*GraphAliveParam
		decoder := json.NewDecoder(r.Body)
		err := decoder.Decode(&body)
		if err != nil {
			StdRender(w, "", err)
			return
		}

		if len(body) == 0 {
			StdRender(w, "", errors.New("empty_payload"))
			return
		}

		data := []*GraphAliveResponse{}
		for _, param := range body {
			var res GraphAliveResponse
			res.Endpoint = param.Endpoint
			last, err := graph.Last(param.Endpoint, "agent.alive")
			if err != nil {
				// can't get data from graph return false
				logger.Trace("graph.last fail, resp: %v, err: %v", last, err)
				res.Status = 0
				data = append(data, &res)
				continue
			}
			if time.Now().Unix()-last.Value.Timestamp <= 120 {
				res.Status = 1
			} else {
				res.Status = 0
			}
			data = append(data, &res)
		}
		StdRender(w, data, nil)
	})

}
