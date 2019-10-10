// EPAM DevOps_Hackathon_2019
// October 3, 2019
// rev 1.0
// Kostiantyn_Nikonenko@epam.com
// Web service with adjustable performance
// to check how SRE monitoring is working

package main

import (
   "net/http"
   "fmt"
   "strconv"
   "encoding/json"
   "time"
   "math/rand"
   "os"
   "log"
   // third-party libs
   "github.com/paulbellamy/ratecounter"
   // prometeus section
   // https://medium.com/@zhimin.wen/custom-prometheus-metrics-for-apps-running-in-kubernetes-498d69ada7aa
   "github.com/prometheus/client_golang/prometheus"
   "github.com/prometheus/client_golang/prometheus/promauto"
   "github.com/prometheus/client_golang/prometheus/promhttp"
)

type PrometheusHttpMetric struct {
	Prefix                string
	ClientConnected       prometheus.Gauge
	TransactionTotal      *prometheus.CounterVec
	ResponseTimeHistogram *prometheus.HistogramVec
	Buckets               []float64
}

func InitPrometheusHttpMetric(prefix string, buckets []float64) *PrometheusHttpMetric {
	phm := PrometheusHttpMetric{
		Prefix: prefix,
		ClientConnected: promauto.NewGauge(prometheus.GaugeOpts{
			Name: prefix + "_client_connected",
			Help: "Number of active client connections",
		}),
		TransactionTotal: promauto.NewCounterVec(prometheus.CounterOpts{
			Name: prefix + "_requests_total",
			Help: "total HTTP requests processed",
		}, []string{"code", "method"},
		),
		ResponseTimeHistogram: promauto.NewHistogramVec(prometheus.HistogramOpts{
			Name:    prefix + "_response_time",
			Help:    "Histogram of response time for handler",
			Buckets: buckets,
		}, []string{"handler", "method"}),
	}

	return &phm
}

func (phm *PrometheusHttpMetric) WrapHandler(handlerLabel string, handlerFunc http.HandlerFunc) http.Handler {
	handle := http.HandlerFunc(handlerFunc)
	wrappedHandler := promhttp.InstrumentHandlerInFlight(phm.ClientConnected,
		promhttp.InstrumentHandlerCounter(phm.TransactionTotal,
			promhttp.InstrumentHandlerDuration(phm.ResponseTimeHistogram.MustCurryWith(prometheus.Labels{"handler": handlerLabel}),
				handle),
		),
	)
	return wrappedHandler
}

// The applicateion beheviour
type AppState struct {
   Latency int64    // ms
   Rate int         // requests per sec
   Errors int       // percent from last arraySize
   Saturation int   // he state of being saturated or the action of saturating. For our example - max rate per sec
}

const arraySize = 1024

var counter ratecounter.RateCounter
var myAppState AppState
var picker PercentPicking

// section to serve target error rate
type PercentPicking struct {
  aSize int
  myStat [arraySize]bool  // array if all is 200
  TargetPercent int
  okPercent int
  curPointer int    // current position in array
}

func NewPercentPicking(targetP int) *PercentPicking {
        p := new(PercentPicking)
        p.TargetPercent = targetP
        p.curPointer = 0
        p.aSize = arraySize
        return p
}

func (p *PercentPicking) NewRqst() bool {
        // in array all 0 = fail, 1 = Ok
        // calsulate the array fail percentage
        okCount := 1
        for _, value := range p.myStat {
          if value == true {
            okCount += 1
          }
        }
        // calculate percent of succes
        p.okPercent =100 - ((okCount * 100) / p.aSize)
        // make the cycle
        p.curPointer += 1
        if p.curPointer == p.aSize { p.curPointer = 0 }
        if p.okPercent < p.TargetPercent {
          // this request shuld be marked as failed to keep error rate
          p.myStat[p.curPointer] = false
          // to make behavior less traight - set the random aitem to failed
          p.myStat[rand.Intn(arraySize)] = false
        } else {
          // this request marked as SUCCESS (200)
          p.myStat[p.curPointer] = true
        }

        return p.myStat[p.curPointer]
}
// end of error section

func myWorkerHandler(w http.ResponseWriter, r *http.Request) {
  // Record an event happening
  counter.Incr(1)
  // check if not axceed the rate and error rate fewer
  if counter.Rate() <= int64(myAppState.Rate) && picker.NewRqst() == true {
    // sleep for latency
    time.Sleep(time.Duration(myAppState.Latency) * time.Millisecond)
    // return Ok 200
    w.WriteHeader(http.StatusOK)
    fmt.Fprintf(w, "Worker Ok! \nCurrent request rate per second %d \n", counter.Rate())
    fmt.Fprintf(w, "Current error rate %d from expected %d ",
       picker.okPercent, picker.TargetPercent)
  } else {
    // return 500
    w.WriteHeader(http.StatusInternalServerError)
    fmt.Fprintf(w, "Worker died! ERROR 500")
  }
}

