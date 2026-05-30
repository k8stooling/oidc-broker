package main

import (
	"crypto/rsa"
	"time"
)

type Config struct {
	Addr           string
	IssuerURL      string
	ClientID       string
	KID            string
	PrivateKeyFile string
	TokenTTL       time.Duration
}
type TokenRequest struct {
	Sub    string   `json:"sub"`
	Upn    string   `json:"upn"`
	Groups []string `json:"groups"`
	Aud    string   `json:"aud,omitempty"`
	TTL    int      `json:"ttl_seconds,omitempty"`
}

type Server struct {
	cfg        Config
	privateKey *rsa.PrivateKey
	publicKey  *rsa.PublicKey
}

type OpenIDConfiguration struct {
	Issuer                           string   `json:"issuer"`
	JWKSURI                          string   `json:"jwks_uri"`
	ResponseTypesSupported           []string `json:"response_types_supported"`
	SubjectTypesSupported            []string `json:"subject_types_supported"`
	IDTokenSigningAlgValuesSupported []string `json:"id_token_signing_alg_values_supported"`
	ClaimsSupported                  []string `json:"claims_supported,omitempty"`
}

type JWK struct {
	Kty string `json:"kty"`
	Use string `json:"use,omitempty"`
	Kid string `json:"kid"`
	Alg string `json:"alg,omitempty"`
	N   string `json:"n"`
	E   string `json:"e"`
}

type JWKS struct {
	Keys []JWK `json:"keys"`
}
