package requestaccess

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
)

var (
	Debug = os.Getenv("DEBUG") == "1"
)

type RequestAccessService struct {
	model *RequestAccessModel
}

func NewRequestAccessService(model *RequestAccessModel) *RequestAccessService {
	if err := model.InitDB(); err != nil {
		log.Println("Failed to initialize transactions model:", err)
		return nil
	}

	return &RequestAccessService{
		model: model,
	}
}

func (s *RequestAccessService) GetAllAccessRequests(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	accessRequests, err := s.model.GetAllAccessRequests(ctx)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get access requests: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(accessRequests)
}

func (s *RequestAccessService) CreateAccessRequest(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var accessRequest AccessRequest

	if err := json.NewDecoder(r.Body).Decode(&accessRequest); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request payload: %v", err), http.StatusBadRequest)
		return
	}

	accessRequest.Status = StatusPending

	if err := s.model.CreateAccessRequest(ctx, &accessRequest); err != nil {
		http.Error(w, fmt.Sprintf("Failed to create access request: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(accessRequest)
}

func (s *RequestAccessService) UpdateAccessRequestStatus(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	requestID := r.PathValue("requestID")

	var statusUpdate struct {
		Status RequestStatus `json:"status"`
	}

	if err := json.NewDecoder(r.Body).Decode(&statusUpdate); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request payload: %v", err), http.StatusBadRequest)
		return
	}

	accessRequest, err := s.model.GetAccessRequestByID(ctx, requestID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get access requests: %v", err), http.StatusInternalServerError)
		return
	}

	if statusUpdate.Status == StatusApproved {
		s.model.approveAccessRequest(r.Context(), s.model.DB, &accessRequest.Permissions)
	}

	if err := s.model.UpdateAccessRequestStatus(ctx, requestID, statusUpdate.Status); err != nil {
		http.Error(w, fmt.Sprintf("Failed to update access request status: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (s *RequestAccessService) GetAllDatabases(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	databases, err := s.model.GetAllTablesFromAllDatabases(ctx)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get databases: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(databases)
}
