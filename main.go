package main

import (
	"encoding/json"
	"github.com/stels-cs/stat-proxy/internal"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"strconv"
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

	internal.InitCache(defaultLogger, env("ACCESS_TOKEN", ""))

	if env("UDP", "") != "" {
		udpServer := internal.CreateUpdServer(core, env("UDP", ""), defaultLogger)
		services.Push(udpServer)
	}

	sum := func(name string, value int) {
		err := internal.LogSum(env("LOG_ADDRESS", "127.0.0.1:1007"), appName, name, value)
		if err != nil {
			defaultLogger.Println("Err sum:", err)
		}
	}

	if env("HTTP", "") != "" {
		if env("POSTGRES", "") == "" {
			defaultLogger.Fatal("Cant start http without POSTGRES env")
		}
		saver := internal.CreateStatSaver(defaultLogger, env("POSTGRES", ""), sum)
		services.Push(saver)
		httpServer := internal.CreateHttpServer(env("HTTP", ""), env("SECRET", "secret"), defaultLogger, saver)
		services.Push(httpServer)
	}

	if env("PROXY_TO", "") != "" {
		saveTime, err := strconv.Atoi(env("SAVE_TIME", "60"))
		if err != nil || saveTime <= 1 {
			defaultLogger.Println("Bad save time, used default: 60", env("SAVE_TIME", "60"))
			saveTime = 60
		}
		proxy := internal.CreateProxySender(core, env("PROXY_TO", ""), env("TMP_FILE", "./data.tmp"), defaultLogger, saveTime)
		services.Push(proxy)
	}

	if services.Count() == 0 {
		defaultLogger.Fatal("No service to run")
	}

	if services.Count() > 0 {
		services.RunAll()
		go func() {
			t := time.NewTimer(5 * time.Second)
			<-t.C
			err := internal.LogSum(env("LOG_ADDRESS", "127.0.0.1:1007"), appName, "start", 1)
			if err != nil {
				defaultLogger.Println("Err log:", err)
			}

			//group := internal.Get([]string{"1"})
			//if g,has:=group["1"]; len(group) != 1 || !has {
			//	defaultLogger.Println("Cant fetch group")
			//} else {
			//	defaultLogger.Println("Group: 1", g.Name)
			//}
			//
			//
			//err = internal.LogStrSum(env("LOG_ADDRESS", "127.0.0.1:1007"), appName, "sum", 2, "5")
			//if err != nil {
			//	defaultLogger.Println("Err log:", err)
			//}
			//err = internal.LogStrSum(env("LOG_ADDRESS", "127.0.0.1:1007"), appName, "sum", 3, "5")
			//if err != nil {
			//	defaultLogger.Println("Err log:", err)
			//}
			//err = internal.LogStrSum(env("LOG_ADDRESS", "127.0.0.1:1007"), appName, "sum", 2, "2")
			//if err != nil {
			//	defaultLogger.Println("Err log:", err)
			//}
			//
			//err = internal.LogStrMin(env("LOG_ADDRESS", "127.0.0.1:1007"), appName, "min", 5, "5")
			//if err != nil {
			//	defaultLogger.Println("Err log:", err)
			//}
			//err = internal.LogStrMin(env("LOG_ADDRESS", "127.0.0.1:1007"), appName, "min", 10, "5")
			//if err != nil {
			//	defaultLogger.Println("Err log:", err)
			//}
			//
			//err = internal.LogStrMax(env("LOG_ADDRESS", "127.0.0.1:1007"), appName, "max", 10, "10")
			//if err != nil {
			//	defaultLogger.Println("Err log:", err)
			//}
			//err = internal.LogStrMax(env("LOG_ADDRESS", "127.0.0.1:1007"), appName, "max", 5, "10")
			//if err != nil {
			//	defaultLogger.Println("Err log:", err)
			//}
			//
			//err = internal.LogStrSet(env("LOG_ADDRESS", "127.0.0.1:1007"), appName, "set", 9, "10")
			//if err != nil {
			//	defaultLogger.Println("Err log:", err)
			//}
			//err = internal.LogStrSet(env("LOG_ADDRESS", "127.0.0.1:1007"), appName, "set", 10, "10")
			//if err != nil {
			//	defaultLogger.Println("Err log:", err)
			//}
			//
			//err = internal.LogStrAvg(env("LOG_ADDRESS", "127.0.0.1:1007"), appName, "avg", 15, "10")
			//if err != nil {
			//	defaultLogger.Println("Err log:", err)
			//}
			//err = internal.LogStrAvg(env("LOG_ADDRESS", "127.0.0.1:1007"), appName, "avg", 5, "10")
			//if err != nil {
			//	defaultLogger.Println("Err log:", err)
			//}
			//err = internal.LogHll(env("LOG_ADDRESS", "127.0.0.1:1007"), appName, "hll", "10")
			//if err != nil {
			//	defaultLogger.Println("Err log:", err)
			//}
			//err = internal.LogHll(env("LOG_ADDRESS", "127.0.0.1:1007"), appName, "hll", "20")
			//if err != nil {
			//	defaultLogger.Println("Err log:", err)
			//}
			//err = internal.LogHll(env("LOG_ADDRESS", "127.0.0.1:1007"), appName, "hll", "10")
			//if err != nil {
			//	defaultLogger.Println("Err log:", err)
			//}
			//err = internal.LogHllDay(env("LOG_ADDRESS", "127.0.0.1:1007"), appName, "hllday", "15")
			//if err != nil {
			//	defaultLogger.Println("Err log:", err)
			//}

		}()

	} else {
		err := internal.LogSum(env("LOG_ADDRESS", "127.0.0.1:1007"), appName, "fail_start", 1)
		if err != nil {
			defaultLogger.Println("Err log:", err)
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
