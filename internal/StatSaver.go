package internal

import (
	"context"
	"fmt"
	"github.com/jackc/pgx"
	"log"
	"regexp"
	"strconv"
	"strings"
)

var r *regexp.Regexp

func init() {
	r, _ = regexp.Compile("^[A-z0-9-_]+/[0-9]+$")
}

func isValidAppName(name string) bool {
	return r.MatchString(name)
}

func getTableName(appName string) string {
	return "t_" + strings.ReplaceAll(appName, "-", "_")
}

type StatSaver struct {
	logger         *log.Logger
	databaseUrl    string
	connection     *pgx.Conn
	stop           chan bool
	existingTables map[string]bool
}

func CreateStatSaver(looger *log.Logger, url string) *StatSaver {
	return &StatSaver{
		logger:         looger,
		databaseUrl:    url,
		existingTables: make(map[string]bool),
	}
}

func (saver *StatSaver) Start() error {
	conn, err := pgx.Connect(context.Background(), saver.databaseUrl)
	if err != nil {
		return err
	}
	saver.connection = conn
	saver.stop = make(chan bool, 1)
	<-saver.stop
	return nil
}

func (saver *StatSaver) Stop() error {
	err := saver.connection.Close(context.Background())
	saver.stop <- false
	return err
}

func (saver *StatSaver) GetName() string {
	return "StatSaver"
}

func (saver *StatSaver) Save(data map[string]map[string]int) {
	for appName, data := range data {
		if isValidAppName(appName) {
			saver.SaveAppData(appName, data)
		} else {
			saver.logger.Println("Invalid app name ", appName)
		}
	}
}

func (saver *StatSaver) SaveAppData(appName string, data map[string]int) {
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
	table := getTableName(appParts[0])
	err = saver.createTable(table)
	if err != nil {
		saver.logger.Println("Table not created", appName, err)
		return
	}
	err = saver.save(table, nodeId, data)
	if err != nil {
		saver.logger.Println("Data save fail", appName, err)
		return
	}
}

func (saver *StatSaver) createTable(name string) error {
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

func (saver *StatSaver) save(tableName string, nodeId int, data map[string]int) error {

	sqlStr := "INSERT INTO " + tableName + "(type,value,node_id) VALUES "
	var values []interface{}

	x := 1
	for name, val := range data {
		sqlStr += fmt.Sprintf("($%d, $%d, $%d),", x, x+1, x+2)
		values = append(values, name, val, nodeId)
		x += 3
	}
	//trim the last ,
	sqlStr = sqlStr[0 : len(sqlStr)-1]
	_, err := saver.connection.Exec(context.Background(), sqlStr, values...)
	return err
}
