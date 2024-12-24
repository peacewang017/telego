package util

import "time"

func timestamp() string {
	return time.Now().Format("2006-01-02-15h04m05s")
}
