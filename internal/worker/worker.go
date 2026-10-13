package worker

import (
	"bytes"
	"fmt"
	"net/http"
	"time"

	"github.com/akhand0ps/relay/internal/models"
)






type DeliveryTask struct {

	Delivery models.Delivery
	Endpoint models.Endpoint
	Event    models.Event
}

//workerpool or dispatcher

type WorkerPool struct {

	Queue chan DeliveryTask
	client *http.Client
}

func NewWorkerPool(bufferSize int) *WorkerPool{
	
	return &WorkerPool{
		Queue: make(chan DeliveryTask,bufferSize),
		client: &http.Client{
			Timeout: 10*time.Second,
		},
	}
}

//spaws N workers goroutines in the background

func (wp *WorkerPool) Start(workerCount int){

	for i :=1; i<= workerCount;i++{

		go func(workerId int){

			for task := range wp.Queue{

				wp.processTask(workerId,task)
			}
		}(i)
	}
}
func (wp *WorkerPool) processTask(workerId int,task DeliveryTask){

	fmt.Printf("[Worker %d] Delivering %s to %s...\n",workerId,task.Event.Type,task.Endpoint.URL)

	reqBody := bytes.NewReader(task.Event.Payload)

	resp,err := wp.client.Post(task.Endpoint.URL,"application/json",reqBody)


	statusCode :=0

	if resp!=nil{

		statusCode = resp.StatusCode

		defer resp.Body.Close()

	}

	if err == nil && statusCode >= 200 && statusCode < 300{

		fmt.Printf("[Worker %d] ✔ Delivery SUCCESS to %s (Status: %d)\n",workerId,task.Endpoint.URL, statusCode)
	}else{



		// fmt.Printf("[Worker %d] ⚠ Delivery returned error status %d from %s\n",workerId,statusCode,task.Endpoint.URL)

		try := ShouldRetry(statusCode, err)

		if !try{
			task.Delivery.Status = models.DeliveryStatusFailed
			fmt.Printf("[Worker %d] ❌ Permanent failure, not retrying",workerId)
			return
		}

		task.Delivery.AttemptCount++

		if task.Delivery.AttemptCount >= 5 {

			task.Delivery.Status = models.DeliveryStatusDLQ
			fmt.Printf("[worker %d] is addeed to DLQ",workerId)
			return
		}

		delay := CalculateBackoff(task.Delivery.AttemptCount)

		// wait for the backoff delay in the background,
		// then push back to queue!!

		go func (t DeliveryTask, d time.Duration){

			fmt.Printf("[worker %d] Scheduling retry #%d for %s in %v...\n", workerId,t.Delivery.AttemptCount,t.Endpoint.URL,d)
			time.Sleep(d)
			wp.Queue <- t
		}(task,delay)
		
	}


}
