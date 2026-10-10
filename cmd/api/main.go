package main

import (
	"fmt"
	"net/http"

	"github.com/akhand0ps/relay/internal/api"
	"github.com/akhand0ps/relay/internal/store"
)




func main(){
	sharedStore := store.NewMemoryStore()
    h:= api.NewHandler(sharedStore)

	mux := http.NewServeMux()



	mux.HandleFunc("/health",api.HealthCheckHandler)
	mux.HandleFunc("POST /v1/events",h.IngestEventHandler)
	mux.HandleFunc("GET  /v1/events/{id}",h.GetEventHandler)



	port := ":3000"
	fmt.Printf(`Relay Servr running on %s\n`,port)

	if err:= http.ListenAndServe(port,mux); err != nil {

		fmt.Printf("Server failed: %v\n",err)
	}
}