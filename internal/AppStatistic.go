package internal

import "sync"

const MaxMetricCount = 50

type AppStatistic struct {
	metrics  map[string]int
	name     string
	overload bool
	mutex    sync.Mutex
}

func CreateAppStatistic(name string) *AppStatistic {
	return &AppStatistic{
		name:     name,
		metrics:  make(map[string]int),
		overload: false,
		mutex:    sync.Mutex{},
	}
}

func (app *AppStatistic) overloadCheck() {
	if len(app.metrics) > MaxMetricCount {
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

func (app *AppStatistic) Slice() *map[string]int {
	app.mutex.Lock()
	defer app.mutex.Unlock()
	result := make(map[string]int)
	for metric, value := range app.metrics {
		result[metric] = value
	}
	app.metrics = make(map[string]int)
	app.metrics = make(map[string]int)
	app.overload = false
	return &result
}
