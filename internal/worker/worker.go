package worker

import (
	"bytes"
	"fmt"
	"net/http"
	"time"

	"github.com/akhand0ps/relay/internal/id"
	"github.com/akhand0ps/relay/internal/models"
	"github.com/akhand0ps/relay/internal/queue"
	"github.com/akhand0ps/relay/internal/signature"
	"github.com/akhand0ps/relay/internal/store"
)




type WorkerPool struct {

	// Queue chan models.DeliveryTask
	Queue queue.Queue
	client *http.Client
	store store.Store
} 

func NewWorkerPool(bufferSize int,q queue.Queue,store store.Store) *WorkerPool{
	
	return &WorkerPool{
		Queue: q,
		client: &http.Client{
			Timeout: 10*time.Second,
		},
		store: store, 
	}
}

//spaws N workers goroutines in the background

func (wp *WorkerPool) Start(workerCount int){

	for i :=1; i<= workerCount;i++{

		go func(workerId int){

			for task := range wp.Queue {

				wp.processTask(workerId,task)
			}
		}(i)
	}
}
func (wp *WorkerPool) processTask(workerId int,task models.DeliveryTask){

	fmt.Printf("[Worker %d] Delivering %s to %s...\n",workerId,task.Event.Type,task.Endpoint.URL)

	// reqBody := bytes.NewReader(task.Event.Payload)

	// resp,err := wp.client.Post(task.Endpoint.URL,"application/json",reqBody)

	//set custom headers

	now := time.Now().Unix()
	sig := signature.Sign(task.Endpoint.Secret,now,task.Event.Payload)

	req,err := http.NewRequest("POST",task.Endpoint.URL,
		bytes.NewReader(task.Event.Payload))
	
	if err != nil{

		fmt.Printf("[Worker %d] Failed to create request: %v\n",workerId,err)
		return
	}

	req.Header.Set("Content-Type","application/json")
	req.Header.Set("X-Relay-Event-Id",task.Event.ID)
	req.Header.Set("X-Relay-TimeStamp",fmt.Sprintf("%d",now))
	req.Header.Set("X-Relay-Signature",sig)

	resp,err := wp.client.Do(req)
	statusCode :=0

	if resp!=nil{

		statusCode = resp.StatusCode

		defer resp.Body.Close()

	}

	var errMessage string
	if err != nil {
		errMessage = err.Error()
	} else {
		errMessage = ""
	}

	attempt := models.DeliveryAttempt{
		ID: id.GenerateEventId("att"),
		DeliveryId: task.Delivery.ID,
		StatusCode: statusCode,
		ErrorMessage: errMessage,
		AttemptedAt: time.Now(),
	}

	if err :=wp.store.RecordDeliveryAttempt(attempt);err != nil{
		
		fmt.Printf("[Worker %d]Failed to record delivery attempt:%v\n",workerId,err)
		//we might still want to retry or fail even if we can't record the attempt
	}
	//determine if delivery succeeded
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
		go func (t models.DeliveryTask, d time.Duration){
			fmt.Printf("[worker %d] Scheduling retry #%d for %s in %v...\n", workerId,t.Delivery.AttemptCount,t.Endpoint.URL,d)
			time.Sleep(d)
			wp.Queue <- t
		}(task,delay)
		
	}


}
