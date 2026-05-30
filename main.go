package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

func main() {
	cfg := Config{
		Addr:           getEnv("ADDR", ":8080"),
		IssuerURL:      mustGetEnv("ISSUER_URL"),
		ClientID:       getEnv("CLIENT_ID", "kubernetes"),
		KID:            getEnv("KID", "oidc-broker-key-1"),
		PrivateKeyFile: getSecretData(),
		TokenTTL:       getEnvDuration("TOKEN_TTL", 15*time.Minute),
	}

	priv, err := parseRSAPrivateKeyFromString(cfg.PrivateKeyFile)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	s := &Server{
		cfg:        cfg,
		privateKey: priv,
		publicKey:  &priv.PublicKey,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/.well-known/openid-configuration", s.handleOpenIDConfiguration)
	mux.HandleFunc("/keys", s.handleJWKS)
	mux.HandleFunc("/token", s.handleToken)
	mux.HandleFunc("/healthz", s.handleHealth)
	mux.HandleFunc("/pubkey", s.handlePubKey)

	log.Printf("starting OIDC issuer on %s (plain HTTP)", cfg.Addr)
	log.Printf("external issuer URL: %s", cfg.IssuerURL)
	log.Printf("client ID (audience): %s", cfg.ClientID)
	log.Printf("kid: %s", cfg.KID)

	if err := http.ListenAndServe(cfg.Addr, loggingMiddleware(mux)); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	_ = enc.Encode(v)
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s from %s took %s", r.Method, r.URL.Path, r.RemoteAddr, time.Since(start))
	})
}
