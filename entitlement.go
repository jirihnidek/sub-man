package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
)

// EntitlementSerialJSON is structure used for un-marshaling of JSON document returned from candlepin server.
// JSON document includes list of this objects
type EntitlementSerialJSON struct {
	Created    interface{} `json:"created"`
	Updated    interface{} `json:"updated"`
	Id         interface{} `json:"id"`
	Serial     int64       `json:"serial"`
	Expiration interface{} `json:"expiration"`
	Revoked    interface{} `json:"revoked"`
}

// EntitlementCertificateJSON is structure used for un-marshaling of JSON returned from candlepin server
// JSON document includes list of this objects
type EntitlementCertificateJSON struct {
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

// getEntitlementSerials tries to get list of entitlement serial numbers for given consumer.
func getEntitlementSerials() (*[]int64, error) {
	var entitlementSerialNumbers []int64
	consumerCertFile := rhsmClient.consumerCertPath()

	uuid, err := getConsumerUUID(consumerCertFile)

	if err != nil {
		return nil, fmt.Errorf("failed to get consumer certificate: %v", err)
	}

	res, err := rhsmClient.ConsumerCertAuthConnection.request(
		http.MethodGet,
		"consumers/"+*uuid+"/certificates/serials",
		"",
		"",
		nil,
		nil)

	if err != nil {
		return nil, fmt.Errorf("getting certificate serials failed: %s", err)
	}

	resBody, err := getResponseBody(res)

	if err != nil {
		return nil, err
	}

	// Define list of structures
	var entitlementSerials []EntitlementSerialJSON
	err = json.Unmarshal([]byte(*resBody), &entitlementSerials)

	if err != nil {
		return nil, err
	}

	for _, item := range entitlementSerials {
		entitlementSerialNumbers = append(entitlementSerialNumbers, item.Serial)
	}

	return &entitlementSerialNumbers, nil
}

// getEntitlementCertificate tries to get entitlement certificate with given serial number.
// When it is possible to get entitlement certificate, then this certificate is written to file.
func getEntitlementCertificate(serialNum int64) error {
	consumerCertFile := rhsmClient.consumerCertPath()

	uuid, err := getConsumerUUID(consumerCertFile)

	if err != nil {
		return fmt.Errorf("failed to get consumer certificate: %v", err)
	}

	res, err := rhsmClient.ConsumerCertAuthConnection.request(
		http.MethodGet,
		"consumers/"+*uuid+"/certificates",
		"serials="+strconv.FormatInt(serialNum, 10),
		"",
		nil,
		nil)

	if err != nil {
		return fmt.Errorf("getting entitlement certificate with serial number: %d failed: %s", serialNum, err)
	}

	resBody, err := getResponseBody(res)

	if err != nil {
		return err
	}

	var entCertificates []EntitlementCertificateJSON
	err = json.Unmarshal([]byte(*resBody), &entCertificates)

	if err != nil {
		return err
	}

	for _, entCert := range entCertificates {
		_ = writeEntitlementCert(&entCert.Cert, entCert.Serial.Serial)
		_ = writeEntitlementKey(&entCert.Key, entCert.Serial.Serial)
	}

	return nil
}

// getEntitlementCertificate tries to get all SCA entitlement certificate(s).
// When it is possible to get entitlement certificate(s), then write these certificate(s)  to file.
func getSCAEntitlementCertificate() error {
	consumerCertFile := rhsmClient.consumerCertPath()

	uuid, err := getConsumerUUID(consumerCertFile)

	if err != nil {
		return fmt.Errorf("failed to get consumer certificate: %v", err)
	}

	res, err := rhsmClient.ConsumerCertAuthConnection.request(
		http.MethodGet,
		"consumers/"+*uuid+"/certificates",
		"",
		"",
		nil,
		nil)

	if err != nil {
		return fmt.Errorf("getting entitlement certificates failed: %s", err)
	}

	resBody, err := getResponseBody(res)
	if err != nil {
		return err
	}

	// Try to get SCA entitlement certificate(s). It should be only one certificate,
	// but it is returned in the list (due to backward compatibility).
	var entCertificates []EntitlementCertificateJSON
	err = json.Unmarshal([]byte(*resBody), &entCertificates)
	if err != nil {
		return err
	}

	var serial int64
	var certContent *string
	var idx = -1

	// Write certificate(s) to file(s)
	for id, entCert := range entCertificates {
		_ = writeEntitlementCert(&entCert.Cert, entCert.Serial.Serial)
		_ = writeEntitlementKey(&entCert.Key, entCert.Serial.Serial)
		serial = entCert.Serial.Serial
		certContent = &entCert.Cert
		idx = id
	}

	if idx == 0 {
		err = generateContentFromEntCert(serial, certContent)
		if err != nil {
			return fmt.Errorf("unable to generate content: %s", err)
		}
	} else {
		if idx > 0 {
			return fmt.Errorf("more than one SCA (%d) entitlement certificates installed", idx+1)
		}
		if idx == -1 {
			return fmt.Errorf("no SCA entitlement certificate installed")
		}
	}

	return nil
}

// writeEntitlementCert tries to write entitlement certificate. It is
// typically /etc/pki/entitlement/<serial_number>.pem
func writeEntitlementCert(entCert *string, serialNum int64) error {
	entCertFile := rhsmClient.entCertPath(serialNum)
	return writePemFile(entCertFile, entCert)
}

// writeEntitlementCert tries to write entitlement certificate. It is
// typically /etc/pki/entitlement/<serial_number>.pem
func writeEntitlementKey(entKey *string, serialNum int64) error {
	entKeyFile := rhsmClient.entKeyPath(serialNum)
	return writePemFile(entKeyFile, entKey)
}
