package internal

import (
	lru "github.com/hashicorp/golang-lru"
	"log"
	"sync"
)

const MaxMetricCount = 70
const PatternSize = 100

type AppStatistic struct {
	metrics  map[string]int
	name     string
	overload bool
	mutex    sync.Mutex
	patterns map[string]*lru.Cache
}

func CreateAppStatistic(name string) *AppStatistic {
	return &AppStatistic{
		name:     name,
		metrics:  make(map[string]int),
		patterns: make(map[string]*lru.Cache),
		overload: false,
		mutex:    sync.Mutex{},
	}
}

func (app *AppStatistic) overloadCheck() {
	if len(app.metrics) > MaxMetricCount {
		app.overload = true
	}
	if len(app.patterns) > MaxMetricCount {
		app.overload = true
	}
}

func (app *AppStatistic) Sum(name string, value int) {
	if app.overload {
		return
	}
	app.mutex.Lock()
	app.metrics[name] += value
	app.overloadCheck()
	app.mutex.Unlock()
}

func (app *AppStatistic) Avg(name string, value int) {
	if app.overload {
		return
	}
	app.mutex.Lock()
	app.metrics[name+"_sum"] += value
	app.metrics[name+"_count"] += 1
	app.metrics[name] = app.metrics[name+"_sum"] / app.metrics[name+"_count"]
	app.overloadCheck()
	app.mutex.Unlock()
}

func (app *AppStatistic) Set(name string, value int) {
	if app.overload {
		return
	}
	app.mutex.Lock()
	app.metrics[name] = value
	app.overloadCheck()
	app.mutex.Unlock()
}

func (app *AppStatistic) Max(name string, value int) {
	if app.overload {
		return
	}
	app.mutex.Lock()
	if _, has := app.metrics[name]; !has || value > app.metrics[name] {
		app.metrics[name] = value
	}
	app.overloadCheck()
	app.mutex.Unlock()
}

func (app *AppStatistic) Min(name string, value int) {
	if app.overload {
		return
	}
	app.mutex.Lock()
	if _, has := app.metrics[name]; !has || value < app.metrics[name] {
		app.metrics[name] = value
	}
	app.overloadCheck()
	app.mutex.Unlock()
}

func (app *AppStatistic) TakeIntMetrics() *map[string]int {
	app.mutex.Lock()
	defer app.mutex.Unlock()
	result := make(map[string]int)
	for metric, value := range app.metrics {
		result[metric] = value
	}
	app.metrics = make(map[string]int)
	app.overload = false
	return &result
}

func (app *AppStatistic) TakeStringMetrics() *map[string]map[string]int {
	app.mutex.Lock()
	defer app.mutex.Unlock()
	result := make(map[string]map[string]int)
	for metric, cache := range app.patterns {
		buff := make(map[string]int)
		keys := cache.Keys()
		for _, keyRaw := range keys {
			pattern, ok := keyRaw.(string)
			if ok {
				if valueRaw, ok := cache.Get(keyRaw); ok {
					if value, ok := valueRaw.(int); ok {
						buff[pattern] = value
					} else {
						log.Printf("Cant cast value to int metric: %s, pattern: %s", metric, pattern)
					}
				}
			} else {
				log.Printf("Cant cast pattern to string metric: %s", metric)
			}
		}
		result[metric] = buff
	}
	app.patterns = make(map[string]*lru.Cache)
	app.overload = false
	return &result
}

func (app *AppStatistic) Str(name string, value int, pattern string) {
	if app.overload {
		return
	}
	app.mutex.Lock()
	if cache, has := app.patterns[name]; has {
		oldValueI, has := cache.Get(pattern)
		if has {
			oldValue, ok := oldValueI.(int)
			if ok {
				cache.Add(pattern, value+oldValue)
			} else {
				cache.Add(pattern, value)
			}
		} else {
			cache.Add(pattern, value)
		}
	} else {
		cache, err := lru.New(PatternSize)
		if err != nil {
			log.Println("Fail create pattern cache for app", app.name, err)
		} else {
			cache.Add(pattern, value)
			app.patterns[name] = cache
		}
	}
	app.overloadCheck()
	app.mutex.Unlock()
}
