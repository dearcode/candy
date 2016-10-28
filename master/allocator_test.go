package master

import (
	"flag"
	"os"
	"testing"

	"github.com/dearcode/candy/util/log"
)

func TestMain(main *testing.M) {
	debug := flag.Bool("V", false, "set log level:debug")
	flag.Parse()
	if *debug {
		log.SetLevel(log.LOG_DEBUG)
	} else {
		log.SetLevel(log.LOG_ERROR)
	}

	var err error
	if a, err = newAllocator(newMstore(), ""); err != nil {
		panic(err.Error())
	}

	os.Exit(main.Run())
}

var (
	a *allocator
)

func TestNewID(t *testing.T) {
	result := make(map[int64]struct{})
	for i := 0; i < 100; i++ {
		id := a.id()
		if _, ok := result[id]; ok {
			t.Fatalf("same id:%d", id)
		}
	}
}

func BenchmarkNewID(b *testing.B) {
	for i := 0; i < b.N; i++ {
		a.id()
	}
}

func BenchmarkParallelNewID(b *testing.B) {
	b.SetParallelism(100)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			a.id()
		}
	})

}
