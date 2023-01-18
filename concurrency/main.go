package main

import (
	"fmt"
	"math/rand"
	"sort"
	"sync"
	"time"
)

/*******************************************
* JOB
*******************************************/

type Job struct {
	JobID     int
	SleepTime int
	Done      bool
}

// Simulate job execution: sleep as if busy
func (j *Job) Execute() bool {
	if j.Done {
		// Failed to execute, already done.
		return false
	}
	// Do the job: in this case, take a nap
	time.Sleep(time.Duration(j.SleepTime) * time.Second)

	// We did it! Set done and return
	j.Done = true
	return j.Done
}

/******************************************
* MONITOR
******************************************/

type Monitor struct {
	status chan int
	err    chan int
	wg     *sync.WaitGroup
	Log    map[int]int
}

// Make channels for us to communicate across threads
func (m *Monitor) listen() {
	m.status = make(chan int)
	m.err = make(chan int)
	m.Log = make(map[int]int)
	go m.receive()
	go func() {
		for {
			m.report()
		}
	}()
}

// Close comm channels
func (m *Monitor) close() {
	m.receive()
	close(m.status)
	close(m.err)
}

// Publish communications received from channels.
func (m *Monitor) receive() {
	done := false
	for !done {
		// block on signal
		select {
		case signal, ok := <-m.status:
			if ok {
				if signal > 0 {
					// Start signal: encode as 0
					m.Log[signal] = 0
				} else {
					// Complete signal: encode as 1
					m.Log[-signal] = 1
				}
			} else {
				fmt.Println("Report Channel CLOSED")
				done = true
				break
			}
		case signal, ok := <-m.err:
			if ok {
				m.Log[signal] = -1
			} else {
				fmt.Println("Error Channel CLOSED")
				done = true
				break
			}
		}
	}
}

func (m *Monitor) report() {
	// Sort by status
	running := make([]int, 0)
	complete := make([]int, 0)
	failed := make([]int, 0)
	for key, val := range m.Log {
		switch val {
		case -1:
			failed = append(failed, key)
		case 0:
			running = append(running, key)
		case 1:
			complete = append(complete, key)
		}
	}
	sort.Ints(running)
	sort.Ints(complete)
	sort.Ints(failed)
	fmt.Println("--------------------------------------")
	fmt.Println("Running: ", running)
	fmt.Println("Complete: ", complete)
	fmt.Println("Failed: ", failed)
	time.Sleep(time.Duration(2) * time.Second)
}

func (m *Monitor) batchProcess(data []Job) {
	for _, j := range data {
		// Send +j.JobID to indicate start
		m.status <- j.JobID
		success := j.Execute()
		if success {
			// Send -j.JobID to indicate end
			m.status <- (-j.JobID)
		} else {
			m.err <- j.JobID
		}
	}
}

func (m *Monitor) Distribute(q []Job, threads int) {
	// Initialize
	seg := len(q) / threads
	wg := new(sync.WaitGroup)
	m.listen()

	// Distribute & Execute in Batch
	for i := 0; i < threads; i++ {
		first := i * seg
		last := (i + 1) * seg
		segment := q[first:last]
		wg.Add(1)
		go func() {
			defer wg.Done()
			m.batchProcess(segment)
		}()
	}
	wg.Wait()

	// Finish
	m.close()
}

/****************
* MAIN
*****************/

func main() {

	rand.Seed(time.Now().UnixNano())

	// Create 100 jobs that will randomly sleep for 1-10 seconds
	queue := make([]Job, 20)
	for i := range queue {
		queue[i].JobID = i + 1
		queue[i].SleepTime = rand.Intn(10) + 1
	}

	// Create a monitor & distribute the jobs to n threads.
	m := Monitor{}
	n := 4
	m.Distribute(queue, n)

}
