// EPAM DevOps_Hackathon_2019
// October 15, 2019
// rev 1.0
// Kostiantyn_nikonenko@epam.com

package main

import (
	"fmt"
	"flag"
	"time"
  "net/http"
  "github.com/paulbellamy/ratecounter"
)

// max thread
var curThread int
var curRate int64
var okSum int64
var erSum int64

// continues print the pates
func printRate(){
	for {
		time.Sleep(time.Second * 5)
		fmt.Printf("Updated every 5 sec\n" +
			         " - Current request rate per second: %d\n" +
							 " - Current threads: %d\n" +
							 " - Amount 200: %d, amount 500: %d\n",
							 curRate, curThread, okSum, erSum)
		erSum = 0
    okSum = 0
	}
}

// thread function to curl
func curl(s string) {
	resp, err := http.Get(s)
	if err != nil {
		fmt.Printf("Error: Something went wrong - can't Get the URL")
	}
	if resp.StatusCode == 200 {
		okSum += 1
	} else {
		erSum += 1
	}
	resp.Body.Close()
	if curThread > 0 {
		curThread -= 1
	}
}

func main() {
	// parse command line
	urlPtr := flag.String("url", "", "a string")
	ratePtr := flag.Int("rate", 0, "an int")
	flag.Parse()
	if *urlPtr == "" || int(*ratePtr) <= 0 || int(*ratePtr) > 300 {
		fmt.Printf("\nFlags specification error: %s\n" +
			         "Usage: ./rate_loader -url=http://localhost:8080/worker -rate=200\n\n", flag.Args())
		return
	} else {
		fmt.Printf("Target host url: %s rate %d\n\n", *urlPtr, *ratePtr)
	}
  // init rate counter
	counter := ratecounter.NewRateCounter(1 * time.Second)
  // init thread counter
	curThread = 0
	// show the current rate every 5 sec
	go printRate()
	// main cycle
  for {
		if curThread < 1023{
			curThread += 1
			counter.Incr(1)
			curRate = counter.Rate()
			// from command prompt read requests rate to calculate the timeout
			// sec / rate
	    time.Sleep(1000 * time.Millisecond / time.Duration(*ratePtr))
	    go curl(*urlPtr)
 		}
  }
}
