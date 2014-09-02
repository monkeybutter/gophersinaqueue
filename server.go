package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"
)

type Job struct {
	Name  string
	Delay int
}

// Job queue length (channel capacity)
const jobQueueLength = 20

// Gophers queue length (channel capacity)
const gopherQueueLength = 4

// Channel of jobs (jobs queue)
var JobsQueue = make(chan Job, jobQueueLength)

// Channel of gophers (workers queue)
var GophersQueue = make(chan bool, gopherQueueLength)

func jobListener(w http.ResponseWriter, r *http.Request) {

	// Send a new job to the queue (send button in UI)
	if r.Method == "POST" {

		var job Job

		if jsonString, err := ioutil.ReadAll(r.Body); err == nil {
			json.Unmarshal(jsonString, &job)
			JobsQueue <- job
			// Channel length / Channel capacity in %
			w.Write([]byte(strconv.Itoa(int(float64(len(JobsQueue)) / float64(cap(JobsQueue)) * 100.0))))
		}
		// Get queue state (refresh button in UI)
	} else if r.Method == "GET" {
		// Channel length / Channel capacity in %
		w.Write([]byte(strconv.Itoa(int(float64(len(JobsQueue)) / float64(cap(JobsQueue)) * 100.0))))
	}
}

// Very long and complicated process
func process(i int) {
	time.Sleep(time.Duration(i) * time.Second)
	// One of our gophers is exhausted but ready for more
	GophersQueue <- true
}

func main() {

	// Queues management is started in an independent routine
	go func() {
		// The gophers queue is filled with hungry gophers
		for i := 0; i < gopherQueueLength; i++ {
			GophersQueue <- true
		}

		// Infinite loop for distributing incoming jobs to idle gophers
		for {
			select {
			case _ = <-GophersQueue:
				go func() {
					select {
					case job := <-JobsQueue:
						process(job.Delay)
					}
				}()
			}
		}
	}()

	r := mux.NewRouter()
	r.Handle("/", http.RedirectHandler("/web_server/index.html", 302))
	r.HandleFunc("/job/", jobListener).Methods("GET", "POST")
	r.PathPrefix("/web_server/").Handler(http.StripPrefix("/web_server", http.FileServer(http.Dir("./static/"))))
	http.Handle("/", r)

	fmt.Println("Set, Ready, Go!")
	panic(http.ListenAndServe(":8000", nil))
}
