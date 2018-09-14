package util

import (
	"fmt"
	"net/http"
	"strconv"
	"time"
)

var (
	// GitTime git log中记录的提交时间.
	GitTime = ""
	// GitMessage git log 中记录的提交信息.
	GitMessage = ""
	//gitTime 转为时间方式的GitTime.
	gitTime time.Time
)

func init() {
	if GitTime == "" {
		return
	}

	sec, err := strconv.ParseInt(GitTime, 10, 64)
	if err != nil {
		panic(err)
	}

	gitTime = time.Unix(sec, 0)
}

// PrintVersion 输出当前程序编译信息.
func PrintVersion() {
	fmt.Printf("Candy\n")
	fmt.Printf("Commit Time: %s\n", gitTime.Format(time.RFC3339))
	fmt.Printf("Commit Message: %s\n", GitMessage)
}

//Version 版本信息.
type Version struct {
}

//GET 输出当前应用版本信息.
func (v *Version) GET(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Candy\n")
	fmt.Fprintf(w, "Commit Time: %s\n", gitTime.Format(time.RFC3339))
	fmt.Fprintf(w, "Commit Message: %s\n", GitMessage)
}
