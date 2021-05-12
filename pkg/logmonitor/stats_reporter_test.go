package logmonitor

import (
	"fmt"
	"testing"
	"time"
)

type testReporter struct {
	//reported busiest section names
	sectionNames []string
	//alert triggered info
	alertHits int
	//to allow test code to retrieve data from StatsReporter goroutine
	//use chan to sync both
	dataCh chan *testReporter
}

func newTestReporter() *testReporter {
	return &testReporter{dataCh: make(chan *testReporter, 2)}
}

func (r *testReporter) ReportBusiestSections(recs []*sectionRecord) {
	fmt.Println("**ReportBusiestSections", recs)
	//reset
	r.sectionNames = []string{}
	//add reported names
	for i := 0; i < len(recs); i++ {
		r.sectionNames = append(r.sectionNames, recs[i].name)
	}
	r.dataCh <- r
}

func (r *testReporter) SetAlert(n int, t time.Time) {
	fmt.Println("**SetAlert", n, t)
	r.alertHits = n
	r.dataCh <- r
}

func (r *testReporter) ClearAlert(n int, t time.Time) {
	fmt.Println("**ClearAlert", n, t)
	r.alertHits = 0
	r.dataCh <- r
}

func (r *testReporter) getData() (sectNames []string, hits int) {
	rr := <-r.dataCh
	return append([]string(nil), rr.sectionNames...), rr.alertHits
}

func TestReportAlert(t *testing.T) {
	config := Config{
		ReportDuration: 1 * time.Second,
		AlertDuration:  2 * time.Second,
		AlertThreshold: 2,
	}
	recChan := make(chan *LogRecord)
	repAlert := newTestReporter()
	statsReporter := newStatsReporter(config, recChan, repAlert)
	defer statsReporter.Close()

	//insert 6 log records at second 0
	for i := 0; i < 4; i++ {
		recChan <- &LogRecord{section: "api", timestamp: time.Now()}
	}
	for i := 0; i < 2; i++ {
		recChan <- &LogRecord{section: "report", timestamp: time.Now()}
	}
	//after about 1 sec, we should have both report & alert
	sectNames, hits := repAlert.getData()
	sectNames, hits = repAlert.getData()
	fmt.Printf("report: %v, alertHits: %d\n", sectNames, hits)
	if len(sectNames) != 1 || sectNames[0] != "api" {
		t.Errorf("expected busiest section names [api], got: %v", sectNames)
	}
	if hits == 0 {
		t.Errorf("expected high traffic alert, got: 0 hits")
	}
	//wait another sec, all should be cleared (reset report & clear alert)
	sectNames, hits = repAlert.getData()
	sectNames, hits = repAlert.getData()
	fmt.Printf("report: %v, alertHits: %d\n", sectNames, hits)
	if len(sectNames) != 0 {
		t.Errorf("expected no section report, got: %v", sectNames)
	}
	if hits != 0 {
		t.Errorf("Alert should be cleared, got: %v", hits)
	}
}
