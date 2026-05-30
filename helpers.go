package main

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"log"
	"math/big"
	"os"
	"strings"
	"time"
)

func rsaPublicKeyToJWK(pub *rsa.PublicKey, kid string) JWK {
	n := base64.RawURLEncoding.EncodeToString(pub.N.Bytes())

	eBig := big.NewInt(int64(pub.E))
	e := base64.RawURLEncoding.EncodeToString(eBig.Bytes())

	return JWK{
		Kty: "RSA",
		Use: "sig",
		Kid: kid,
		Alg: "RS256",
		N:   n,
		E:   e,
	}
}

func signJWT_RS256(priv *rsa.PrivateKey, kid string, claims map[string]interface{}) (string, error) {
	header := map[string]interface{}{
		"alg": "RS256",
		"typ": "JWT",
		"kid": kid,
	}

	headerJSON, err := json.Marshal(header)
	if err != nil {
		return "", fmt.Errorf("marshal header: %w", err)
	}

	claimsJSON, err := json.Marshal(claims)
	if err != nil {
		return "", fmt.Errorf("marshal claims: %w", err)
	}

	encodedHeader := base64.RawURLEncoding.EncodeToString(headerJSON)
	encodedClaims := base64.RawURLEncoding.EncodeToString(claimsJSON)
	signingInput := encodedHeader + "." + encodedClaims

	hash := sha256.Sum256([]byte(signingInput))
	sig, err := rsa.SignPKCS1v15(rand.Reader, priv, crypto.SHA256, hash[:])
	if err != nil {
		return "", fmt.Errorf("sign jwt: %w", err)
	}

	encodedSig := base64.RawURLEncoding.EncodeToString(sig)
	return signingInput + "." + encodedSig, nil
}

func parseRSAPrivateKeyFromString(keyStr string) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode([]byte(keyStr))
	if block == nil {
		return nil, errors.New("failed to decode PEM block")
	}

	// PKCS1 ("BEGIN RSA PRIVATE KEY")
	if block.Type == "RSA PRIVATE KEY" {
		return x509.ParsePKCS1PrivateKey(block.Bytes)
	}

	// PKCS8 ("BEGIN PRIVATE KEY")
	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	rsaKey, ok := key.(*rsa.PrivateKey)
	if !ok {
		return nil, errors.New("not an RSA private key")
	}

	return rsaKey, nil
}

func mustGetEnv(k string) string {
	v := strings.TrimSpace(os.Getenv(k))
	if v == "" {
		log.Fatalf("missing required env var: %s", k)
	}
	return v
}

func getEnv(k, def string) string {
	v := strings.TrimSpace(os.Getenv(k))
	if v == "" {
		return def
	}
	return v
}

func getEnvDuration(k string, def time.Duration) time.Duration {
	v := strings.TrimSpace(os.Getenv(k))
	if v == "" {
		return def
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		return def
	}
	return d
}
