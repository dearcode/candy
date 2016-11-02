package util

import (
	"testing"
	"time"
)

func TestRetry(t *testing.T) {

	b := time.Now()
	for r := NewRetry(RetryAttemptMax(1), RetryDurationMin(time.Millisecond*20)); r.Valid(); r.Next() {
	}
	e := time.Now()

	if e.Sub(b).Nanoseconds()/100000 < 20 {
		t.Fatalf("sleep time must be > 20ms, used:%v", e.Sub(b))
	}

	b = time.Now()
	for r := NewRetry(RetryAttemptMax(1), RetryDurationMin(time.Millisecond*10), RetryDurationMax(time.Millisecond*50)); r.Valid(); r.Next() {
	}
	e = time.Now()

	if e.Sub(b).Nanoseconds()/1000000 < 10 {
		t.Fatalf("sleep time must be > 10ms, used:%v", e.Sub(b))
	}

	if e.Sub(b).Nanoseconds()/1000000 > 50 {
		t.Fatalf("sleep time must be < 50ms, used:%v", e.Sub(b))
	}

	i := 0
	for r := NewRetry(RetryAttemptMax(1), RetryDurationMin(time.Millisecond*10), RetryDurationMax(time.Millisecond*50)); r.Valid(); r.Next() {
		i++
	}

	if i != 1 {
		t.Fatalf("expect retry 1, current i:%d", i)
	}

	i = 0
	for r := NewRetry(RetryAttemptMax(3), RetryDurationMin(time.Millisecond*10), RetryDurationMax(time.Millisecond*50)); r.Valid(); r.Next() {
		i++
	}

	if i != 3 {
		t.Fatalf("expect retry 3, current i:%d", i)
	}

}
