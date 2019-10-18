package internal

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/pgxpool"
	"log"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var r *regexp.Regexp

func init() {
	r, _ = regexp.Compile("^[A-z0-9-_]+/[0-9]+$")
}

func isValidAppName(name string) bool {
	return r.MatchString(name)
}

func getIntTableName(appName string) string {
	return "t_" + strings.ReplaceAll(appName, "-", "_")
}

func getStringTableName(appName string) string {
	return "t_str_" + strings.ReplaceAll(appName, "-", "_")
}

type StatSaver struct {
	logger         *log.Logger
	databaseUrl    string
	connection     *pgxpool.Pool
	stop           chan bool
	existingTables map[string]bool
	sum            func(name string, value int)
}

func CreateStatSaver(logger *log.Logger, url string, sum func(name string, value int)) *StatSaver {
	return &StatSaver{
		logger:         logger,
		databaseUrl:    url,
		existingTables: make(map[string]bool),
		sum:            sum,
	}
}

func (saver *StatSaver) Start() error {
	config, err := pgxpool.ParseConfig(saver.databaseUrl)
	if err != nil {
		return err
	}
	config.MaxConns = 10
	config.HealthCheckPeriod = 180 * time.Second
	config.MaxConnLifetime = 10 * time.Minute

	conn, err := pgxpool.ConnectConfig(context.Background(), config)
	if err != nil {
		return err
	}
	saver.connection = conn
	saver.stop = make(chan bool, 1)
	<-saver.stop
	return nil
}

func (saver *StatSaver) Stop() error {
	saver.connection.Close()
	//err := saver.connection.Close(context.Background())
	saver.stop <- false
	return nil
}

func (saver *StatSaver) GetName() string {
	return "StatSaver"
}

func (saver *StatSaver) SaveInt(data map[string]map[string]int) {
	count := 0
	for appName, data := range data {
		if isValidAppName(appName) {
			saver.SaveAppDataInt(appName, data)
			count++
		} else {
			saver.logger.Println("Invalid app name", appName)
			saver.sum("invalid_app", 1)
		}
	}
	saver.sum("saved", 1)
}

func (saver *StatSaver) SaveAppDataInt(appName string, data map[string]int) {
	appParts := strings.Split(appName, "/")
	if len(appParts) != 2 {
		saver.logger.Println("Bad app parts", appName)
		return
	}
	nodeId, err := strconv.Atoi(appParts[1])
	if err != nil {
		saver.logger.Println("Bad node id", appName, err)
		return
	}
	table := getIntTableName(appParts[0])
	err = saver.createTableIntMetric(table)
	if err != nil {
		saver.logger.Println("Table not created", appName, err)
		return
	}
	err = saver.saveIntMetrics(table, nodeId, data)
	if err != nil {
		saver.logger.Println("Data saveIntMetrics fail", appName, err)
		return
	}
}

func (saver *StatSaver) createTableIntMetric(name string) error {
	if _, has := saver.existingTables[name]; has {
		return nil
	}
	_, err := saver.connection.Exec(context.Background(), `create table IF NOT EXISTS `+name+`
(
	created_at timestamp default now(),
	type varchar(50) not null,
	value integer not null,
	node_id integer not null
);
create index IF NOT EXISTS `+name+`_created_at_index on `+name+` (created_at desc);
`)

	if err == nil {
		saver.existingTables[name] = true
	}
	return err
}

func (saver *StatSaver) saveIntMetrics(tableName string, nodeId int, data map[string]int) error {
	now := time.Now().UTC()
	sqlStr := "INSERT INTO " + tableName + "(created_at,type,value,node_id) VALUES "
	var values []interface{}

	x := 1
	for name, val := range data {
		sqlStr += fmt.Sprintf("($%d,$%d, $%d, $%d),", x, x+1, x+2, x+3)
		values = append(values, now, name, val, nodeId)
		x += 4
	}
	//trim the last ,
	sqlStr = sqlStr[0 : len(sqlStr)-1]
	_, err := saver.connection.Exec(context.Background(), sqlStr, values...)
	return err
}

func (saver *StatSaver) SaveString(data map[string]map[string]map[string]int) {
	count := 0
	for appName, data := range data {
		if isValidAppName(appName) {
			saver.SaveAppDataString(appName, data)
			count++
		} else {
			saver.logger.Println("Invalid app name", appName)
			saver.sum("invalid_app", 1)
		}
	}
	saver.sum("saved", 1)
}

func (saver *StatSaver) SaveAppDataString(appName string, data map[string]map[string]int) {
	appParts := strings.Split(appName, "/")
	if len(appParts) != 2 {
		saver.logger.Println("Bad app parts", appName)
		return
	}
	nodeId, err := strconv.Atoi(appParts[1])
	if err != nil {
		saver.logger.Println("Bad node id", appName, err)
		return
	}
	table := getStringTableName(appParts[0])
	err = saver.createTableStringMetric(table)
	if err != nil {
		saver.logger.Println("Table not created", appName, err)
		return
	}
	err = saver.saveStringMetrics(table, nodeId, data)
	if err != nil {
		saver.logger.Println("Data saveStringMetrics fail", appName, err)
		return
	}
}

func (saver *StatSaver) createTableStringMetric(name string) error {
	if _, has := saver.existingTables[name]; has {
		return nil
	}
	_, err := saver.connection.Exec(context.Background(), `create table IF NOT EXISTS `+name+`
(
	created_at timestamp default now(),
	type varchar(50) not null,
	pattern varchar(200) not null,
	value integer not null,
	node_id integer not null
);
create index IF NOT EXISTS `+name+`_created_at_index on `+name+` (created_at desc);
`)

	if err == nil {
		saver.existingTables[name] = true
	}
	return err
}

func (saver *StatSaver) saveStringMetrics(tableName string, nodeId int, data map[string]map[string]int) error {
	now := time.Now().UTC()
	sqlStr := "INSERT INTO " + tableName + "(created_at,type,pattern,value,node_id) VALUES "
	var values []interface{}

	x := 1
	for name, list := range data {
		for pattern, count := range list {
			sqlStr += fmt.Sprintf("($%d,$%d,$%d,$%d,$%d),", x, x+1, x+2, x+3, x+4)
			values = append(values, now, name, pattern, count, nodeId)
			x += 5
		}
	}
	//trim the last ,
	sqlStr = sqlStr[0 : len(sqlStr)-1]
	_, err := saver.connection.Exec(context.Background(), sqlStr, values...)
	if err != nil {
		saver.logger.Println("Bad sql", sqlStr)
	}
	return err
}
