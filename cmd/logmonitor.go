package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"time"

	"github.com/yglcode/logmonitor/pkg/logmonitor"
)

func main() {
	//parse cmdline args for config parameters
	cfg := logmonitor.Config{}
	flag.StringVar(&cfg.LogFileName, "file", "/tmp/access.log", "name of log file to monitor")
	flag.IntVar(&cfg.InternalBufferSize, "bufsize", 256, "size of internal buffer of read log lines")
	var reportT int
	flag.IntVar(&reportT, "reportSec", 10, "report duration in seconds")
	var alertT int
	flag.IntVar(&alertT, "alertSec", 120, "alert duration in seconds")
	flag.IntVar(&cfg.AlertThreshold, "alertThreshold", 10, "number of hits/second to trigger alert")
	flag.Parse()
	cfg.ReportDuration = time.Duration(reportT) * time.Second
	cfg.AlertDuration = time.Duration(alertT) * time.Second

	fmt.Printf("logmonitor config:\n%v", cfg)

	lm, err := logmonitor.New(cfg)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer lm.Close()
	//logmonitor will keep running, until ctrl-C is hit
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)
	select {
	case <-sigChan:
		//user shuts down
		break
	}
}
