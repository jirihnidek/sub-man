package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
)

// SystemFacts is collection of system facts necessary during registration
type SystemFacts struct {
	SystemCertificateVersion string `json:"system.certificate_version"`
}

// RegisterData is structure representing JSON data used for register request
type RegisterData struct {
	Type              string             `json:"type"`
	Name              string             `json:"name"`
	Facts             *SystemFacts       `json:"facts"`
	InstalledProducts []InstalledProduct `json:"installedProducts"`
	ContentTags       []string           `json:"contentTags"`
	Role              string             `json:"role"`
	AddOns            []interface{}      `json:"addOns"`
	Usage             string             `json:"usage"`
	ServiceLevel      string             `json:"serviceLevel"`
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

// registerUsernamePasswordOrg tries to register system using organization id, username and password
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
	facts := SystemFacts{
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

	installedProducts := getInstalledProducts()

	contentTags := createListOfContentTags(installedProducts)

	// Create body for the register request
	headers["Content-type"] = "application/json"
	registerData := RegisterData{
		Type:              "system",
		Name:              hostname,
		Facts:             &facts,
		Role:              sysPurpose.Role,
		Usage:             sysPurpose.Usage,
		ServiceLevel:      sysPurpose.ServiceLevelAgreement,
		InstalledProducts: installedProducts,
		ContentTags:       contentTags,
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

	fmt.Printf("System registered\n")

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

	// When we are in SCA mode, then we can get entitlement cert(s) and generate content
	if consumerData.Owner.ContentAccessMode == "org_environment" {
		err = enableContent()
		if err != nil {
			return err
		}
	} else {
		return fmt.Errorf("organization %s does not use Simple Content Access Mode",
			consumerData.Owner.DisplayName)
	}

	return nil
}

// enableContent tries to get SCA entitlement certificate and generate redhat.repo from this
// certificate
func enableContent() error {
	// Try to get entitlement certificate(s) from server
	entCertKeys, err := getSCAEntitlementCertificate()
	if err != nil {
		return err
	}

	if len(entCertKeys) > 1 {
		fmt.Printf("Entitlement certificates installed\n")
	} else {
		fmt.Printf("Entitlement certificate installed\n")
	}

	// Get content from entitlement certificates
	// Note: candlepin returns only one entitlement certificate in SCA mode, but
	// in theory more entitlement certificate can be returned
	var engineeringProducts = make(map[int64][]EngineeringProduct)
	for _, entCertKey := range entCertKeys {
		serial := entCertKey.Serial.Serial
		certContent := &entCertKey.Cert
		products, err := getContentFromEntCert(certContent)
		if err != nil {
			// TODO: write some warning to log file
			continue
		}
		engineeringProducts[serial] = products
	}

	// Write content to redhat.repo file
	if len(engineeringProducts) > 0 {
		err = writeRepoFile(DefaultRepoFilePath, engineeringProducts)
		if err != nil {
			return fmt.Errorf("unable to write repo file: %s: %s", DefaultRepoFilePath, err)
		}
	}

	fmt.Printf("%s was generated\n", DefaultRepoFilePath)

	return nil
}

// getInstalledProducts tries to get all installed products. Typically from directories:
// /etc/pki/product and /etc/pki/product-default
func getInstalledProducts() []InstalledProduct {
	installedProducts, err := readAllProductCertificates(rhsmClient.RHSMConf.RHSM.ProductCertDir)
	if err != nil {
		// TODO: print some warning message to log file
		// log.Printf("failed reading prod dir: %s\n", err)
	}

	installedDefaultProducts, err := readAllProductCertificates(DirectoryDefaultProductCertificate)
	if err != nil {
		// TODO: print some warning message to log file
		// log.Printf("failed reading default prod dir: %s\n", err)
	}

	installedProducts = append(installedProducts, installedDefaultProducts...)
	return installedProducts
}

// createListOfContentTags creates list of unique tags from the list of installed products
func createListOfContentTags(installedProducts []InstalledProduct) []string {
	var contentTags []string
	// We use map, because there is nothing like set
	var contentTagsMap = make(map[string]bool)
	for _, prod := range installedProducts {
		for _, tag := range prod.providedTags {
			_, exists := contentTagsMap[tag]
			if !exists {
				contentTagsMap[tag] = true
			}
		}
	}
	// Create list from the map
	for tagName, _ := range contentTagsMap {
		contentTags = append(contentTags, tagName)
	}
	return contentTags
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
