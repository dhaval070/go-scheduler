package main

import (
	_ "fmt"
	"gsch/scheduler"
	"sync"
	"time"
)

var leagues = []string{"teamsite"}

func main() {
	var wg sync.WaitGroup
	wg.Add(len(leagues))

	for _, l := range leagues {
		go process(l, &wg)
	}

	wg.Wait()
}

func process(league string, wg *sync.WaitGroup) {
	defer wg.Done()

	var sch = scheduler.NewScheduler(league)
	for {
		sch.Work()
		//        break;
		time.Sleep(5000 * time.Millisecond)
	}
}
