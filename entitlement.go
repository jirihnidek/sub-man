package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
)

// EntitlementCertificateKeyJSON is structure used for un-marshaling of JSON returned from candlepin server
// JSON document includes list of this objects
type EntitlementCertificateKeyJSON struct {
	Created string `json:"created"`
	Updated string `json:"updated"`
	Id      string `json:"id"`
	Key     string `json:"key"`
	Cert    string `json:"cert"`
	Serial  struct {
		Created    string `json:"created"`
		Updated    string `json:"updated"`
		Id         int64  `json:"id"`
		Serial     int64  `json:"serial"`
		Expiration string `json:"expiration"`
		Revoked    bool   `json:"revoked"`
	} `json:"serial"`
}

// getEntitlementCertificate tries to get all SCA entitlement certificate(s) from candlepin server.
// When it is possible to get entitlement certificate(s), then write these certificate(s) to file.
// Note: candlepin server returns only one SCA entitlement certificate ATM, but REST API allows to
// return more entitlement certificates.
func getSCAEntitlementCertificate() ([]EntitlementCertificateKeyJSON, error) {
	consumerCertFile := rhsmClient.consumerCertPath()

	uuid, err := getConsumerUUID(consumerCertFile)

	if err != nil {
		return nil, fmt.Errorf("failed to get consumer certificate: %v", err)
	}

	res, err := rhsmClient.ConsumerCertAuthConnection.request(
		http.MethodGet,
		"consumers/"+*uuid+"/certificates",
		"",
		"",
		nil,
		nil)

	if err != nil {
		return nil, fmt.Errorf("getting entitlement certificates failed: %s", err)
	}

	resBody, err := getResponseBody(res)
	if err != nil {
		return nil, err
	}

	// Try to get SCA entitlement certificate(s). It should be only one certificate,
	// but it is returned in the list (due to backward compatibility).
	var entCertKeys []EntitlementCertificateKeyJSON
	err = json.Unmarshal([]byte(*resBody), &entCertKeys)
	if err != nil {
		return nil, err
	}

	// When one entitlement certificate was returned, then generate redhat.repo from this
	// entitlement certificate
	l := len(entCertKeys)
	if l != 1 {
		if l == 0 {
			return nil, fmt.Errorf("no SCA entitlement certificate returned from server")
		}
		// if l > 0 {} TODO: print warning that more than one entitlement certificate was returned
		// log.Printf("more than one SCA (%d) entitlement certificates installed", l)
	}

	// Write certificate(s) and key(s) to file(s)
	for _, entCertKey := range entCertKeys {
		entCertFilePath, err := writeEntitlementCert(&entCertKey.Cert, entCertKey.Serial.Serial)
		if err != nil {
			// TODO: print error that it was not possible to install entitlement certificate
			// log.Printf("%s", err)
			continue
		}
		_, err = writeEntitlementKey(&entCertKey.Key, entCertKey.Serial.Serial)
		if err != nil {
			log.Printf("unable to write entitlement key: %s", err)

			// When it is not possible to install key, then remove certificate file, because
			// certificate is useless without key
			err = os.Remove(*entCertFilePath)
			if err != nil {
				log.Printf("unable to remove entitlement certificate: %s", err)
			}
		}
	}

	return entCertKeys, nil
}

// writeEntitlementCert tries to write entitlement certificate. It is
// typically /etc/pki/entitlement/<serial_number>.pem
func writeEntitlementCert(entCert *string, serialNum int64) (*string, error) {
	entCertFilePath := rhsmClient.entCertPath(serialNum)
	return entCertFilePath, writePemFile(entCertFilePath, entCert, nil)
}

// writeEntitlementCert tries to write entitlement certificate. It is
// typically /etc/pki/entitlement/<serial_number>-key.pem
func writeEntitlementKey(entKey *string, serialNum int64) (*string, error) {
	entKeyFilePath := rhsmClient.entKeyPath(serialNum)
	return entKeyFilePath, writePemFile(entKeyFilePath, entKey, nil)
}
