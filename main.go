package main

import (
	"encoding/json"
	"github.com/stels-cs/stat-proxy/internal"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var defaultLogger *log.Logger

func init() {
	rand.Seed(time.Now().UnixNano())
	defaultLogger = log.New(os.Stdout, "", log.Ldate|log.Ltime|log.Lshortfile|log.LUTC)
}

func GetConfigMap() (map[string]string, error) {
	configuration := make(map[string]string)
	filename := "config.json"
	file, err := os.Open(filename)
	if err != nil {  return configuration,err }
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&configuration)
	if err != nil {  return configuration,err }
	return configuration,nil
}

func env(name string, def string) string {
	value := os.Getenv(name)
	if value == "" {
		config, err := GetConfigMap()
		if err != nil {
			return def
		}
		value, has := config[name]
		if has {
			return value
		} else {
			return def
		}
		//return ff(v != "", v, def)
	} else {
		return value
	}
}

func main() {

	services := internal.GetServicePoll(defaultLogger)
	core := internal.CreateCoreStatistic()

	appName := env("APP", "dev_log_saver/0")

	if env("UDP", "") != "" {
		udpServer := internal.CreateUpdServer(core, env("UDP", ""), defaultLogger)
		services.Push(udpServer)
	}


	if env("HTTP", "") != "" {
		if env("POSTGRES", "") == "" {
			defaultLogger.Fatal("Cant start http without POSTGRES env")
		}
		saver := internal.CreateStatSaver(defaultLogger, env("POSTGRES", ""))
		services.Push(saver)
		httpServer := internal.CreateHttpServer(env("HTTP", ""), env("SECRET", "secret"), defaultLogger, saver)
		services.Push(httpServer)
	}

	if env("PROXY_TO", "") != "" {
		proxy := internal.CreateProxySender(core, env("PROXY_TO", ""), defaultLogger)
		services.Push(proxy)
	}

	if services.Count() == 0 {
		defaultLogger.Fatal("No service to run")
	}

	if services.Count() > 0 {
		err := internal.LogSum( env("LOG_ADDRESS", "127.0.0.1:1007"), appName, "start", 1 )
		if err != nil {
			defaultLogger.Println("Err log:",err)
		}
		services.RunAll()
	} else {
		err := internal.LogSum( env("LOG_ADDRESS", "127.0.0.1:1007"), appName, "fail_start", 1 )
		if err != nil {
			defaultLogger.Println("Err log:",err)
		}
		defaultLogger.Println("No service to start")
		return
	}

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT)
	signal.Notify(signalChan, syscall.SIGUSR1)
	sig := <-signalChan
	defaultLogger.Println(sig.String())
	defaultLogger.Println("Stopping...")
	<-services.StopAll()
	defaultLogger.Println("Done!")
}