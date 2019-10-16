package internal

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
)

type HttpSever struct {
	host   string
	key    string
	server *http.Server
	logger *log.Logger
	saver  *StatSaver
}

func CreateHttpServer(host, key string, logger *log.Logger, saver *StatSaver) *HttpSever {
	return &HttpSever{
		host:   host,
		key:    key,
		logger: logger,
		saver:  saver,
	}
}

func (server *HttpSever) Start() error {
	var defaultServeMux http.ServeMux
	defaultServeMux.HandleFunc("/", server.handler)
	server.server = &http.Server{Addr: server.host, Handler: &defaultServeMux}
	return server.server.ListenAndServe()
}

func (server *HttpSever) GetName() string {
	return "HttpServer"
}

func (server *HttpSever) Stop() error {
	return server.server.Close()
}

func (server *HttpSever) handler(w http.ResponseWriter, r *http.Request) {
	if strings.Contains(r.URL.Path, server.key) {
		decoder := json.NewDecoder(r.Body)
		buff := make(map[string]map[string]int)
		err := decoder.Decode(&buff)
		if err != nil {
			fmt.Fprintf(w, "Bad body")
			server.logger.Println("Bad body: ", err)
			return
		}
		fmt.Fprintf(w, "OK")
		go server.saver.Save(buff)
	} else {
		fmt.Fprintf(w, "BAD KEY")
	}
}
