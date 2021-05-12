package logmonitor

import (
	"fmt"
	"sort"
	"time"
)

/*
* StatsReporter reports most visited sections regularly
* and report/clear alerts when traffic hits exceed certain threshold
 */
type StatsReporter struct {
	config   Config
	recChan  chan *LogRecord
	exitChan chan struct{}
	//traffic statistics and alerting
	logTimes    []time.Time
	sectMap     map[string]*sectionRecord
	mostVisited map[string]bool
	maxVisit    int
	alertActive bool
	reportApi   ReportAlertApi
}

type sectionRecord struct {
	name string
	//count in last ReportDuration seconds
	count int
}

func (sr sectionRecord) String() string {
	return fmt.Sprintf("%s-%d", sr.name, sr.count)
}

//ReportAlertApi is common api implemented by report/alert facility
//right now just simple print
type ReportAlertApi interface {
	ReportBusiestSections([]*sectionRecord)
	SetAlert(int, time.Time)
	ClearAlert(int, time.Time)
}

func newStatsReporter(cfg Config, recChan chan *LogRecord, rpt ...ReportAlertApi) (sp *StatsReporter) {
	exitChan := make(chan struct{}, 1)
	var reporter ReportAlertApi
	if len(rpt) == 0 {
		reporter = &reportAlerter{}
	} else {
		reporter = rpt[0]
	}
	sp = &StatsReporter{
		config:    cfg,
		recChan:   recChan,
		exitChan:  exitChan,
		sectMap:   make(map[string]*sectionRecord),
		reportApi: reporter,
	}
	go sp.StatsReport()
	return
}

func (sp *StatsReporter) Close() {
	close(sp.recChan)
	<-sp.exitChan
}

func (sp *StatsReporter) StatsReport() {
	ticker := time.NewTimer(sp.config.ReportDuration)
	alertMax := sp.config.AlertThreshold * int(sp.config.AlertDuration/time.Second)
	alertClearTimer := time.NewTimer(sp.config.AlertDuration)
	alertClearTimer.Stop() //only start this timer when alert is active

LOOP:
	for {
		select {
		case rec, ok := <-sp.recChan:
			if !ok {
				break LOOP
			}
			//fmt.Println(rec)
			now := time.Now()
			age := now.Sub(rec.timestamp)
			if age < sp.config.AlertDuration {
				sp.logTimes = append(sp.logTimes, rec.timestamp)
				sp.checkAlerts(now, alertMax, alertClearTimer, false)
			}
			if age >= sp.config.ReportDuration {
				continue //skip report, out of range
			}
			srec, ok := sp.sectMap[rec.section]
			if ok {
				srec.count++
			} else {
				srec = &sectionRecord{rec.section, 1}
				sp.sectMap[rec.section] = srec
			}
			if len(sp.mostVisited) == 0 || srec.count > sp.maxVisit {
				sp.mostVisited = map[string]bool{srec.name: true}
				sp.maxVisit = srec.count
			} else if srec.count == sp.maxVisit && !sp.mostVisited[srec.name] {
				sp.mostVisited[srec.name] = true
			}
		case <-alertClearTimer.C:
			sp.checkAlerts(time.Now(), alertMax, alertClearTimer, true)
		case <-ticker.C:
			sp.reportStats()
			ticker.Reset(sp.config.ReportDuration)
		}
	}
	ticker.Stop()
	alertClearTimer.Stop()
	fmt.Println("xxx Stats reporter exit")
	sp.exitChan <- struct{}{}
}

func (sp *StatsReporter) reportStats() {
	sects := []*sectionRecord{}
	for k, _ := range sp.mostVisited {
		sects = append(sects, sp.sectMap[k])
	}
	sp.reportApi.ReportBusiestSections(sects)
	//clear statis in last reportDuration
	sp.sectMap = make(map[string]*sectionRecord)
	sp.mostVisited = nil
}

type reportAlerter struct{}

func (r reportAlerter) ReportBusiestSections(recs []*sectionRecord) {
	for i := 0; i < len(recs); i++ {
		fmt.Printf("section %s: %d\n", recs[i].name, recs[i].count)
	}
}

func (r reportAlerter) SetAlert(n int, t time.Time) {
	fmt.Printf("High traffic generated an alert - hits = %d, triggered at %v\n", n, t)
}

func (r reportAlerter) ClearAlert(n int, t time.Time) {
	fmt.Printf("Alert cleared - hits = %d, cleared at %v\n", n, t)
}

func (sp *StatsReporter) checkAlerts(now time.Time, alertMax int, alertClearTimer *time.Timer, timerEvent bool) {
	if len(sp.logTimes) < alertMax {
		return
	}
	alertDuration := sp.config.AlertDuration
	//remove entries older than Now()+alertDuration
	idx := sort.Search(len(sp.logTimes), func(i int) bool {
		return now.Sub(sp.logTimes[i]) < alertDuration
	})
	sp.logTimes = sp.logTimes[idx:]
	if len(sp.logTimes) >= alertMax {
		if sp.alertActive {
			if !timerEvent {
				if !alertClearTimer.Stop() {
					<-alertClearTimer.C
				}
			}
		} else {
			//alert
			sp.reportApi.SetAlert(len(sp.logTimes), now)
			sp.alertActive = true
		}
		diff := len(sp.logTimes) - alertMax
		alertClearTimer.Reset(sp.logTimes[diff].Add(alertDuration).Sub(now))
	} else {
		if sp.alertActive {
			sp.reportApi.ClearAlert(len(sp.logTimes), now)
			sp.alertActive = false
			if !timerEvent && !alertClearTimer.Stop() {
				<-alertClearTimer.C
			}
		}
	}
}
