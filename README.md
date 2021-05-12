### logmonitor: command line tool to monitor log file ###

logmonitor will monitor specified w3c-formatted HTTP access log file (by default /tmp/access.log) continuously until stopped by ctrl-C. It will report most visited section names periodically and set alerts when average traffic hits during alert duration exceed certain threshold and clear alerts when average traffic drop below.

```
Usage of logmonitor:
  -alertSec int
    	alert duration in seconds (default 120)
  -alertThreshold int
    	number of hits/second to trigger alert (default 10)
  -bufsize int
    	size of internal buffer of read log lines (default 256)
  -file string
    	name of log file to monitor (default "/tmp/access.log")
  -reportSec int
    	report duration in seconds (default 10)
```
