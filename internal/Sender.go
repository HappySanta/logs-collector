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

func LogStatistic(address, appName, paramName, paramType string, value int) (err error) {
	// Resolve the UDP address so that we can make use of DialUDP
	// with an actual IP and port instead of a name (in case a
	// hostname is specified).
	raddr, err := net.ResolveUDPAddr("udp", address)
	if err != nil {
		return
	}

	// Although we're not in a connection-oriented transport,
	// the act of `dialing` is analogous to the act of performing
	// a `connect(2)` syscall for a socket of type SOCK_DGRAM:
	// - it forces the underlying socket to only read and write
	//   to and from a specific remote address.
	conn, err := net.DialUDP("udp", nil, raddr)
	defer conn.Close()

	if err != nil {
		return
	}

	data := fmt.Sprintf("RL:%s:%s:%s:%d", appName, paramName, paramType, value)

	_, err = conn.Write([]byte(data))
	if err != nil {
		return err
	}

	return nil
}
