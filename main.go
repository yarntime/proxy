package main

import (
	"flag"
	"github.com/golang/glog"
	"net/http"
	"io/ioutil"
	"sync"
	"strings"
	io "io/ioutil"
	"encoding/json"
	"fmt"
)

var port *string
var configFile *string
var rootDir    *string
var records map[string]string
var lock sync.Mutex

func init() {
	port = flag.String("http", "8080", "proxy listen addr")
	configFile = flag.String("config", "./config.json", "config file")
	rootDir = flag.String("rootDir", "./", "root dir")
	flag.Set("alsologtostderr", "true")
	flag.Parse()
}

func GetTarget(path string) string {
	lock.Lock()
	defer lock.Unlock()
	for k, v := range records {
		if strings.HasPrefix(path[1:], k) {
			return v
		}
	}
	return ""
}

func ReDirect(w http.ResponseWriter, r *http.Request) {

	target := GetTarget(r.URL.Path)

	if target != "" {
		client := &http.Client{}
		req, _ := http.NewRequest(r.Method, target + r.URL.Path, r.Body)
		for k, v := range r.Header {
			req.Header[k] = v
		}
		res, err := client.Do(req)
		if err != nil {
			msg := fmt.Sprintf("failed to redirect request: %s", r.URL)
			glog.Errorf(msg)
			w.Write([]byte(msg))
			return
		}

		defer res.Body.Close()
		body, _ := ioutil.ReadAll(res.Body)
		w.Write(body)
	} else {
		path := string(http.Dir(*rootDir + "/" + r.URL.Path))
		http.ServeFile(w, r, path)
	}

}

func LoadConfig(filename string) error {
	data, err := io.ReadFile(filename)
	if err != nil {
		return err
	}

	records = make(map[string]string)
	dataJson := []byte(data)
	err = json.Unmarshal(dataJson, &records)
	if err != nil {
		return err
	}

	return nil
}

func main() {
	glog.Info("starting proxy server.")

	if err := LoadConfig(*configFile); err != nil {
		glog.Fatal("failed to load config file.")
	}

	http.HandleFunc("/", ReDirect)

	if err := http.ListenAndServe("0.0.0.0:" + *port, nil); err != nil {
		glog.Error("failed to start proxy server.")
	}
}
