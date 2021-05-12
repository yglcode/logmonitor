package logmonitor

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/fsnotify/fsnotify"
)

/*
* FileReader keep reading from specified log file. When reaching to
* the end of file, waiting for file-change event and restart reading.
* Forward read log records to StatsReporter for statistics and reporting
 */
type FileReader struct {
	fname    string
	file     *os.File
	watcher  *fsnotify.Watcher
	recChan  chan *LogRecord
	exitChan chan struct{}
}

func newFileReader(fname string, recChan chan *LogRecord) (fr *FileReader, err error) {
	file, err := os.Open(fname)
	if err != nil {
		return
	}
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return
	}
	exitChan := make(chan struct{}, 1)
	fr = &FileReader{
		fname:    fname,
		file:     file,
		watcher:  watcher,
		recChan:  recChan,
		exitChan: exitChan,
	}
	go fr.FileRead()
	return
}

func (fr *FileReader) Close() (err error) {
	//this will close watcher's chans, notify filereader to exit
	fr.watcher.Close()
	//let filereader exit from readLoop
	err = fr.file.Close()
	<-fr.exitChan
	return
}

func (fr *FileReader) FileRead() {
	frdr := bufio.NewReader(fr.file)
	//seek to proper position and read existing content in file
	fr.readLoop(frdr)
LOOP:
	for {
		//enable watcher to monitor log file status change
		fr.watcher.Add(fr.fname)
		select {
		case err, ok := <-fr.watcher.Errors:
			if !ok { //watcher closed
				break LOOP
			}
			log.Println("error:", err)
		case event, ok := <-fr.watcher.Events:
			if !ok { //watcher closed
				break LOOP
			}
			if event.Op&fsnotify.Write == fsnotify.Write {
				log.Println("logfile modified:", event)
				//logfile has new content, start reading
				//disable file watcher first
				fr.watcher.Remove(fr.fname)
				fr.readLoop(frdr)
			}
		}
	}
	fmt.Println("xxx FileReader exit")
	fr.exitChan <- struct{}{}
}

func (fr *FileReader) readLoop(frdr *bufio.Reader) {
	for {
		line, err := frdr.ReadString('\n')
		if err != nil {
			fmt.Println(err)
			break
		}
		line = strings.TrimSpace(line)
		if len(line) > 0 {
			rec, err := parseLogLine(line)
			if err == nil {
				fr.recChan <- rec
			} else {
				//log parse error
				fmt.Println(err)
			}
		}
	}
}
