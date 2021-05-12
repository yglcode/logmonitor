package logmonitor

import (
	"fmt"
	"strings"
	"time"
)

type Config struct {
	LogFileName        string
	InternalBufferSize int
	ReportDuration     time.Duration
	AlertDuration      time.Duration
	AlertThreshold     int
}

func (c Config) String() string {
	buf := &strings.Builder{}
	fmt.Fprintf(buf, "LogFileName: %s\n", c.LogFileName)
	fmt.Fprintf(buf, "InternalBufferSize: %d\n", c.InternalBufferSize)
	fmt.Fprintf(buf, "ReportDuration: %v\n", c.ReportDuration)
	fmt.Fprintf(buf, "AlertDuration: %v\n", c.AlertDuration)
	fmt.Fprintf(buf, "AlertThreshold: %d/second\n", c.AlertThreshold)
	return buf.String()
}

type LogMonitor struct {
	//internal chan forwarding decoded log lines
	//from FileReader to StatsReporter
	recChan   chan *LogRecord
	freader   *FileReader
	sreporter *StatsReporter
}

func New(cfg Config) (lm *LogMonitor, err error) {
	recChan := make(chan *LogRecord, cfg.InternalBufferSize)
	fr, err := newFileReader(cfg.LogFileName, recChan)
	if err != nil {
		return
	}

	lm = &LogMonitor{
		recChan:   recChan,
		freader:   fr,
		sreporter: newStatsReporter(cfg, recChan),
	}
	return
}

func (lm *LogMonitor) Close() (err error) {
	lm.freader.Close()
	lm.sreporter.Close()
	return
}
