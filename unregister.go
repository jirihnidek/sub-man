package main

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
)

func unregister() error {
	consumerCertFile := rhsmClient.consumerCertPath()
	consumerKeyFile := rhsmClient.consumerKeyPath()

	uuid, err := getConsumerUUID(consumerCertFile)

	if err != nil {
		return err
	}

	res, err := rhsmClient.ConsumerCertAuthConnection.request(
		http.MethodDelete,
		"consumers/"+*uuid,
		"",
		"",
		nil,
		nil)
	if err != nil {
		return fmt.Errorf("unable to unregister system: %s", err)
	}

	//_, err = getResponseBody(res)
	//if err != nil {
	//	return err
	//}

	// TODO: handle unusual state in better way
	if res.Status != "204" {
		fmt.Printf("System unregistered\n")
	}

	err = os.Remove(*consumerCertFile)
	if err != nil {
		return fmt.Errorf("unable to remove consumer certificate: %s", err)
	}

	err = os.Remove(*consumerKeyFile)
	if err != nil {
		return fmt.Errorf("unable to remove consumer key: %s", err)
	}

	// Remove entitlement certificate(s)
	entCertDir := &rhsmClient.RHSMConf.RHSM.EntitlementCertDir
	pemFiles, err := os.ReadDir(*entCertDir)
	if err != nil {
		return fmt.Errorf("unable to read directory: %s", *entCertDir)
	}

	for _, pemFile := range pemFiles {
		pemFilePath := filepath.Join(*entCertDir, pemFile.Name())
		_ = os.Remove(pemFilePath)
		// TODO: log that it was not possible to remove file
	}

	// TODO: remove definitions of repositories from redhat.repo

	return nil
}
