package main

import (
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"net/http"
	"strings"
	"time"
)

func (s *Server) handleToken(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed, use POST", http.StatusMethodNotAllowed)
		return
	}

	var req TokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON body: "+err.Error(), http.StatusBadRequest)
		return
	}

	if req.Sub == "" {
		http.Error(w, "sub is required", http.StatusBadRequest)
		return
	}

	aud := s.cfg.ClientID
	if req.Aud != "" {
		aud = req.Aud
	}

	ttl := s.cfg.TokenTTL
	if req.TTL > 0 {
		ttl = time.Duration(req.TTL) * time.Second
	}

	now := time.Now().UTC()

	claims := map[string]interface{}{
		"iss": s.cfg.IssuerURL,
		"sub": req.Sub,
		"aud": aud,
		"iat": now.Unix(),
		"nbf": now.Unix(),
		"exp": now.Add(ttl).Unix(),
		"tid": getEnv("TENANT_ID", ""),
	}

	if req.Upn != "" {
		claims["upn"] = req.Upn
	}

	if len(req.Groups) > 0 {
		claims["groups"] = req.Groups
	}

	token, err := signJWT_RS256(s.privateKey, s.cfg.KID, claims)
	if err != nil {
		http.Error(w, "failed to sign token: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// add custom response headers here
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("X-Token-Issuer", s.cfg.IssuerURL)
	w.Header().Set("X-Token-Kid", s.cfg.KID)
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("X-Oidc-Broker-Version", "1.0.0")

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"token_type": "Bearer",
		"id_token":   token,
		"expires_in": int(ttl.Seconds()),
		"claims":     claims,
	})
}

func (s *Server) handleOpenIDConfiguration(w http.ResponseWriter, r *http.Request) {
	cfg := OpenIDConfiguration{
		Issuer:                           s.cfg.IssuerURL,
		JWKSURI:                          strings.TrimRight(s.cfg.IssuerURL, "/") + "/keys",
		ResponseTypesSupported:           []string{"id_token"},
		SubjectTypesSupported:            []string{"public"},
		IDTokenSigningAlgValuesSupported: []string{"RS256"},
		ClaimsSupported:                  []string{"iss", "sub", "aud", "exp", "iat", "nbf", "groups", "upn", "tid"},
	}
	writeJSON(w, http.StatusOK, cfg)
}

func (s *Server) handleJWKS(w http.ResponseWriter, r *http.Request) {
	jwks := JWKS{
		Keys: []JWK{rsaPublicKeyToJWK(s.publicKey, s.cfg.KID)},
	}
	writeJSON(w, http.StatusOK, jwks)
}

func (s *Server) handlePubKey(w http.ResponseWriter, r *http.Request) {
	pubASN1, err := x509.MarshalPKIXPublicKey(s.publicKey)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	block := &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: pubASN1,
	}

	w.Header().Set("Content-Type", "application/x-pem-file")
	w.WriteHeader(http.StatusOK)
	w.Write(pem.EncodeToMemory(block))
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
