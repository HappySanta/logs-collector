package internal

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

type ProxySender struct {
	core        *CoreStatistic
	logger      *log.Logger
	url         string
	stop        bool
	timer       *time.Ticker
	stopCh      chan bool
	saveTimeSec int
}

func CreateProxySender(core *CoreStatistic, url string, logger *log.Logger, saveTime int) *ProxySender {
	return &ProxySender{
		core:        core,
		url:         url,
		logger:      logger,
		stop:        false,
		saveTimeSec: saveTime,
	}
}

func (proxy *ProxySender) Start() error {
	proxy.stop = false
	now := time.Now()
	x := proxy.saveTimeSec - now.Second()%proxy.saveTimeSec
	proxy.timer = time.NewTicker(time.Duration(x) * time.Second)
	reset := true
	proxy.stopCh = make(chan bool, 1)
	for {
		select {
		case <-proxy.timer.C:
		case <-proxy.stopCh:
		}
		if proxy.stop {
			proxy.timer.Stop()
			return nil
		}
		if reset {
			reset = false
			proxy.timer.Stop()
			proxy.timer = time.NewTicker(time.Duration(proxy.saveTimeSec) * time.Second)
		}
		go proxy.send()
	}
}

func (proxy *ProxySender) GetName() string {
	return "ProxyServer"
}

func (proxy *ProxySender) Stop() error {
	proxy.stop = true
	proxy.stopCh <- true
	return nil
}

func (proxy *ProxySender) send() {
	data := proxy.core.Slice()
	if data == nil || len(*data) <= 0 {
		return
	}
	raw, err := json.Marshal(data)
	if err != nil {
		proxy.logger.Println("Fail marshal data: ", err)
		return
	}
	tr := http.Client{Timeout: time.Second * 300}

	resp, err := tr.Post(proxy.url, "application/json", bytes.NewReader(raw))
	defer resp.Body.Close()
	if err != nil {
		proxy.logger.Println("Send data error: ", err)
		return
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		proxy.logger.Println("Read data error: ", err)
		return
	}
	if string(body) != "OK" {
		proxy.logger.Println("Bad response: ", string(body))
	}
}
