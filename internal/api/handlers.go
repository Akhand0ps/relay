package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/akhand0ps/relay/internal/id"
	"github.com/akhand0ps/relay/internal/models"
	"github.com/akhand0ps/relay/internal/store"
)


type Handler struct {

	store *store.MemoryStore //lowecase s-> means it is unexported
}

// to make the store exported, provide a constructor function

func NewHandler(s *store.MemoryStore) *Handler{

	return &Handler{store: s}

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

	var req struct {

		Type string   			`json:"type"`
		Payload json.RawMessage `json:"data"`
	}


	// decode incoming JOSN BODY

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {

		http.Error(w,"Invalid JSON payload",http.StatusBadRequest)
		return
	}

	//validation 

	if req.Type == ""{
		http.Error(w,"Event type is required",http.StatusBadRequest)
		return
	}

	//create event

	event := models.Event{
		ID: id.GenerateEventId("evt"),
		Type: req.Type,
		Payload:  req.Payload,
		Status: models.StatusQueued,
		CreatedAt: time.Now(),
	}

	//202 return

	if err := h.store.Save(event); err != nil {
		http.Error(w,"Failed to save event",http.StatusInternalServerError)
		return
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