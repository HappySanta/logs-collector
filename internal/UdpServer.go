package internal

import (
	"log"
	"net"
	"strconv"
	"strings"
)

const SumTag = "P"
const SetTag = "S"
const MaxTag = "M"
const MinTag = "I"
const AvgTag = "A"
const HllTag = "L"
const HllDayTag = "D"
const StrSumTag = "T"
const StrSetTag = "E"
const StrMinTag = "N"
const StrMaxTag = "X"
const StrAvgTag = "G"

type UpdServer struct {
	core           *CoreStatistic
	pc             net.PacketConn
	host           string
	stop           bool
	debounceLogger *log.Logger
}

func CreateUpdServer(core *CoreStatistic, host string, logger *log.Logger) *UpdServer {
	return &UpdServer{
		core:           core,
		pc:             nil,
		host:           host,
		stop:           false,
		debounceLogger: logger,
	}
}

func (server *UpdServer) Start() error {
	pc, err := net.ListenPacket("udp", server.host)
	if err != nil {
		return err
	}
	server.pc = pc
	server.stop = false
	server.debounceLogger.Println("Start udp on:", server.host)
	for {
		buf := make([]byte, 2048)
		n, addr, err := pc.ReadFrom(buf)
		if err != nil {
			if server.stop {
				return nil
			}
			server.debounceLogger.Println(err)
		} else {
			server.serve(pc, addr, buf[:n])
		}
		if server.stop {
			return nil
		}
	}
}

func (server *UpdServer) Stop() error {
	server.stop = true
	if server.pc != nil {
		return server.pc.Close()
	}
	return nil
}

func (server *UpdServer) GetName() string {
	return "UDP Server"
}

func (server *UpdServer) serve(pc net.PacketConn, addr net.Addr, buf []byte) {
	// RL:AppName:ParamName:TYPE:VALUE
	if len(buf) < 9 {
		//Bad pack
		server.debounceLogger.Printf("Too short message: %s", addr.String())
		return
	}
	if buf[0] != 'R' || buf[1] != 'L' || buf[2] != ':' {
		//Bad pack
		server.debounceLogger.Printf("Bad message header: %d %d %d %s", buf[0], buf[1], buf[2], addr.String())
		return
	}
	data := string(buf)
	dataParts := strings.Split(data, ":")
	if len(dataParts) != 5 && len(dataParts) != 6 {
		//Bad pack
		server.debounceLogger.Printf("Bad message format: %s %s", data, addr.String())
		return
	}
	appName := dataParts[1]
	paramName := dataParts[2]
	paramType := dataParts[3]
	paramValue := dataParts[4]

	value, err := strconv.Atoi(paramValue)
	if err != nil {
		//bad pack
		server.debounceLogger.Printf("Bad value: metric:%s value:%s app:%s addr:%s", paramName, paramValue, appName, addr.String())
		return
	}

	if paramType == SetTag {
		server.core.Set(appName, paramName, value)
	} else if paramType == SumTag {
		server.core.Sum(appName, paramName, value)
	} else if paramType == MaxTag {
		server.core.Max(appName, paramName, value)
	} else if paramType == MinTag {
		server.core.Min(appName, paramName, value)
	} else if paramType == AvgTag {
		server.core.Avg(appName, paramName, value)
	} else if paramType == HllDayTag {
		if len(dataParts) != 6 {
			server.debounceLogger.Printf("Bad message format for string: %s %s", data, addr.String())
			return
		}
		stringValue := dataParts[5]
		server.core.HllDay(appName, paramName, stringValue)
	} else if paramType == HllTag {
		if len(dataParts) != 6 {
			server.debounceLogger.Printf("Bad message format for string: %s %s", data, addr.String())
			return
		}
		stringValue := dataParts[5]
		server.core.Hll(appName, paramName, stringValue)
	} else if paramType == StrSetTag {
		if len(dataParts) != 6 {
			server.debounceLogger.Printf("Bad message format for string: %s %s", data, addr.String())
			return
		}
		stringValue := dataParts[5]
		server.core.StrSet(appName, paramName, value, stringValue)
	} else if paramType == StrMinTag {
		if len(dataParts) != 6 {
			server.debounceLogger.Printf("Bad message format for string: %s %s", data, addr.String())
			return
		}
		stringValue := dataParts[5]
		server.core.StrMin(appName, paramName, value, stringValue)
	} else if paramType == StrMaxTag {
		if len(dataParts) != 6 {
			server.debounceLogger.Printf("Bad message format for string: %s %s", data, addr.String())
			return
		}
		stringValue := dataParts[5]
		server.core.StrMax(appName, paramName, value, stringValue)
	} else if paramType == StrAvgTag {
		if len(dataParts) != 6 {
			server.debounceLogger.Printf("Bad message format for string: %s %s", data, addr.String())
			return
		}
		stringValue := dataParts[5]
		server.core.StrAvg(appName, paramName, value, stringValue)
	} else if paramType == StrSumTag {
		if len(dataParts) != 6 {
			server.debounceLogger.Printf("Bad message format for string: %s %s", data, addr.String())
			return
		}
		stringValue := dataParts[5]
		server.core.StrSum(appName, paramName, value, stringValue)
	} else {
		server.debounceLogger.Printf("Unknown param type: [%s] %s %s", paramType, appName, addr.String())
		return
	}
}
