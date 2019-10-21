package internal

import "sync"

type CoreStatistic struct {
	mutex sync.RWMutex
	apps  map[string]*AppStatistic
}

func CreateCoreStatistic() *CoreStatistic {
	return &CoreStatistic{
		apps:  make(map[string]*AppStatistic),
		mutex: sync.RWMutex{},
	}
}

func (core *CoreStatistic) GetApp(name string) *AppStatistic {
	core.mutex.RLock()
	if app, has := core.apps[name]; has {
		core.mutex.RUnlock()
		return app
	}
	core.mutex.RUnlock()
	newApp := CreateAppStatistic(name)
	core.mutex.Lock()
	defer core.mutex.Unlock()
	if app, has := core.apps[name]; has {
		return app
	} else {
		core.apps[name] = newApp
		return newApp
	}
}

func (core *CoreStatistic) Sum(appName, param string, value int) {
	core.GetApp(appName).Sum(param, value)
}
func (core *CoreStatistic) Set(appName, param string, value int) {
	core.GetApp(appName).Set(param, value)
}
func (core *CoreStatistic) Max(appName, param string, value int) {
	core.GetApp(appName).Max(param, value)
}
func (core *CoreStatistic) Min(appName, param string, value int) {
	core.GetApp(appName).Min(param, value)
}
func (core *CoreStatistic) Avg(appName, param string, value int) {
	core.GetApp(appName).Avg(param, value)
}
func (core *CoreStatistic) StrSum(appName, param string, value int, pattern string) {
	core.GetApp(appName).StrSum(param, value, pattern)
}
func (core *CoreStatistic) StrSet(appName, param string, value int, pattern string) {
	core.GetApp(appName).StrSet(param, value, pattern)
}
func (core *CoreStatistic) StrMin(appName, param string, value int, pattern string) {
	core.GetApp(appName).StrMin(param, value, pattern)
}
func (core *CoreStatistic) StrMax(appName, param string, value int, pattern string) {
	core.GetApp(appName).StrMax(param, value, pattern)
}
func (core *CoreStatistic) StrAvg(appName, param string, value int, pattern string) {
	core.GetApp(appName).StrAvg(param, value, pattern)
}
func (core *CoreStatistic) Hll(appName, param string, pattern string) {
	core.GetApp(appName).Hll(param, pattern)
}
func (core *CoreStatistic) HllDay(appName, param string, pattern string) {
	core.GetApp(appName).HllDay(param, pattern)
}

func (core *CoreStatistic) TakeIntMetrics() *map[string]*map[string]int {
	result := make(map[string]*map[string]int)
	buff := core.apps
	for appName, app := range buff {
		x := app.TakeIntMetrics()
		if x != nil && len(*x) > 0 {
			result[appName] = x
		}
	}
	return &result
}

func (core *CoreStatistic) TakeIntDayMetrics() *map[string]*map[string]int {
	result := make(map[string]*map[string]int)
	buff := core.apps
	core.mutex.Lock()
	core.apps = make(map[string]*AppStatistic)
	core.mutex.Unlock()
	for appName, app := range buff {
		x := app.TakeIntDayMetrics()
		if x != nil && len(*x) > 0 {
			result[appName] = x
		}
	}
	return &result
}

func (core *CoreStatistic) TakeStringMetrics() *map[string]*map[string]map[string]int {
	result := make(map[string]*map[string]map[string]int)
	buff := core.apps
	for appName, app := range buff {
		m := app.TakeStringMetrics()
		if m != nil && len(*m) > 0 {
			result[appName] = m
		}
	}
	return &result
}

func (core *CoreStatistic) GetDataToSave() map[string]map[string][]byte {
	core.mutex.Lock()
	defer core.mutex.Unlock()
	buff := make(map[string]map[string][]byte)
	for appName, app := range core.apps {
		data := app.GetData()
		if len(data) > 0 {
			buff[appName] = data
		}
	}
	return buff
}

func (core *CoreStatistic) RestoreData(data map[string]map[string][]byte) {
	for appName, raw := range data {
		core.GetApp(appName).RestoreData(raw)
	}
}
