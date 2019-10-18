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
func (core *CoreStatistic) Str(appName, param string, value int, pattern string) {
	core.GetApp(appName).Str(param, value, pattern)
}

func (core *CoreStatistic) TakeIntMetrics() *map[string]*map[string]int {
	result := make(map[string]*map[string]int)
	buff := core.apps
	for appName, app := range buff {
		result[appName] = app.TakeIntMetrics()
	}
	return &result
}

func (core *CoreStatistic) TakeStringMetrics() *map[string]*map[string]map[string]int {
	result := make(map[string]*map[string]map[string]int)
	buff := core.apps
	core.mutex.Lock()
	core.apps = make(map[string]*AppStatistic)
	core.mutex.Unlock()
	for appName, app := range buff {
		m := app.TakeStringMetrics()
		if m != nil && len(*m) > 0 {
			result[appName] = m
		}
	}
	return &result
}
