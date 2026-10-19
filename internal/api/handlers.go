package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/akhand0ps/relay/internal/id"
	"github.com/akhand0ps/relay/internal/models"
	"github.com/akhand0ps/relay/internal/store"
	"github.com/akhand0ps/relay/internal/worker"
)


type Handler struct {

	//store *store.MemoryStore //lowecase s-> means it is unexported
	store store.Store
	workerPool *worker.WorkerPool
}

// to make the store exported, provide a constructor function

func NewHandler(s *store.MemoryStore,wp *worker.WorkerPool) *Handler{

	return &Handler{store: s,workerPool: wp}

}


func HealthCheckHandler(w http.ResponseWriter,r * http.Request) {
	w.Header().Set("Content-Type","applicatio/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ok"}`))
}

func (h * Handler) IngestEventHandler(w http.ResponseWriter,r *http.Request){


	if r.Method != http.MethodPost {

		http.Error(w,"Method not allowed",http.StatusMethodNotAllowed)
		return
	}

	idempontencyKey := r.Header.Get("Idempotency-Key")

	if idempontencyKey != ""{

		//check if we've already seen this event before

		if existingEvent,exists := h.store.GetByIdempotencyKey(idempontencyKey);exists {

			w.Header().Set("Content-Type","application/json")
			w.WriteHeader(http.StatusOK) //return 200Ok as it is duplicate

			json.NewEncoder(w).Encode(map[string]interface{}{
				"id":existingEvent.ID,
				"status":existingEvent.Status,
				"message":"Duplicate event detected,returning existing record",
			})

			return //finlayy return the duplicate event;
		}
	}

	var req struct {

		Type string   			`json:"type"`
		Payload json.RawMessage  `json:"data"`
	}


	// decode incoming JSON BODY

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {

		http.Error(w,"Invalid JSON payload",http.StatusBadRequest)
		return
	}

	//validation 

	if req.Type == "" || len(req.Payload) == 0{
		http.Error(w,"Event type and payload are required",http.StatusBadRequest)
		return
	}

	//create event


	event := models.Event{
		ID: id.GenerateEventId("evt"),
		Type: req.Type,
		Payload:  req.Payload,
		IdempotencyKey: idempontencyKey,
		Status: models.StatusQueued,
		CreatedAt: time.Now(),
	}

	//202 return

	if err := h.store.Save(event); err != nil {
		http.Error(w,"Failed to save event",http.StatusInternalServerError)
		return
	}


	endpoints,err := h.store.FindEndPointForEvent(event.Type)

	if err != nil{

		http.Error(w,"Failed to finda endpoints",http.StatusInternalServerError)
		return
	}

	for _,ep := range endpoints {

		
		delivery := models.Delivery{
			ID: id.GenerateEventId("del"),
			EventId: event.ID,
			EndpointId: ep.ID,
			Status: models.DeliveryStatusQueued,
			CreatedAt: time.Now(),
		}

		h.workerPool.Queue <- worker.DeliveryTask{
			Delivery: delivery,
			Endpoint: ep,
			Event: event,
		}
	}


	w.Header().Set("Content-Type","application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":event.ID,
		"status":event.Status,
		"message":"Event accepted for async processing",
	})
}

func (h * Handler) GetEventHandler(w http.ResponseWriter,r * http.Request){
	eventId:= r.PathValue("id")
 
	if eventId == "" {
		http.Error(w,"Event id mssing from url",http.StatusBadRequest)
		return 
	}

	event,err := h.store.Get(eventId)

	if err != nil {
		http.Error(w,"Event not found",http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type","application/json")
	w.WriteHeader(http.StatusOK)
	
	json.NewEncoder(w).Encode(&event)
}

func (h * Handler) CreateEndpointHandler(w http.ResponseWriter,r * http.Request){


	if r.Method != http.MethodPost {
		http.Error(w,"Method not allowed",http.StatusMethodNotAllowed)
		return 
	}

	var req struct {
		URL string  `json:"url"`
		Event_types []string `json:"event_types"`
	}

	if err:= json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w,"Invalid JSON payload",http.StatusBadRequest)
		return
	}

	if req.URL == "" {
		http.Error(w,"URL is required",http.StatusBadRequest)
		return
	}
	if len(req.Event_types) == 0 {
		http.Error(w,"Event types is required",http.StatusBadRequest)
		return
	}


	
	endpoint := models.Endpoint{
		ID: id.GenerateEventId("ep"),
		URL:req.URL,
		EventTypes: req.Event_types,
		Secret: id.GenerateEventId("whsec_"),
		Status: models.EndpointStatusActive,
		CreatedAt: time.Now(),
	}

	if err := h.store.SaveEndpoint(endpoint); err != nil {
		http.Error(w,"Failed to save endpoint",http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type","application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(endpoint)
}

func (h * Handler) ListEndpointHandler(w http.ResponseWriter,r * http.Request){

	if r.Method != http.MethodGet {
		http.Error(w,"Method not allowed",http.StatusMethodNotAllowed)
		return 
	}

	endpoints,err := h.store.ListEndpoint()
	if err != nil {
		http.Error(w,"Failed to list endpoints",http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type","application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(endpoints)
}

func (h * Handler) ReplayHandler(w http.ResponseWriter, r *http.Request){

  //i will get event id from url


	if r.Method != http.MethodPost {
		http.Error(w,"Method not allowed",http.StatusMethodNotAllowed)
		return 
	}

	var req struct {

		Event_ids []string `json:"event_ids"`
	}
    
	if err:= json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w,"Invalid request body",http.StatusBadRequest)
		return
	}
	

	if len(req.Event_ids) == 0 {
		http.Error(w,"EventIds are required",http.StatusBadRequest)
		return
	}

	//find the the types of the each events and create the delivery and add it in the queue


	for _,eventId := range req.Event_ids {

		event,err := h.store.Get(eventId)
		
		if err != nil{

			http.Error(w,"Event not found",http.StatusNotFound)
			return
		}
		//find the required endpoints of that event

		endpoints,err := h.store.FindEndPointForEvent(event.Type)
		
		if err != nil{
			http.Error(w,"Endpoint not ffound",http.StatusNotFound)
			return
		}

		//create the delivery and add it in the queue
		
		for _,endpoint := range endpoints {

			delivery := models.Delivery {
				ID: id.GenerateEventId("del"),
				EventId: event.ID,
				EndpointId: endpoint.ID,
				Status: models.DeliveryStatusQueued,
				CreatedAt: time.Now(),
			}

			h.workerPool.Queue <- worker.DeliveryTask{
				Delivery: delivery,
				Endpoint: endpoint,
				Event: *event,
			}
		}			
	}
	
	w.Header().Set("Content-Type","application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":"success",
		"message":"Events replayed successfully",
	})
}