package log

import "time"

const TimeFormat = "2006-01-02 15:04:05.000000"

var (
	startTime time.Time
)

func StartTime() {
	startTime = time.Now()
	Infof("Started at %s", startTime.Format(TimeFormat))
}

func StopTime() {
	stopTime := time.Now()
	runTime := float64(stopTime.UnixMicro()-startTime.UnixMicro()) / 1000000.0
	Infof("Finished at %s", stopTime.Format(TimeFormat))
	Infof("Runtime (sec): %f", runTime)
}
