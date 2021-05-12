package logmonitor

import (
	"fmt"
	"strings"
	"time"
)

type LogRecord struct {
	timestamp                                   time.Time
	addr, userName, method, path, proto, status string
	section                                     string
}

func (lr *LogRecord) String() string {
	return fmt.Sprintf("%s: %s, %s, %s, %s", lr.section, lr.timestamp, lr.method, lr.path, lr.status)
}

var timeLayout = "02/Jan/2006:15:04:05 +0000"

func parseLogLine(line string) (r *LogRecord, err error) {
	ss := strings.Split(line, " ")
	if len(ss) < 10 {
		err = fmt.Errorf("invalid line:%s", line)
		return
	}
	timeStr := ss[3][1:] + " " + ss[4][:len(ss[4])-1]
	t, err := time.Parse(timeLayout, timeStr)
	if err != nil {
		return
	}
	sect := ss[6]
	if i := strings.Index(sect[1:], "/"); i > 0 {
		sect = sect[:1+i]
	}
	r = &LogRecord{
		addr:      ss[0],
		userName:  ss[2],
		timestamp: t,
		method:    ss[5][1:],
		path:      ss[6],
		proto:     ss[7][:len(ss[7])-1],
		status:    ss[8],
		section:   sect,
	}
	//fmt.Println(r)
	return
}
