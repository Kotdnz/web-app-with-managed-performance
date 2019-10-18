package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/Kotdnz/web-app-with-managed-performance/ratecounter"
)

var (
	urlPtr  string
	ratePtr int
)

type loader struct {
	url                   string
	mu                    sync.RWMutex
	curThread             int
	curRate, okSum, erSum int64
}

func (rp *loader) String() string {
	rp.mu.RLock()
	defer rp.mu.RUnlock()
	return fmt.Sprintf("Updated every 5 sec\n"+
		" - Current request rate per second: %d\n"+
		" - Current threads: %d\n"+
		" - Amount 200: %d, amount 500: %d\n",
		rp.curRate, rp.curThread, rp.okSum, rp.erSum)
}

func (rp *loader) Reset() {
	rp.mu.Lock()
	defer rp.mu.Unlock()
	rp.okSum = 0
	rp.erSum = 0
}

func (rp *loader) update(s string) {
	rp.mu.Lock()
	defer rp.mu.Unlock()

	switch s {
	case "err":
		rp.erSum++
	case "ok":
		rp.okSum++
	case "thread-":
		if rp.curThread > 0 {
			rp.curThread--
		}
	case "thread+":
		rp.curThread++
	default:
		// skip other string
	}
}

func (rp *loader) Curl() {
	if rp.curThread > 1023 {
		return
	}

	resp, err := http.Get(rp.url)

	defer rp.update("thread-")

	if err != nil {
		log.Println("Error: Something went wrong - can't Get the URL")
		rp.update("err")
		return
	}

	defer resp.Body.Close()
	if resp.StatusCode == 200 {
		rp.update("ok")
		return
	}
	rp.update("err")
}

func (rp *loader) CurrRate(r int64) {
	rp.mu.Lock()
	defer rp.mu.Unlock()

	rp.curRate = r
}

func init() {
	flag.StringVar(&urlPtr, "url", "", "specify target url")
	flag.IntVar(&ratePtr, "rate", 0, "specify rate")
	flag.Parse()
}

func main() {
	if urlPtr == "" || ratePtr <= 0 {
		log.Panicln("wrong url", urlPtr, "or rate", ratePtr)
		os.Exit(1)
	}
	// init rate counter
	counter := ratecounter.NewRateCounter(1 * time.Second)

	tickerPrint := time.NewTicker(5 * time.Second)
	// run ratePtr rutines every second
	tickerCurl := time.NewTicker(1 * time.Second)
	loadPrinter := loader{
		url: urlPtr,
	}

	for {
		select {
		case <-tickerPrint.C:
			fmt.Println(loadPrinter.String())
			loadPrinter.Reset()
		case <-tickerCurl.C:
			for i := 0; i <= ratePtr; i++ {
				counter.Incr(1)
				loadPrinter.CurrRate(counter.Rate())
				loadPrinter.update("thread+")
				go loadPrinter.Curl()
			}
		}
	}
}
