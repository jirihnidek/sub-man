package main

import (
	"encoding/json"
	"fmt"
	"os"
)

const DefaultSystemPurposeFilePath = "/etc/rhsm/syspurpose/syspurpose.json"

// SysPurposeJSON is structure holding system purpose attributes
type SysPurposeJSON struct {
	Role                  string `json:"role"`
	ServiceLevelAgreement string `json:"service_level_agreement"`
	Usage                 string `json:"usage"`
}

// loadFromFile tries to load system purpose from the file
func (sysPurpose *SysPurposeJSON) loadFromFile(filePath string) error {
	sysPurposeContent, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("unable to read system purpose file: %s, %s", filePath, err)
	}

	err = json.Unmarshal(sysPurposeContent, sysPurpose)
	if err != nil {
		return fmt.Errorf("unable to unmarshal system purpose file: %s: %s", filePath, err)
	}

	return nil
}

// getSystemPurpose tries to load system purpose from given file
func getSystemPurpose(filePath string) (*SysPurposeJSON, error) {
	var sysPurpose = SysPurposeJSON{"", "", ""}

	_, err := os.Stat(filePath)
	if err != nil {
		return &sysPurpose, nil
	}

	err = sysPurpose.loadFromFile(filePath)
	if err != nil {
		return nil, err
	}

	return &sysPurpose, nil
}