// main function
func main() {
  // prometeus
  phm := InitPrometheusHttpMetric("myWorker", prometheus.LinearBuckets(0, 5, 20))
  http.Handle("/metrics", promhttp.Handler())

  // Our latency - 10ms + overhead
  // Over 50 req / sec we will return 500
  // Over 120 req / sec we stop the reply
  myAppState = AppState{ 100, 20, 10, 50 }

  // We're recording marks-per-1second
  counter = *ratecounter.NewRateCounter(1 * time.Second)

  // initialize the error picker percentage
  picker = *NewPercentPicking(myAppState.Errors)

  // handling main page
  // output the current stage after the modification, if exist
  // the definition above can be applied by the following
  // ?latency=10&rate=50&errors=10&saturation=120
  //
  http.HandleFunc("/" , func(w http.ResponseWriter, r *http.Request) {
    // Record an event happening
    counter.Incr(1)
    // Adjust the parameters if specified
		if latency := r.FormValue("latency"); latency != "" {
			val, err := strconv.ParseInt(latency, 10, 64);
      if err == nil && val >= 10 {
        myAppState.Latency = val
      }
		}
    if rate := r.FormValue("rate"); rate != "" {
			val, err := strconv.Atoi(rate);
      if err == nil && val >= 10 {
        myAppState.Rate = val
      }
		}
    if errors := r.FormValue("errors"); errors != "" {
      val, err := strconv.Atoi(errors);
      if err == nil && val >= 1 {
        myAppState.Errors = val
        picker.TargetPercent = 100-val
      }
    }
    if saturation := r.FormValue("saturation"); saturation != "" {
      val, err := strconv.Atoi(saturation);
      if err == nil && val >= 10 {
        myAppState.Saturation = val
      }
    }
    // status code 200
    w.WriteHeader(http.StatusOK)

    // output the help
    curPar, _ := json.Marshal(myAppState)
    fmt.Fprintf(w, "<p> EPAM DevOps Hackathon 2019 <p>")
    fmt.Fprintf(w, "<p> Current configuration is: %s ", string(curPar))
    fmt.Fprintf(w, "<br> Our latency for WORKER - %d ms + overhead" +
                   "<br> Over %d req / sec WORKER and health will return 500" +
                   "<br> Over %d req / sec live will return 500", myAppState.Latency, myAppState.Rate, myAppState.Saturation)
    fmt.Fprintf(w, "<br> URL to adjust: http://%s?latency=100&rate=20&errors=10&saturation=50", r.Host)
    fmt.Fprintf(w, "<p> Main WORKER <a href=\"http://%s/worker\">link</a> to apply the load. This page have %d error rate.", r.Host, myAppState.Errors)
    fmt.Fprintf(w, "<br> Ready check <a href=\"http://%s/ready\">link</a> - 200 while not exceed Rate limit", r.Host)
    fmt.Fprintf(w, "<br> Live check <a href=\"http://%s/live\">link</a> - 200 while not exceed Saturation limit", r.Host)
    fmt.Fprintf(w, "<br> Always error 500 <a href=\"http://%s/error500\">link</a>", r.Host)
	})

  // this function will always return 500 for test purposes
  http.HandleFunc("/error500", func(w http.ResponseWriter, r *http.Request){
    // Record an event happening
    counter.Incr(1)
    w.WriteHeader(http.StatusInternalServerError)
    w.Write([]byte("500 - Something bad happened!"))
  })

  // main worker function with timeout for Latency
  http.Handle("/worker", phm.WrapHandler("myWorker", myWorkerHandler))

  // check if our app able to handle the requests
  http.HandleFunc("/ready", func(w http.ResponseWriter, r *http.Request){
    // Record an event happening
    counter.Incr(1)
    if counter.Rate() <= int64(myAppState.Rate) {
      w.WriteHeader(http.StatusOK)
      w.Write([]byte("Ok!"))
    } else {
      w.WriteHeader(http.StatusInternalServerError)
      w.Write([]byte("500 - rate limit exceed!"))
    }
  })

  // Check if our app is live
  http.HandleFunc("/live", func(w http.ResponseWriter, r *http.Request){
    // Record an event happening
    counter.Incr(1)
    if counter.Rate() <= int64(myAppState.Saturation) {
      w.WriteHeader(http.StatusOK)
      w.Write([]byte("Live!"))
    } else {
      w.WriteHeader(http.StatusInternalServerError)
      w.Write([]byte("500 - Seturation limit exceed!"))
    }
  })

  port := os.Getenv("LISTENING_PORT")

  if port == "" {
    port = "8080"
  }

  log.Printf("listening on port:%s", port)

	err := http.ListenAndServe(":"+port, nil)
	if err != nil {
		log.Fatalf("Failed to start server:%v", err)
	}
}
