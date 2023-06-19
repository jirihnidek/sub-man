package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// RHSMStatus is structure used for storing GET response from REST API
// endpoint "/status". This endpoint can be called using no-auth or
// consumer-cert-auth connection
type RHSMStatus struct {
	Mode                string      `json:"mode"`
	ModeReason          interface{} `json:"modeReason"`
	ModeChangeTime      interface{} `json:"modeChangeTime"`
	Result              bool        `json:"result"`
	Version             string      `json:"version"`
	Release             string      `json:"release"`
	Standalone          bool        `json:"standalone"`
	TimeUTC             time.Time   `json:"timeUTC"`
	RulesSource         string      `json:"rulesSource"`
	RulesVersion        string      `json:"rulesVersion"`
	ManagerCapabilities []string    `json:"managerCapabilities"`
	KeycloakRealm       interface{} `json:"keycloakRealm"`
	KeycloakAuthUrl     interface{} `json:"keycloakAuthUrl"`
	KeycloakResource    interface{} `json:"keycloakResource"`
	DeviceAuthRealm     interface{} `json:"deviceAuthRealm"`
	DeviceAuthUrl       interface{} `json:"deviceAuthUrl"`
	DeviceAuthClientId  interface{} `json:"deviceAuthClientId"`
	DeviceAuthScope     interface{} `json:"deviceAuthScope"`
}

// RHSMCompliant is structure used for storing GET response from REST API
// endpoint "/consumers/<consumer-UUID>/compliance"
type RHSMCompliant struct {
	Status            string      `json:"status"`
	Compliant         bool        `json:"compliant"`
	Date              string      `json:"date"`
	CompliantUntil    interface{} `json:"compliantUntil"`
	CompliantProducts struct {
	} `json:"compliantProducts"`
	PartiallyCompliantProducts struct {
	} `json:"partiallyCompliantProducts"`
	PartialStacks struct {
	} `json:"partialStacks"`
	NonCompliantProducts        []interface{} `json:"nonCompliantProducts"`
	Reasons                     []interface{} `json:"reasons"`
	ProductComplianceDateRanges struct {
	} `json:"productComplianceDateRanges"`
}

// RHSMSyspurposeCompliant is structure used for storing GET response from REST API
// endpoint "/consumers/<consumer-UUID>/purpose_compliance". This REST API endpoint
// should be called only in the case, when entitlement content access mode is used.
type RHSMSyspurposeCompliant struct {
	Status                  string      `json:"status"`
	Compliant               bool        `json:"compliant"`
	Date                    string      `json:"date"`
	NonCompliantRole        interface{} `json:"nonCompliantRole"`
	NonCompliantSLA         interface{} `json:"nonCompliantSLA"`
	NonCompliantUsage       interface{} `json:"nonCompliantUsage"`
	NonCompliantServiceType interface{} `json:"nonCompliantServiceType"`
	CompliantRole           struct {
	} `json:"compliantRole"`
	CompliantAddOns struct {
	} `json:"compliantAddOns"`
	CompliantSLA struct {
	} `json:"compliantSLA"`
	CompliantUsage struct {
	} `json:"compliantUsage"`
	NonCompliantAddOns   []interface{} `json:"nonCompliantAddOns"`
	CompliantServiceType struct {
	} `json:"compliantServiceType"`
	Reasons []interface{} `json:"reasons"`
}

// status tries to pretty print system status
func status() error {
	consumerCertFile := rhsmClient.consumerCertPath()

	uuid, err := getConsumerUUID(consumerCertFile)

	if err != nil {
		fmt.Printf("Status: unknown\n")
		fmt.Printf("System Purpose Status: unknown\n")
		return nil
	}

	res, err := rhsmClient.ConsumerCertAuthConnection.request(
		http.MethodGet,
		"consumers/"+*uuid+"/compliance",
		"",
		"",
		nil,
		nil)

	if err != nil {
		return fmt.Errorf("getting consumer compliance failed: %s", err)
	}

	resBody, err := getResponseBody(res)

	if err != nil {
		return err
	}

	rhsmCompliant := RHSMCompliant{}
	err = json.Unmarshal([]byte(*resBody), &rhsmCompliant)

	if err != nil {
		return fmt.Errorf("unable to parse compliant status: %s", err)
	}

	fmt.Printf("Status: %s\n", rhsmCompliant.Status)

	if rhsmCompliant.Status == "disabled" {
		// When system uses SCA mode, then it is not necessary to try to get
		// system purpose status. We can prettyPrint disabled, because system purpose
		// has always this status in SCA mode
		fmt.Printf("System Purpose Status: disabled\n")
	} else {
		// When entitlement mode is used, then we need to get system purpose
		// status.
		rhsmSyspurposeCompliant := RHSMSyspurposeCompliant{}
		res, err := rhsmClient.ConsumerCertAuthConnection.request(
			http.MethodGet,
			"consumers/"+*uuid+"/purpose_compliance",
			"",
			"",
			nil,
			nil)
		if err != nil {
			return err
		}

		resBody, err := getResponseBody(res)
		if err != nil {
			return err
		}

		err = json.Unmarshal([]byte(*resBody), &rhsmSyspurposeCompliant)
		if err != nil {
			return err
		}

		fmt.Printf("System Purpose Status: %s\n", rhsmSyspurposeCompliant.Status)
	}

	return nil
}
