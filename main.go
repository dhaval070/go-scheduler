package main
import (
    "github.com/joho/godotenv"
_    "log"
_    "fmt"
    "gsch/model"
    "sync"
    "time"
)

func init() {
    godotenv.Load()
}

var leagues = []string {"teamsite" }
var wg sync.WaitGroup

func main() {
    for _, l := range leagues {
        wg.Add(1)
        go process(l)
    }

    wg.Wait()
}

func process(league string) {
    var sch = event.NewScheduler(league)
    defer sch.Destroy()

    for {
        sch.Work()
        break;
        time.Sleep(10000 * time.Millisecond)
    }
    wg.Done()
}

