package main

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
)

// getConsumerUUID tries to get consumer UUID from installed consumer certificate
func getConsumerUUID(consumerCertFileName *string) (*string, error) {
	consumerCert, err := os.ReadFile(*consumerCertFileName)

	if err != nil {
		return nil, fmt.Errorf("failed to read consumer certificate: %v", err)
	}

	block, _ := pem.Decode(consumerCert)
	if block == nil {
		return nil, fmt.Errorf("failed to parse: %s (PEM block containing the public key)", *consumerCertFileName)
	}

	if block.Type == "CERTIFICATE" {
		certificate, err := x509.ParseCertificate(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("failed to parse PEM certificate: %s: %v", *consumerCertFileName, err)
		}

		// Note: Similar method could be used for getting other useful information (e.g. org ID).
		// fmt.Printf("Org ID: %v\n", certificate.Subject.Organization[0])

		return &certificate.Subject.CommonName, nil
	}

	return nil, fmt.Errorf("file %s does not contain CERTIFICATE block", *consumerCertFileName)
}
