package internal

import (
	"fmt"
	"log"
	"time"
)

type Service interface {
	Start() error
	GetName() string
	Stop() error
}

type ServicePoll struct {
	logger   *log.Logger
	poll     []Service
	stop     chan bool
	allStop  bool
	stopWait int
}

func GetServicePoll(logger *log.Logger) ServicePoll {
	return ServicePoll{
		logger: logger,
		poll:   []Service{},
	}
}

func (sp *ServicePoll) Push(service Service) {
	sp.poll = append(sp.poll, service)
}

func (sp *ServicePoll) Count() int {
	return len(sp.poll)
}

func (sp *ServicePoll) RunAll() {
	sp.allStop = false
	for _, v := range sp.poll {
		sp.run(v)
	}
}

func (sp *ServicePoll) StopAll() chan bool {
	sp.allStop = true
	sp.stopWait = 0
	sp.stop = make(chan bool, 1)
	for _, v := range sp.poll {
		sp.stopWait++
		sp.logger.Println(fmt.Sprintf("[%s] stopping...", v.GetName()))
		err := v.Stop()
		if err != nil {
			sp.logger.Println(fmt.Sprintf("[%s] stopping error: %s", v.GetName(), err))
		}
	}
	return sp.stop
}

func (sp *ServicePoll) run(service Service) {
	sp.logger.Println(fmt.Sprintf("[%s] is started", service.GetName()))
	errorCount := 0
	lastEventTime := time.Now()
	go func() {
		for {
			err := service.Start()
			if sp.allStop {
				if err != nil {
					sp.logger.Println(fmt.Sprintf("[%s] %s", service.GetName(), err.Error()))
				}
				sp.logger.Println(fmt.Sprintf("[%s] stopped", service.GetName()))
				sp.onStopService()
				return
			} else if err != nil {
				sp.logger.Println(fmt.Sprintf("[%s] %s", service.GetName(), err.Error()))

				errorCount++
				if errorCount > 1000 {
					panic(service.GetName() + " generate too more errors")
				}
				if time.Now().Sub(lastEventTime) > 10*time.Second {
					errorCount = 0
				}
			} else {
				sp.logger.Println(fmt.Sprintf("[%s] Restarted", service.GetName()))
			}
		}
	}()
}

func (sp *ServicePoll) onStopService() {
	sp.stopWait--
	if sp.stopWait <= 0 {
		sp.stop <- true
	} else {
		sp.logger.Printf("Wait for %d servces", sp.stopWait)
	}
}
