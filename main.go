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
var records map[string]string
var lock sync.Mutex

func init() {
	port = flag.String("http", "8080", "proxy listen addr")
	configFile = flag.String("config", "./config.json", "config file")
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
	return fmt.Sprintf("http://127.0.0.1:%s/", port)
}

func ReDirect(w http.ResponseWriter, r *http.Request) {

	target := GetTarget(r.URL.Path)

	client := &http.Client{}
	req, _ := http.NewRequest(r.Method, target + r.URL.Path, r.Body)

	res, err := client.Do(req)
	if err != nil {
		glog.Errorf("failed to redirect request: %s", r.URL.Path)
		w.Write([]byte("failed to redirect request."))
		return
	}

	defer res.Body.Close()
	body, _ := ioutil.ReadAll(res.Body)
	w.Write(body)
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

	LoadConfig(*configFile)

	http.HandleFunc("/", ReDirect)

	if err := http.ListenAndServe("0.0.0.0:" + *port, nil); err != nil {
		glog.Error("failed to start proxy server.")
	}
}
