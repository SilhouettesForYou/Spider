package controllers

import (
	"fmt"

	"Spider/color"
)

type QueueScheduler struct {
	requestChannel chan Request
	workerChannel chan chan Request
}

func (this *QueueScheduler) CreateWorkerChannel() chan Request {
	return make(chan Request)
}

func (this *QueueScheduler) Submit(request Request) {
	this.requestChannel <- request
}

func (this *QueueScheduler) WorkerReady(worker chan Request) {
	this.workerChannel <- worker
}

func (this *QueueScheduler) Run() {
	this.workerChannel = make(chan chan Request)
	this.requestChannel = make(chan Request)

	go func() {
		var requestQueue [] Request
		var workerQueue [] chan Request

		for {
			var activeRequest Request
			var activeWorker chan Request
			if len(requestQueue) > 0 && len(workerQueue) > 0 {
				activeWorker = workerQueue[0]
				activeRequest = requestQueue[0]
			}
			select {
			case r := <- this.requestChannel:
				requestQueue = append(requestQueue, r)
				
				color.ColorLog.Logs(
					"                                                            ", color.NormalBold, 
					" REQUEST QUEUE LENGTH ", 				    color.NormalBold, 
					fmt.Sprintf("%3d\r", len(requestQueue)),    color.MagentaBold)
					
			case w:= <- this.workerChannel:
				workerQueue = append(workerQueue, w)

				color.ColorLog.Logs(
					"                                                            ", color.NormalBold, 
					" WORKER QUEUE LENGTH ", 				    color.NormalBold, 
					fmt.Sprintf("%3d\r", len(workerQueue)),	    color.MagentaBold)
			case activeWorker <- activeRequest:
				workerQueue = workerQueue[1:]
				requestQueue = requestQueue[1:]
			}
		}
	}()
}