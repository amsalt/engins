package utime

import "time"

// package time provides kinds of time related utilities for easy time operations.

// support modify time
var timeDiff time.Duration

// SetDiff sets the different duration with normal Time.
func SetDiff(d time.Duration) {
	timeDiff = d
}

// SetTime sets datetime as current time.
// datetime format: time.RFC3339
func SetTime(datetime string) error {
	diff, err := settimediff(datetime)
	if err == nil {
		timeDiff = diff
	}

	return err
}

// Time represents an instant in time with nanosecond precision.
// it timeDiff not zero, a modified time will be returned.
func Time() time.Time {
	t := time.Now()
	if timeDiff != 0 {
		t = t.Add(timeDiff)
	}
	return t
}

// Now returns current milliseconds with timeDiff
func Now() int64 {
	return Time().UnixNano() / int64(time.Millisecond)
}

// NowUnix returns seconds with timeDiff.
func NowUnix() int64 {
	return Time().Unix()
}

// NowMillSec returns millisecond with timeDiff.
func NowMillSec(t time.Time) int64 {
	return t.UnixNano() / int64(time.Millisecond)
}

func settimediff(datetime string) (time.Duration, error) {
	var timeDiff time.Duration
	if datetime == "" {
		timeDiff = 0
	} else {
		t, err := time.Parse(time.RFC3339, datetime)
		if err != nil {
			return timeDiff, err
		}

		now := time.Now()
		timeDiff = t.Sub(now)
	}

	return timeDiff, nil
}
