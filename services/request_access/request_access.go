package requestaccess

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	dbmanager "pam_postgres/services/db_manager"
	"time"
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
		authDetails := &dbmanager.PostgresAuthDetails{
			Host:     s.model.host,
			Port:     s.model.port,
			User:     s.model.rootUser,
			Password: s.model.rootPassword,
		}
		tempUserAuth, err := dbmanager.GenerateTempUserWithPermissions(ctx, authDetails, &accessRequest.Permissions)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to generate temporary user: %v", err), http.StatusInternalServerError)
			return
		}

		accessRequest.AuthDetails = *tempUserAuth
		if err := s.model.UpdateAccessRequestWithTempUser(ctx, requestID, tempUserAuth); err != nil {
			http.Error(w, fmt.Sprintf("Failed to update access request with temp user: %v", err), http.StatusInternalServerError)
			return
		}
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

func (s *RequestAccessService) WaitForUpdate(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "text/event-stream")

	ctx := r.Context()
	requestID := r.PathValue("requestID")

	// rc := http.NewResponseController(w)
	t := time.NewTicker(time.Second)
	defer t.Stop()

	for {
		select {
		case <-ctx.Done():
			fmt.Println("Client disconnected")
			return
		case <-t.C:
			log.Println("Fetching again...")
			accessRequest, err := s.model.GetAccessRequestByID(ctx, requestID)
			if err != nil {
				log.Printf("Error fetching access request ID %s: %v\n", requestID, err)
				return
			}
			if accessRequest.Status != StatusPending {
				log.Printf("Sending response for request ID %s with status %s\n", requestID, accessRequest.Status)
				jsonData, err := json.Marshal(accessRequest)
				if err != nil {
					log.Printf("Error marshaling access request ID %s: %v\n", requestID, err)
					return
				}
				_, err = fmt.Fprintf(w, "data: %s\n\n", jsonData)
				if err != nil {
					log.Printf("Error writing to ResponseWriter for request ID %s: %v\n", requestID, err)
					return
				}
				return
			}
		}
	}
}
