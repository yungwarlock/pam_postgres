package requestaccess

import "net/http"

func (s *RequestAccessService) SetupRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /databases", s.GetAllDatabases)
	mux.HandleFunc("POST /request-access", s.CreateAccessRequest)
	mux.HandleFunc("GET /request-access", s.GetAllAccessRequests)
	mux.HandleFunc("/request-access/{requestID}/sse", s.WaitForUpdate)
	mux.HandleFunc("POST /request-access/{requestID}", s.UpdateAccessRequestStatus)
}
