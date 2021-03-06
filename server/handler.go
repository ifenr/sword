package server

import (
	"encoding/json"
	"fmt"
	"github.com/Moekr/sword/common"
	"io/ioutil"
	"net/http"
	"sort"
	"strconv"
)

func httpConf(w http.ResponseWriter, r *http.Request) {
	if !checkToken(w, r) {
		return
	}
	if bs, err := json.Marshal(conf.Targets); err != nil {
		http.Error(w, "marshal conf error: "+err.Error(), http.StatusInternalServerError)
	} else {
		w.Header().Set("Content-Type", "application/json")
		w.Write(bs)
	}
}

func httpData(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "http method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if !checkToken(w, r) {
		return
	}
	targetId, err := parseIntParam(r, "t", false, -1)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	observerId, err := parseIntParam(r, "o", false, -1)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	bs, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "read post data error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	record := &common.Record{}
	if err := json.Unmarshal(bs, record); err != nil {
		http.Error(w, "unmarshal port data error: "+err.Error(), http.StatusBadRequest)
		return
	}
	if dataSets := dataSets[targetId]; dataSets != nil {
		if dataSet := dataSets[observerId]; dataSet != nil {
			dataSet.Put(record)
			return
		}
	}
	http.Error(w, "no such target id or observer id", http.StatusBadRequest)
}

func checkToken(w http.ResponseWriter, r *http.Request) bool {
	if r.Header.Get(common.TokenHeaderName) != _args.Token {
		http.Error(w, "token invalid", http.StatusForbidden)
		return false
	}
	return true
}

func httpAbbrData(w http.ResponseWriter, r *http.Request) {
	targetId, err := parseIntParam(r, "t", false, -1)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	timeRange, err := parseIntParam(r, "r", true, 1)
	if err != nil {
		timeRange = rangeDay
	}
	data := dataSets[targetId]
	result := make([]*AbbrDataSet, 0, len(data))
	for _, dataSet := range data {
		result = append(result, dataSet.GetAbbrData(timeRange))
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Observer.Id < result[j].Observer.Id
	})
	if bs, err := json.Marshal(result); err != nil {
		http.Error(w, "marshal result error: "+err.Error(), http.StatusInternalServerError)
	} else {
		w.Header().Set("Content-Type", "application/json")
		w.Write(bs)
	}
}

func httpFullData(w http.ResponseWriter, r *http.Request) {
	targetId, err := parseIntParam(r, "t", false, -1)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	observerId, err := parseIntParam(r, "o", false, -1)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	timeRange, err := parseIntParam(r, "r", true, 1)
	if err != nil {
		timeRange = rangeDay
	}
	if dataSets := dataSets[targetId]; dataSets != nil {
		if dataSet := dataSets[observerId]; dataSet != nil {
			data := dataSet.GetFullData(timeRange)
			if bs, err := json.Marshal(data); err != nil {
				http.Error(w, "marshal result error: "+err.Error(), http.StatusInternalServerError)
			} else {
				w.Header().Set("Content-Type", "application/json")
				w.Write(bs)
			}
			return
		}
	}
	http.Error(w, "no such target id or observer id", http.StatusBadRequest)
}

func httpStatData(w http.ResponseWriter, r *http.Request) {
	if !checkToken(w, r) {
		return
	}
	targetId, err := parseIntParam(r, "t", false, -1)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	interval, _ := parseIntParam(r, "i", true, 10)
	if dataSets := dataSets[targetId]; dataSets != nil {
		var target *common.Target
		result := make([]*StatDataSet, 0, len(dataSets))
		for _, dataSet := range dataSets {
			target = dataSet.Target
			result = append(result, dataSet.GetStatData(int(interval)))
		}
		sort.Slice(result, func(i, j int) bool {
			return result[i].Observer.Id < result[j].Observer.Id
		})
		data := map[string]interface{}{
			"target": target,
			"data"  : result,
		}
		if bs, err := json.Marshal(data); err != nil {
			http.Error(w, "marshal result error: "+err.Error(), http.StatusInternalServerError)
		} else {
			w.Header().Set("Content-Type", "application/json")
			w.Write(bs)
		}
		return
	}
	http.Error(w, "no such target id or observer id", http.StatusBadRequest)
}

var errMissingParam = fmt.Errorf("missing required param")

func parseIntParam(r *http.Request, key string, nullable bool, defaultValue int64) (int64, error) {
	r.ParseForm()
	str := r.Form.Get(key)
	if str == "" {
		if nullable {
			return defaultValue, nil
		} else {
			return defaultValue, errMissingParam
		}
	}
	val, err := strconv.ParseInt(str, 10, 64)
	if err != nil {
		return defaultValue, err
	}
	return val, nil
}
