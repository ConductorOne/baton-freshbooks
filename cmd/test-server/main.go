// Command test-server is an in-memory mock of the FreshBooks API used to
// exercise the baton-freshbooks connector in CI when no real-tenant
// credentials are available. It replicates the auth flow, the identity and
// team-member endpoints, and the pagination contract documented at
// https://www.freshbooks.com/api.
package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

const (
	listenAddr = ":8765"

	// Hardcoded test credentials — never real secrets.
	testAccessToken  = "test-access-token"
	testClientID     = "test-client-id"
	testClientSecret = "test-client-secret"
	testRefreshToken = "test-refresh-token"

	defaultPerPage = 50
	maxPerPage     = 100 // FreshBooks caps list results at 100 per page.
)

type server struct {
	state *State
}

func main() {
	srv := &server{state: NewState()}

	mux := http.NewServeMux()

	// Doc URL: https://www.freshbooks.com/api/authentication
	// OAuth2 token endpoint (refresh_token grant). Connector token URL is
	// baseURL + "/oauth/token".
	mux.HandleFunc("POST /oauth/token", srv.handleToken)

	// Doc URL: https://www.freshbooks.com/api/me_endpoint
	// Identity endpoint the connector calls to resolve the business ID.
	mux.HandleFunc("GET /api/v1/users/me", srv.requireAuth(srv.handleMe))

	// Doc URL: https://www.freshbooks.com/api/team_members
	// Paginated team-member list (page / per_page query params).
	mux.HandleFunc("GET /api/v1/businesses/{businessID}/team_members", srv.requireAuth(srv.handleTeamMembers))

	httpSrv := &http.Server{
		Addr:              listenAddr,
		Handler:           mux,
		ReadHeaderTimeout: 30 * time.Second,
	}

	log.Printf("test-server listening on %s", listenAddr)
	if err := httpSrv.ListenAndServe(); err != nil {
		log.Println(err)
		os.Exit(1)
	}
}

// requireAuth rejects requests without the expected bearer token, mirroring
// the real API's 401 on missing/invalid credentials.
func (s *server) requireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer "+testAccessToken {
			writeError(w, http.StatusUnauthorized, "unauthenticated", "missing or invalid bearer token")
			return
		}
		next(w, r)
	}
}

// handleToken implements the OAuth2 refresh_token grant. It rejects the same
// inputs the real token endpoint rejects (RFC 6749).
func (s *server) handleToken(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20) // cap token request body at 1 MiB
	if err := r.ParseForm(); err != nil {
		writeOAuthError(w, http.StatusBadRequest, "invalid_request", "could not parse form body")
		return
	}

	if gt := r.PostForm.Get("grant_type"); gt != "refresh_token" {
		writeOAuthError(w, http.StatusBadRequest, "unsupported_grant_type", "only refresh_token is supported")
		return
	}

	// The golang.org/x/oauth2 client may send client credentials via HTTP
	// Basic auth or in the form body; accept either.
	clientID, clientSecret, ok := r.BasicAuth()
	if !ok {
		clientID = r.PostForm.Get("client_id")
		clientSecret = r.PostForm.Get("client_secret")
	}
	if clientID != testClientID || clientSecret != testClientSecret {
		writeOAuthError(w, http.StatusUnauthorized, "invalid_client", "unknown client credentials")
		return
	}

	if rt := r.PostForm.Get("refresh_token"); rt != testRefreshToken {
		writeOAuthError(w, http.StatusBadRequest, "invalid_grant", "unknown refresh token")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"access_token":  testAccessToken,
		"token_type":    "Bearer",
		"expires_in":    3600,
		"refresh_token": testRefreshToken,
	})
}

// handleMe returns the identity payload the connector reads to find the
// business ID (BusinessMemberships[0].Business.ID).
func (s *server) handleMe(w http.ResponseWriter, _ *http.Request) {
	b := s.state.Business()
	writeJSON(w, http.StatusOK, meResponse{
		Response: meUser{
			ID:           1001,
			IdentityID:   2001,
			IdentityUUID: "identity-uuid-2001",
			BusinessMemberships: []businessMembership{
				{ID: 3001, Business: b},
			},
		},
	})
}

// handleTeamMembers serves a paginated team-member list. The connector derives
// the next page from meta (page*per_page < total), so meta must be accurate.
func (s *server) handleTeamMembers(w http.ResponseWriter, r *http.Request) {
	businessID := r.PathValue("businessID")
	if businessID != strconv.FormatInt(s.state.Business().ID, 10) {
		writeError(w, http.StatusNotFound, "not_found", "unknown business id")
		return
	}

	page := parseIntDefault(r.URL.Query().Get("page"), 1)
	perPage := parseIntDefault(r.URL.Query().Get("per_page"), defaultPerPage)
	if perPage <= 0 || perPage > maxPerPage {
		perPage = maxPerPage
	}

	members, total := s.state.ListTeamMembers(page, perPage)
	writeJSON(w, http.StatusOK, teamMembersResponse{
		Response: members,
		Meta:     meta{Page: page, PerPage: perPage, Total: total},
	})
}

func parseIntDefault(s string, def int) int {
	if s == "" {
		return def
	}
	n, err := strconv.Atoi(s)
	if err != nil {
		return def
	}
	return n
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

func writeError(w http.ResponseWriter, status int, code, message string) {
	writeJSON(w, status, map[string]any{"error": code, "message": message})
}

func writeOAuthError(w http.ResponseWriter, status int, code, desc string) {
	writeJSON(w, status, map[string]any{"error": code, "error_description": desc})
}
