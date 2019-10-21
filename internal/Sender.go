package internal

import (
	"fmt"
	"net"
)

func LogSum(address, appName, paramName string, value int) error {
	return LogStatistic(address, appName, paramName, SumTag, value)
}
func LogAvg(address, appName, paramName string, value int) error {
	return LogStatistic(address, appName, paramName, AvgTag, value)
}
func LogMax(address, appName, paramName string, value int) error {
	return LogStatistic(address, appName, paramName, MaxTag, value)
}
func LogMin(address, appName, paramName string, value int) error {
	return LogStatistic(address, appName, paramName, MinTag, value)
}
func LogSet(address, appName, paramName string, value int) error {
	return LogStatistic(address, appName, paramName, SetTag, value)
}

func LogStrSum(address, appName, paramName string, value int, pattern string) error {
	return LogStatisticEx(address, appName, paramName, StrSumTag, value, pattern)
}

func LogStrAvg(address, appName, paramName string, value int, pattern string) error {
	return LogStatisticEx(address, appName, paramName, StrAvgTag, value, pattern)
}

func LogStrMax(address, appName, paramName string, value int, pattern string) error {
	return LogStatisticEx(address, appName, paramName, StrMaxTag, value, pattern)
}

func LogStrMin(address, appName, paramName string, value int, pattern string) error {
	return LogStatisticEx(address, appName, paramName, StrMinTag, value, pattern)
}

func LogStrSet(address, appName, paramName string, value int, pattern string) error {
	return LogStatisticEx(address, appName, paramName, StrSetTag, value, pattern)
}
func LogHll(address, appName, paramName string, pattern string) error {
	return LogStatisticEx(address, appName, paramName, HllTag, 0, pattern)
}
func LogHllDay(address, appName, paramName string, pattern string) error {
	return LogStatisticEx(address, appName, paramName, HllDayTag, 0, pattern)
}

func LogStatistic(address, appName, paramName, paramType string, value int) (err error) {
	data := fmt.Sprintf("RL:%s:%s:%s:%d", appName, paramName, paramType, value)

	conn, err := net.Dial("udp", address)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(conn, data)

	if err != nil {
		return err
	}
	return conn.Close()
}

func LogStatisticEx(address, appName, paramName, paramType string, value int, pattern string) (err error) {
	data := fmt.Sprintf("RL:%s:%s:%s:%d:%s", appName, paramName, paramType, value, pattern)

	conn, err := net.Dial("udp", address)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(conn, data)

	if err != nil {
		return err
	}
	return conn.Close()
}
