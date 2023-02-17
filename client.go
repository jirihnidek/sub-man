package main

import (
	"os"
	"path/filepath"
	"strconv"
)

// RHSMClient contains information about client. It can hold up to 3 different
// type of connections, but usually it is necessary to use only ConsumerCertAuthConnection.
// The NoAuthConnection is used only during registration process, when no consumer
// certificate/key is installed. Note: we do not create special connection for
// "Base Auth", because it is actually NoAuthConnection with special HTTP header.
// EntitlementCertAuthConnection could be used for communication with CDN.
type RHSMClient struct {
	RHSMConf                      *RHSMConf
	NoAuthConnection              *RHSMConnection
	ConsumerCertAuthConnection    *RHSMConnection
	EntitlementCertAuthConnection *RHSMConnection
}

// createRHSMClient tries to create structure holding information
func createRHSMClient(confFilePath *string) error {
	// Try to load configuration file
	rhsmConf, err := loadRHSMConf(confFilePath)
	if err != nil {
		return err
	}

	rhsmClient = &RHSMClient{
		RHSMConf:                      rhsmConf,
		NoAuthConnection:              nil,
		ConsumerCertAuthConnection:    nil,
		EntitlementCertAuthConnection: nil,
	}

	// Try to create connection without authentication
	// Note: It doesn't do any TCP/TLS handshake
	noAuthConnection, err := createNoAuthConnection(
		&rhsmConf.Server.Hostname,
		&rhsmConf.Server.Port,
		&rhsmConf.Server.Prefix)
	if err != nil {
		return err
	}
	rhsmClient.NoAuthConnection = noAuthConnection

	// When consumer key and certificate exist, then it is possible
	// to create connection using consumer cert/key for authentication
	var consumerCertAuthConnection *RHSMConnection = nil
	certFilePath := filepath.Join(rhsmConf.RHSM.ConsumerCertDir, "cert.pem")
	if _, err := os.Stat(certFilePath); err == nil {
		keyFilePath := filepath.Join(rhsmConf.RHSM.ConsumerCertDir, "key.pem")
		if _, err := os.Stat(keyFilePath); err == nil {
			consumerCertAuthConnection, err = createCertAuthConnection(
				&rhsmConf.Server.Hostname,
				&rhsmConf.Server.Port,
				&rhsmConf.Server.Prefix,
				&certFilePath,
				&keyFilePath,
			)
		}
	}
	rhsmClient.ConsumerCertAuthConnection = consumerCertAuthConnection

	// TODO: try to create connection using entitlement cert for authentication
	//       It is not necessary ATM

	return nil
}

func __consumerPEMFile(_rhsmClient *RHSMClient, fileName string) *string {
	consumerCerDir := _rhsmClient.RHSMConf.RHSM.ConsumerCertDir
	consumerCertPath := filepath.Join(consumerCerDir, fileName)
	return &consumerCertPath
}

func __entitlementPEMFile(_rhsmClient *RHSMClient, fileName string) *string {
	entCerDir := _rhsmClient.RHSMConf.RHSM.EntitlementCertDir
	entCertPath := filepath.Join(entCerDir, fileName)
	return &entCertPath
}

// entCertPath tries to return path of entitlement certificate for given serial number
func (rhsmClient *RHSMClient) entCertPath(serialNum int64) *string {
	return __entitlementPEMFile(rhsmClient, strconv.FormatInt(serialNum, 10)+".pem")
}

// entKeyPath tries to return path of entitlement key for given serial number
func (rhsmClient *RHSMClient) entKeyPath(serialNum int64) *string {
	return __entitlementPEMFile(rhsmClient, strconv.FormatInt(serialNum, 10)+"-key.pem")
}

// consumerCertPath tries to return path of consumer certificate
func (rhsmClient *RHSMClient) consumerCertPath() *string {
	return __consumerPEMFile(rhsmClient, "cert.pem")
}

// consumerCertPath tries to return path of consumer certificate
func (rhsmClient *RHSMClient) consumerKeyPath() *string {
	return __consumerPEMFile(rhsmClient, "key.pem")
}
