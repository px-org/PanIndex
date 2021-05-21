package Util

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

func ShortDur(d time.Duration) string {
	v, _ := strconv.ParseFloat(fmt.Sprintf("%.1f", d.Seconds()), 64)
	return fmt.Sprint(v) + "s"
}
func GetCurBetweenStr(str, start, end string) string {
	n := strings.Index(str, start)
	if n == -1 {
		return ""
	}
	n = n + len(start)
	str = string([]byte(str)[n:])
	m := strings.Index(str, end)
	if m == -1 {
		return ""
	}
	str = string([]byte(str)[:m])
	return str
}
