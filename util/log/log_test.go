package log

import (
	"testing"
)

func TestLog(t *testing.T) {
	Debug("xxxxxxxx")
	Debugf("%s, %+v", "abc", mlog)
}
