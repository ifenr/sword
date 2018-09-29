package server

import (
	"encoding/json"
	"github.com/Moekr/sword/common"
	"github.com/Moekr/sword/util"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

var (
	args     *util.Args
	conf     *common.Conf
	dataSets map[int64]map[int64]*DataSet
)

func Start(serverArgs *util.Args) error {
	args = serverArgs
	if err := loadConf(); err != nil {
		return err
	}
	dataSets = make(map[int64]map[int64]*DataSet, len(conf.Targets))
	for _, target := range conf.Targets {
		dataSets[target.Id] = make(map[int64]*DataSet, len(conf.Observers))
	}
	loadData()
	go deferKill()
	defer saveData()
	http.HandleFunc("/api/conf", httpConf)
	http.HandleFunc("/api/data", httpData)
	http.HandleFunc("/api/data/abbr", httpAbbrData)
	http.HandleFunc("/api/data/full", httpFullData)
	http.HandleFunc("/", httpIndex)
	return http.ListenAndServe(args.Bind, nil)
}

func loadConf() error {
	if bs, err := ioutil.ReadFile(args.ConfFile); err != nil {
		return err
	} else {
		return json.Unmarshal(bs, &conf)
	}
}

func deferKill() {
	ch := make(chan os.Signal)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)
	log.Printf("receive signal %v\n", <-ch)
	saveData()
	os.Exit(0)
}
