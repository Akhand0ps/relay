package main

import (
	"fmt"
	"net/http"

	"github.com/akhand0ps/relay/internal/api"
	"github.com/akhand0ps/relay/internal/store"
	"github.com/akhand0ps/relay/internal/worker"
)




func main(){
	// sharedStore := store.NewMemoryStore()

	connStr := "postgres://postgres:pgpassword@localhost:5432/relay?sslmode=disable"

	sharedStore,err := store.NewPostgresStore(connStr)

	if err != nil{

	 fmt.Printf("Database connection failed: %v\n",err)
	 return
	}

	

	wp := worker.NewWorkerPool(100,sharedStore) //buffer of 100 tasks

	wp.Start(3) // 3 workers in background

    h:= api.NewHandler(sharedStore,wp) //passing the workerpool

	mux := http.NewServeMux()



	mux.HandleFunc("/health",api.HealthCheckHandler)
	mux.HandleFunc("POST /v1/events",h.IngestEventHandler)
	mux.HandleFunc("GET  /v1/events/{id}",h.GetEventHandler)
	mux.HandleFunc("POST /v1/endpoints",h.CreateEndpointHandler)
	mux.HandleFunc("GET  /v1/endpoints",h.ListEndpointHandler)

	mux.HandleFunc("POST /v1/events/replay",h.ReplayHandler)



	port := ":3000"
	fmt.Printf(`Relay Servr running on %s\n`,port)

	if err:= http.ListenAndServe(port,mux); err != nil {

		fmt.Printf("Server failed: %v\n",err)
	}
}