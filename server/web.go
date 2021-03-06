package server

import (
	"github.com/Moekr/sword/common"
	"github.com/Moekr/sword/util/logs"
	"html/template"
	"net/http"
)

var (
	htmlTemplate = template.New("template")
)

func init() {
	htmlTemplate.Funcs(map[string]interface{}{
		"dict": func(values ...interface{}) map[string]interface{} {
			if len(values)%2 != 0 {
				return nil
			}
			dict := make(map[string]interface{}, len(values)/2)
			for i := 0; i < len(values); i += 2 {
				key, ok := values[i].(string)
				if !ok {
					return nil
				}
				dict[key] = values[i+1]
			}
			return dict
		},
	})
	template.Must(htmlTemplate.Parse(HeadTemplate))
	template.Must(htmlTemplate.Parse(CategoryTemplate))
	template.Must(htmlTemplate.Parse(HeaderTemplate))
	template.Must(htmlTemplate.Parse(FooterTemplate))
	template.Must(htmlTemplate.Parse(IndexTemplate))
	template.Must(htmlTemplate.Parse(DetailTemplate))
}

func httpIndex(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	if path != "/" && path != "/index.html" {
		http.Redirect(w, r, "/index.html", http.StatusMovedPermanently)
		return
	}
	cid, err := parseIntParam(r, "c", true, conf.DefaultCid)
	if err != nil {
		cid = conf.DefaultCid
	}
	timeRange, err := parseIntParam(r, "r", true, 1)
	if err != nil {
		timeRange = rangeDay
	}
	category := conf.GetCategory(cid)
	if category == nil {
		http.Redirect(w, r, "/index.html", http.StatusMovedPermanently)
		return
	}
	params := map[string]interface{}{
		"categories": conf.Categories,
		"category":   category,
		"targets":    conf.GetTargets(cid),
		"timeRange":  timeRange,
	}
	if err := htmlTemplate.ExecuteTemplate(w, "index", params); err != nil {
		logs.Error("parse index template error: %s", err.Error())
	}
}

func httpDetail(w http.ResponseWriter, r *http.Request) {
	targetId, err := parseIntParam(r, "t", false, -1)
	if err != nil {
		http.Redirect(w, r, "/index.html", http.StatusMovedPermanently)
		return
	}
	var target *common.Target
	for _, tar := range conf.Targets {
		if tar.Id == targetId {
			target = tar
			break
		}
	}
	if target == nil {
		http.Redirect(w, r, "/index.html", http.StatusMovedPermanently)
		return
	}
	timeRange, err := parseIntParam(r, "r", true, 1)
	if err != nil {
		timeRange = rangeDay
	}
	params := map[string]interface{}{
		"categories": conf.Categories,
		"target":     target,
		"observers":  conf.Observers,
		"timeRange":  timeRange,
	}
	if err := htmlTemplate.ExecuteTemplate(w, "detail", params); err != nil {
		logs.Error("parse detail template error: %s", err.Error())
	}
}

func httpCSS(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/css")
	w.Write([]byte(IndexCSS))
}

func httpJS(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/javascript")
	w.Write([]byte(IndexJS))
}

func httpFavicon(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "image/x-icon")
	w.Write(FaviconData)
}
