package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

func main() {
	cfg := Config{
		Addr:           getEnv("ADDR", ":"+SERVER_PORT),
		IssuerURL:      mustGetEnv("ISSUER_URL"),
		ClientID:       getEnv("CLIENT_ID", "kubernetes"),
		KID:            getEnv("KID", "oidc-broker-key-1"),
		PrivateKeyFile: getSecretData(),
		TokenTTL:       getEnvDuration("TOKEN_TTL", TOKEN_TTL),
	}

	if !strings.HasPrefix(cfg.IssuerURL, "https://") {
		log.Fatal("ISSUER_URL must use https")
	}

	priv, err := parseRSAPrivateKeyFromString(cfg.PrivateKeyFile)
	if err != nil {
		log.Fatalf("failed to parse RSA private key: %v", err)
	}

	s := &Server{
		cfg:        cfg,
		privateKey: priv,
		publicKey:  &priv.PublicKey,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/.well-known/openid-configuration", methodOnly(http.MethodGet, s.handleOpenIDConfiguration))
	mux.HandleFunc("/keys", methodOnly(http.MethodGet, s.handleJWKS))
	mux.HandleFunc("/token", methodOnly(http.MethodPost, maxBodyBytes(16*1024, s.handleToken)))
	mux.HandleFunc("/healthz", methodOnly(http.MethodGet, s.handleHealth))

	handler := chain(
		mux,
		recoverMiddleware,
		requestIDMiddleware,
		securityHeadersMiddleware,
		loggingMiddleware,
	)

	srv := &http.Server{
		Addr:              cfg.Addr,
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
		MaxHeaderBytes:    8 * 1024,
	}

	log.Printf("starting OIDC issuer on %s", cfg.Addr)
	log.Printf("external issuer URL: %s", cfg.IssuerURL)
	log.Printf("client ID (audience): %s", cfg.ClientID)
	log.Printf("kid: %s", cfg.KID)

	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("server failed: %v", err)
		}
	}()

	// Graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	log.Printf("shutdown signal received")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("graceful shutdown failed: %v", err)
		if err := srv.Close(); err != nil {
			log.Printf("server close failed: %v", err)
		}
	}

	log.Printf("server stopped")
}
