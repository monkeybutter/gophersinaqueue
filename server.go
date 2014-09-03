package gophersinaqueue

import (
	"encoding/json"
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
const JobQueueLength = 20

// Gophers ofice size (channel capacity)
const GophersOfficeSize = 4

// Jobs queue (channel of jobs)
var JobsQueue = make(chan Job, JobQueueLength)

// Gophers office (channel of whatever)
var GophersOffice = make(chan bool, GophersOfficeSize)

func jobListener(w http.ResponseWriter, r *http.Request) {

	// Send a new job to the queue (send button in UI)
	if r.Method == "POST" {

		var job Job

		if jsonString, err := ioutil.ReadAll(r.Body); err == nil {
			json.Unmarshal(jsonString, &job)
			// The received job is added to the job queue
			JobsQueue <- job
			// Channel length / Channel capacity in %
			w.Write([]byte(strconv.Itoa(int(float64(len(JobsQueue)) / float64(cap(JobsQueue)) * 100.0))))
		}

	} else if r.Method == "GET" {
		// Channel length / Channel capacity in %
		w.Write([]byte(strconv.Itoa(int(float64(len(JobsQueue)) / float64(cap(JobsQueue)) * 100.0))))
	}
}

// Very long and complicated process
func process(i int) {
	time.Sleep(time.Duration(i) * time.Second)
	// Gopher is exhausted and leaves the office
	<-GophersOffice
}

func init() {
	// Queues management is started in an independent routine
	go func() {
		// Infinite loop for distributing incoming jobs to idle gophers
		for {
			// Gopher gets into the office
			GophersOffice <- true
			go func() {
				select {
				case job := <-JobsQueue:
					process(job.Delay)
				}
			}()
		}
	}()

	http.HandleFunc("/job/", handler)
}

func handler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {

	// Send a new job to the queue (send button in UI)
	case "POST":
		var job Job

		if jsonString, err := ioutil.ReadAll(r.Body); err == nil {
			json.Unmarshal(jsonString, &job)
			// The received job is added to the job queue
			JobsQueue <- job
			// Channel length / Channel capacity in %
			w.Write([]byte(strconv.Itoa(int(float64(len(JobsQueue)) / float64(cap(JobsQueue)) * 100.0))))
		}

	// Get queue state (refresh button in UI)
	case "GET":
		// Channel length / Channel capacity in %
		w.Write([]byte(strconv.Itoa(int(float64(len(JobsQueue)) / float64(cap(JobsQueue)) * 100.0))))

	}
}
