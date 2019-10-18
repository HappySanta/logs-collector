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
	proxy.timer = time.NewTicker(time.Duration(proxy.saveTimeSec) * time.Second)
	proxy.stopCh = make(chan bool, 1)
	tick := 0
	for {
		select {
		case <-proxy.timer.C:
		case <-proxy.stopCh:
		}
		if proxy.stop {
			proxy.timer.Stop()
			return nil
		}
		go proxy.sendInt()
		tick++
		if tick >= 5 {
			tick = 0
			go proxy.sendString()
		}
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

func (proxy *ProxySender) sendInt() {
	data := proxy.core.TakeIntMetrics()
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

func (proxy *ProxySender) sendString() {
	data := proxy.core.TakeStringMetrics()
	if data == nil || len(*data) <= 0 {
		return
	}
	raw, err := json.Marshal(data)
	if err != nil {
		proxy.logger.Println("Fail marshal data: ", err)
		return
	}
	tr := http.Client{Timeout: time.Second * 600}

	req, err := http.NewRequest("POST", proxy.url, bytes.NewReader(raw))
	if err != nil {
		proxy.logger.Println("Creating request error: ", err)
		return
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set(StringHeader, "1")

	resp, err := tr.Do(req)
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
