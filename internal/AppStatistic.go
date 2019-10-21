package internal

import (
	"github.com/axiomhq/hyperloglog"
	lru "github.com/hashicorp/golang-lru"
	"log"
	"sync"
)

const MaxMetricCount = 70
const PatternSize = 100

type AppStatistic struct {
	metrics  map[string]int
	patterns map[string]*lru.Cache
	hll      map[string]*hyperloglog.Sketch
	hllDay   map[string]*hyperloglog.Sketch
	name     string
	overload bool
	mutex    sync.Mutex
}

func CreateAppStatistic(name string) *AppStatistic {
	return &AppStatistic{
		name:     name,
		metrics:  make(map[string]int),
		patterns: make(map[string]*lru.Cache),
		hll:      make(map[string]*hyperloglog.Sketch),
		hllDay:   make(map[string]*hyperloglog.Sketch),
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
	if len(app.hll) > MaxMetricCount {
		app.overload = true
	}
	if len(app.hllDay) > MaxMetricCount {
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
	for metric, hll := range app.hll {
		result[metric] = int(hll.Estimate())
	}
	app.hll = make(map[string]*hyperloglog.Sketch)
	app.metrics = make(map[string]int)
	app.overload = false
	return &result
}

func (app *AppStatistic) TakeIntDayMetrics() *map[string]int {
	app.mutex.Lock()
	defer app.mutex.Unlock()
	result := make(map[string]int)
	for metric, hll := range app.hllDay {
		result[metric] = int(hll.Estimate())
	}
	app.hllDay = make(map[string]*hyperloglog.Sketch)
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
					} else if value, ok := valueRaw.([2]int); ok {
						buff[pattern] = value[0] / value[1]
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

func (app *AppStatistic) StrSum(name string, value int, pattern string) {
	if app.overload {
		return
	}
	app.mutex.Lock()
	if cache, has := app.patterns[name]; has {
		oldValueI, has := cache.Peek(pattern)
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

func (app *AppStatistic) StrSet(name string, value int, pattern string) {
	if app.overload {
		return
	}
	app.mutex.Lock()
	if cache, has := app.patterns[name]; has {
		cache.Add(pattern, value)
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

func (app *AppStatistic) StrMin(name string, value int, pattern string) {
	if app.overload {
		return
	}
	app.mutex.Lock()
	if cache, has := app.patterns[name]; has {
		oldValueI, has := cache.Peek(pattern)
		if has {
			oldValue, ok := oldValueI.(int)
			if ok && oldValue < value {
				cache.Add(pattern, oldValue)
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

func (app *AppStatistic) StrMax(name string, value int, pattern string) {
	if app.overload {
		return
	}
	app.mutex.Lock()
	if cache, has := app.patterns[name]; has {
		oldValueI, has := cache.Peek(pattern)
		if has {
			oldValue, ok := oldValueI.(int)
			if ok && oldValue > value {
				cache.Add(pattern, oldValue)
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

func (app *AppStatistic) StrAvg(name string, value int, pattern string) {
	if app.overload {
		return
	}
	app.mutex.Lock()
	if cache, has := app.patterns[name]; has {
		oldValueI, has := cache.Peek(pattern)
		if has {
			oldValue, ok := oldValueI.([2]int)
			if ok {
				sum := oldValue[0] + value
				count := oldValue[1] + 1
				cache.Add(pattern, [2]int{sum, count})
			} else {
				cache.Add(pattern, [2]int{value, 1})
			}
		} else {
			cache.Add(pattern, [2]int{value, 1})
		}
	} else {
		cache, err := lru.New(PatternSize)
		if err != nil {
			log.Println("Fail create pattern cache for app", app.name, err)
		} else {
			cache.Add(pattern, [2]int{value, 1})
			app.patterns[name] = cache
		}
	}
	app.overloadCheck()
	app.mutex.Unlock()
}

func (app *AppStatistic) Hll(name, pattern string) {
	if app.overload {
		return
	}
	app.mutex.Lock()
	if hll, has := app.hll[name]; has {
		hll.Insert([]byte(pattern))
	} else {
		hll := hyperloglog.New16()
		hll.Insert([]byte(pattern))
		app.hll[name] = hll
	}
	app.overloadCheck()
	app.mutex.Unlock()
}

func (app *AppStatistic) HllDay(name, pattern string) {
	if app.overload {
		return
	}
	app.mutex.Lock()
	if hll, has := app.hllDay[name]; has {
		hll.Insert([]byte(pattern))
	} else {
		hll := hyperloglog.New16()
		hll.Insert([]byte(pattern))
		app.hllDay[name] = hll
	}
	app.overloadCheck()
	app.mutex.Unlock()
}

func (app *AppStatistic) GetData() map[string][]byte {
	app.mutex.Lock()
	defer app.mutex.Unlock()
	buff := make(map[string][]byte)
	for key, hll := range app.hllDay {
		tmp, err := hll.MarshalBinary()
		if err == nil {
			buff[key] = tmp
		} else {
			log.Println("Fail marshal data", key, err)
		}
	}
	return buff
}

func (app *AppStatistic) RestoreData(res map[string][]byte) {
	app.mutex.Lock()
	defer app.mutex.Unlock()
	for key, data := range res {
		h := hyperloglog.New16()
		err := h.UnmarshalBinary(data)
		if err == nil {
			app.hllDay[key] = h
		} else {
			log.Println("Fail unmarshal data", key, err)
		}
	}
}
