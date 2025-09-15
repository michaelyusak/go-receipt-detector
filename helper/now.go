package helper

import "time"

func NowUnixMilli() int64 {
	return time.Now().UnixMilli()
}
