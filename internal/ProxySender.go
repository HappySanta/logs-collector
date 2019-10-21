package internal

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"
)

func ReadDataFromFile(fileName string) (map[string]map[string][]byte, error) {
	configuration := make(map[string]map[string][]byte)
	file, err := os.Open(fileName)
	if err != nil {
		return configuration, err
	}
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&configuration)
	if err != nil {
		return configuration, err
	}
	return configuration, nil
}

func SaveDatToFile(fileName string, data map[string]map[string][]byte) error {
	file, err := os.Create(fileName)
	if err != nil {
		return err
	}
	encoder := json.NewEncoder(file)
	return encoder.Encode(data)
}

type ProxySender struct {
	core        *CoreStatistic
	logger      *log.Logger
	url         string
	file        string
	stop        bool
	timer       *time.Ticker
	stopCh      chan bool
	saveTimeSec int
}

func CreateProxySender(core *CoreStatistic, url, file string, logger *log.Logger, saveTime int) *ProxySender {
	return &ProxySender{
		core:        core,
		url:         url,
		logger:      logger,
		stop:        false,
		saveTimeSec: saveTime,
		file:        file,
	}
}

func (proxy *ProxySender) Start() error {
	proxy.stop = false
	proxy.timer = time.NewTicker(time.Duration(proxy.saveTimeSec) * time.Second)
	proxy.stopCh = make(chan bool, 1)
	tick := 0

	data, err := ReadDataFromFile(proxy.file)
	if err == nil {
		proxy.core.RestoreData(data)
	} else {
		proxy.logger.Println("Fail read data", proxy.file, err)
	}
	dayLog := false
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
		now := time.Now()
		if now.Hour() == 3 {
			if !dayLog {
				dayLog = true
				go proxy.sendDayInt()
			}
		} else {
			dayLog = false
		}
	}
}

func (proxy *ProxySender) GetName() string {
	return "ProxyServer"
}

func (proxy *ProxySender) Stop() error {
	proxy.stop = true
	proxy.stopCh <- true
	proxy.OnStop()
	return nil
}

func (proxy *ProxySender) OnStop() {
	proxy.sendInt()
	proxy.sendString()
	save := proxy.core.GetDataToSave()
	if len(save) > 0 {
		err := SaveDatToFile(proxy.file, save)
		if err == nil {
			proxy.logger.Println("Data saved to file", proxy.file)
		} else {
			proxy.logger.Println("FAIL saving to file", proxy.file, err)
		}
	}
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
	if err != nil {
		proxy.logger.Println("Send data error: ", err)
		return
	}
	body, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		proxy.logger.Println("Read data error: ", err)
		return
	}
	if string(body) != "OK" {
		proxy.logger.Println("Bad response: ", string(body))
	}
}

func (proxy *ProxySender) sendDayInt() {
	data := proxy.core.TakeIntDayMetrics()
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
	if err != nil {
		proxy.logger.Println("Send data error: ", err)
		return
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		proxy.logger.Println("Read data error: ", err)
		return
	}
	if string(body) != "OK" {
		proxy.logger.Println("Bad response: ", string(body))
	}
}
