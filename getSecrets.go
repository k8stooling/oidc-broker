package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func getCurrentNamespace() (string, error) {
	// Kubernetes usually mounts the namespace here inside the pod
	const nsPath = "/var/run/secrets/kubernetes.io/serviceaccount/namespace"

	data, err := os.ReadFile(nsPath)
	if err != nil {
		return "", fmt.Errorf("failed to read namespace file: %w", err)
	}

	namespace := strings.TrimSpace(string(data))
	if namespace == "" {
		return "", fmt.Errorf("namespace file is empty")
	}

	return namespace, nil
}

func getSecretData() string {

	// detect current namespace dynamically
	namespace, err := getCurrentNamespace()
	if err != nil {
		panic(fmt.Errorf("failed to detect namespace: %w", err))
	}

	// create in-cluster config
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(fmt.Errorf("failed to create in-cluster config: %w", err))
	}

	// create Kubernetes client
	clientSet, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(fmt.Errorf("failed to create clientSet: %w", err))
	}

	// retrieve the secret from the detected namespace
	secret, err := clientSet.CoreV1().
		Secrets(namespace).
		Get(context.TODO(), secretName, metav1.GetOptions{})
	if err != nil {
		panic(fmt.Errorf("failed to get secret '%s' in namespace '%s': %w", secretName, namespace, err))
	}

	// extract the desired key
	value, exists := secret.Data[secretKey]
	if !exists {
		panic(fmt.Errorf("key '%s' not found in secret '%s'", secretKey, secretName))
	}

	return string(value)
}
