package requestaccess

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

type RequestAccessService struct {
	model *requestAccessModel
}

func NewRequestAccessService(db *sql.DB) *RequestAccessService {
	m := &requestAccessModel{DB: db}
	if err := m.InitDB(); err != nil {
		log.Println("Failed to initialize transactions model:", err)
		return nil
	}

	return &RequestAccessService{
		model: m,
	}
}

func (s *RequestAccessService) GetAllAccessRequests(w http.ResponseWriter, r *http.Request) {
	accessRequests, err := s.model.GetAllAccessRequests()
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get access requests: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(accessRequests)
}

func (s *RequestAccessService) CreateAccessRequest(w http.ResponseWriter, r *http.Request) {
	var accessRequest AccessRequest
	if err := json.NewDecoder(r.Body).Decode(&accessRequest); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request payload: %v", err), http.StatusBadRequest)
		return
	}

	accessRequest.Status = StatusPending

	if err := s.model.CreateAccessRequest(&accessRequest); err != nil {
		http.Error(w, fmt.Sprintf("Failed to create access request: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(accessRequest)
}

func (s *RequestAccessService) UpdateAccessRequestStatus(w http.ResponseWriter, r *http.Request) {
	requestID := r.PathValue("requestID")
	var statusUpdate struct {
		Status RequestStatus `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&statusUpdate); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request payload: %v", err), http.StatusBadRequest)
		return
	}

	if statusUpdate.Status == StatusApproved {
		approveAccessRequest()
	}

	if err := s.model.UpdateAccessRequestStatus(requestID, statusUpdate.Status); err != nil {
		http.Error(w, fmt.Sprintf("Failed to update access request status: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
