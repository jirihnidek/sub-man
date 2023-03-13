package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
)

type Facts struct {
	SystemCertificateVersion string `json:"system.certificate_version"`
}

// RegisterData is structure representing JSON data used for register request
type RegisterData struct {
	Type              string `json:"type"`
	Name              string `json:"name"`
	Facts             *Facts `json:"facts"`
	InstalledProducts []struct {
		ProductId   string `json:"productId"`
		ProductName string `json:"productName"`
		Version     string `json:"version"`
		Arch        string `json:"arch"`
	} `json:"installedProducts"`
	ContentTags  []interface{} `json:"contentTags"`
	Role         string        `json:"role"`
	AddOns       []interface{} `json:"addOns"`
	Usage        string        `json:"usage"`
	ServiceLevel string        `json:"serviceLevel"`
}

// ConsumerData is structure used for parsing JSON data returned during registration
// when system was successfully registered and consumer was created
type ConsumerData struct {
	Created             string        `json:"created"`
	Updated             string        `json:"updated"`
	Id                  string        `json:"id"`
	Uuid                string        `json:"uuid"`
	Name                string        `json:"name"`
	Username            string        `json:"username"`
	EntitlementStatus   string        `json:"entitlementStatus"`
	ServiceLevel        string        `json:"serviceLevel"`
	Role                string        `json:"role"`
	Usage               string        `json:"usage"`
	AddOns              []interface{} `json:"addOns"`
	SystemPurposeStatus string        `json:"systemPurposeStatus"`
	ReleaseVer          struct {
		ReleaseVer interface{} `json:"releaseVer"`
	} `json:"releaseVer"`
	Owner struct {
		Id                string `json:"id"`
		Key               string `json:"key"`
		DisplayName       string `json:"displayName"`
		Href              string `json:"href"`
		ContentAccessMode string `json:"contentAccessMode"`
	} `json:"owner"`
	Environment      interface{} `json:"environment"`
	EntitlementCount int         `json:"entitlementCount"`
	Facts            struct {
	} `json:"facts"`
	LastCheckin       interface{} `json:"lastCheckin"`
	InstalledProducts interface{} `json:"installedProducts"`
	CanActivate       bool        `json:"canActivate"`
	Capabilities      interface{} `json:"capabilities"`
	HypervisorId      interface{} `json:"hypervisorId"`
	ContentTags       interface{} `json:"contentTags"`
	Autoheal          bool        `json:"autoheal"`
	Annotations       interface{} `json:"annotations"`
	ContentAccessMode interface{} `json:"contentAccessMode"`
	Type              struct {
		Created  interface{} `json:"created"`
		Updated  interface{} `json:"updated"`
		Id       string      `json:"id"`
		Label    string      `json:"label"`
		Manifest bool        `json:"manifest"`
	} `json:"type"`
	IdCert struct {
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
	} `json:"idCert"`
	GuestIds       []interface{} `json:"guestIds"`
	Href           string        `json:"href"`
	ActivationKeys []interface{} `json:"activationKeys"`
	ServiceType    interface{}   `json:"serviceType"`
	Environments   interface{}   `json:"environments"`
}

func registerUsernamePasswordOrg(username *string, password *string, org *string) error {
	var headers = make(map[string]string)

	headers["username"] = *username
	headers["password"] = *password

	// TODO: when organization is not specified using CLI option --organization,
	//       then get list of available organization using: GET /users/<username>/owners

	var query string
	if *org != "" {
		query = "owner=" + *org
	} else {
		query = ""
	}

	// It is necessary to set system certificate version to value 3.0 or higher
	facts := Facts{
		SystemCertificateVersion: "3.2",
		// TODO: try to get some real facts.
	}

	hostname, err := os.Hostname()
	if err != nil {
		return fmt.Errorf("unable to get hostname: %s", err)
	}

	sysPurpose, err := getSystemPurpose(DefaultSystemPurposeFilePath)
	if err != nil {
		return err
	}

	// Create body for the register request
	headers["Content-type"] = "application/json"
	registerData := RegisterData{
		Type:         "system",
		Name:         hostname,
		Facts:        &facts,
		Role:         sysPurpose.Role,
		Usage:        sysPurpose.Usage,
		ServiceLevel: sysPurpose.ServiceLevelAgreement,
		// TODO: get list of installed product certificates
		// TODO: get list of all tags in installed product certificates
	}
	body, err := json.Marshal(registerData)
	if err != nil {
		return err
	}

	res, err := rhsmClient.NoAuthConnection.request(
		http.MethodPost,
		"consumers",
		query,
		"",
		&headers,
		&body)

	if err != nil {
		return err
	}

	resBody, err := getResponseBody(res)

	if err != nil {
		return err
	}

	consumerData := ConsumerData{}
	err = json.Unmarshal([]byte(*resBody), &consumerData)
	if err != nil {
		return err
	}

	err = writeConsumerCert(&consumerData.IdCert.Cert)
	if err != nil {
		return err
	}

	err = writeConsumerKey(&consumerData.IdCert.Key)
	if err != nil {
		return err
	}

	certFilePath := filepath.Join(rhsmClient.RHSMConf.RHSM.ConsumerCertDir, "cert.pem")
	keyFilePath := filepath.Join(rhsmClient.RHSMConf.RHSM.ConsumerCertDir, "key.pem")
	rhsmClient.ConsumerCertAuthConnection, err = createCertAuthConnection(
		&rhsmClient.RHSMConf.Server.Hostname,
		&rhsmClient.RHSMConf.Server.Port,
		&rhsmClient.RHSMConf.Server.Prefix,
		&certFilePath,
		&keyFilePath,
	)

	if err != nil {
		return err
	}

	// When we are in SCA mode, then we can get entitlement cert and generate content
	if consumerData.Owner.ContentAccessMode == "org_environment" {
		err = getSCAEntitlementCertificate()
		if err != nil {
			return err
		}
	}

	// TODO: when not-SCA mode is used, then try to do auto-attach, when it was requested and generate content

	return nil
}

// writeConsumerCert tries to write consumer certificate. It is
// typically /etc/pki/consumer/cert.pem
func writeConsumerCert(consumerCert *string) error {
	consumerCertFile := rhsmClient.consumerCertPath()
	return writePemFile(consumerCertFile, consumerCert)
}

// writeConsumerKey tries to write consumer key. It is typically
// /etc/pki/consumer/key.pem
func writeConsumerKey(consumerKey *string) error {
	consumerKeyFile := rhsmClient.consumerKeyPath()
	return writePemFile(consumerKeyFile, consumerKey)
}
