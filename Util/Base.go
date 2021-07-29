package Util

import (
	"fmt"
	"io/fs"
	"math/rand"
	"sort"
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

func GetRandomString(n int) string {
	randBytes := make([]byte, n/2)
	rand.Read(randBytes)
	return fmt.Sprintf("%x", randBytes)
}

func GetNextOrPrevious(slice []fs.FileInfo, fs fs.FileInfo, flag int) fs.FileInfo {
	index := 0
	for p, v := range slice {
		if v.Name() == fs.Name() {
			index = p
		}
	}
	if flag == 0 {
		return fs
	} else if flag == -1 {
		if index > 0 {
			index = index + flag
		} else {
			return nil
		}
	} else if flag == 1 {
		if index < len(slice)-1 {
			index = index + flag
		} else {
			return nil
		}
	}
	return slice[index]
}

func FilterFiles(slice []fs.FileInfo) []fs.FileInfo {
	sort.Slice(slice, func(i, j int) bool {
		d1 := 0
		if slice[i].IsDir() {
			d1 = 1
		}
		d2 := 0
		if slice[j].IsDir() {
			d2 = 1
		}
		if d1 > d2 {
			return true
		} else if d1 == d2 {
			return slice[i].ModTime().After(slice[j].ModTime())
		} else {
			return false
		}
	})
	arr := []fs.FileInfo{}
	for _, v := range slice {
		if !v.IsDir() {
			arr = append(arr, v)
		}
	}
	return arr
}
