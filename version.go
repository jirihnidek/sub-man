package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

const VERSION = "0.1"

// clientVersion tries to print version of client
func clientVersion() (string, error) {
	// TODO: this should be more sophisticated (version from git tag)
	return VERSION, nil
}

// serverVersion tries to get version of server and version of rules
func serverVersion() (*string, *string, error) {
	var connection *RHSMConnection
	consumerCertFile := rhsmClient.consumerCertPath()

	_, err := getConsumerUUID(consumerCertFile)

	if err != nil {
		connection = rhsmClient.NoAuthConnection
	} else {
		connection = rhsmClient.ConsumerCertAuthConnection
	}

	if connection == nil {
		return nil, nil, fmt.Errorf("unable to establish any connection")
	}

	res, err := connection.request(
		http.MethodGet,
		"status",
		"",
		"",
		nil,
		nil,
	)

	if err != nil {
		return nil, nil, fmt.Errorf("unable to get server status :%v", err)
	}

	resBody, err := getResponseBody(res)

	if err != nil {
		return nil, nil, err
	}

	rhsmStatus := RHSMStatus{}
	err = json.Unmarshal([]byte(*resBody), &rhsmStatus)

	if err != nil {
		return nil, nil, fmt.Errorf("unable to parse server status: %s", err)
	}

	serverVersionRelease := rhsmStatus.Version + rhsmStatus.Release
	return &serverVersionRelease, &rhsmStatus.RulesVersion, nil

}
