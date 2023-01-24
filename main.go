package main

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"flag"
	"fmt"
	"github.com/urfave/cli/v2"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

const serverHostname = "centos8"
const serverPort = 8443
const prefix = "candlepin"
const insecure = true

//const serverHostname = "subscription.rhsm.redhat.com"
//const serverPort = 443
//const prefix = "subscription"
//const insecure = false

var (
	certFile = flag.String(
		"cert",
		"/etc/pki/consumer/cert.pem",
		"A PEM encoded certificate file.",
	)
	keyFile = flag.String(
		"key",
		"/etc/pki/consumer/key.pem",
		"A PEM encoded private key file.",
	)
	caFile = flag.String(
		"CA",
		"/etc/rhsm/ca/redhat-uep.pem",
		"A PEM encoded CA's certificate file.",
	)
)

// getConsumerUUID tries to get consumer UUID from installed consumer certificate
// TODO: the content (the parsed certificate) should be kept in memory cache to minimize reading
//       consumer certificate file. There should be also i-notify monitor used for obsoleting this
//       in-memory cache, when some other process change content of consumer certificate file.
func getConsumerUUID(consumerCertFileName string) (*string, error) {
	consumerCert, err := os.ReadFile(consumerCertFileName)

	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode(consumerCert)
	if block == nil {
		return nil, fmt.Errorf("failed to parse PEM block containing the public key")
	}

	certificate, err := x509.ParseCertificate(block.Bytes)

	if err != nil {
		return nil, err
	}

	// Similar method could be use for getting other useful information (e.g. org ID).
	// fmt.Printf("Org ID: %v\n", certificate.Subject.Organization[0])

	return &certificate.Subject.CommonName, nil
}

// request tries to call HTTP request to candlepin server
func request(method string, handler string) (*string, error) {
	// Load CA certificate
	caCert, err := ioutil.ReadFile(*caFile)
	if err != nil {
		return nil, fmt.Errorf("error: Unable to read CA certificate: %s\n", err)
	}
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	// Try to load client certificate and key
	cert, err := tls.LoadX509KeyPair(*certFile, *keyFile)
	if err != nil {
		return nil, fmt.Errorf("error: unable to load client certificate and key: %s\n", err)
	}

	// Setup HTTPS client
	tlsConfig := &tls.Config{
		Certificates:       []tls.Certificate{cert},
		RootCAs:            caCertPool,
		InsecureSkipVerify: insecure,
	}

	transport := &http.Transport{TLSClientConfig: tlsConfig}
	client := &http.Client{Transport: transport}

	requestUrl := fmt.Sprintf("https://%s:%d/%s/%s", serverHostname, serverPort, prefix, handler)

	//fmt.Println("HTTP client: creating GET request:", requestUrl)

	req, err := http.NewRequest(method, requestUrl, nil)
	if err != nil {
		return nil, fmt.Errorf("error: creating http request GET: %s\n", err)
	}

	//fmt.Println("HTTP client: sending GET request:", req)

	res, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error: making http request GET: %s\n", err)
	}

	//fmt.Printf("HTTP client: got response!\n")
	//fmt.Printf("HTTP client: status code: %d\n", res.StatusCode)

	resBody, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("error: reading response body: %s\n", err)
	}
	//fmt.Printf("HTTP client: response body: %s\n", resBody)

	response := string(resBody[:])

	return &response, nil
}

// RHSMCompliant is structure used for storing GET response from REST API
// endpoint consumers/<consumer-UUID>/compliance
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
// endpoint consumers/<consumer-UUID>/purpose_compliance. This REST API endpoint
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

// statusAction tries to print status
func statusAction(ctx *cli.Context) error {
	uuid, err := getConsumerUUID(*certFile)

	if err != nil {
		return err
	}

	response, err := request(http.MethodGet, "consumers/"+*uuid+"/compliance")

	if err != nil {
		return err
	}

	rhsmCompliant := RHSMCompliant{}

	err = json.Unmarshal([]byte(*response), &rhsmCompliant)

	if err != nil {
		return err
	}

	fmt.Printf("Status: %s\n", rhsmCompliant.Status)

	if rhsmCompliant.Status == "disabled" {
		// When system uses SCA mode, then it is not necessary to try to get
		// system purpose status. We can print disabled, because system purpose
		// has always this status in SCA mode
		fmt.Printf("System Purpose Status: disabled\n")
	} else {
		// When entitlement mode is used, then we need to get system purpose
		// status.
		rhsmSyspurposeCompliant := RHSMSyspurposeCompliant{}
		response, err = request(http.MethodGet, "consumers/"+*uuid+"/purpose_compliance")

		if err != nil {
			return err
		}

		err = json.Unmarshal([]byte(*response), &rhsmSyspurposeCompliant)

		if err != nil {
			return err
		}

		fmt.Printf("System Purpose Status: %s\n", rhsmSyspurposeCompliant.Status)
	}

	return nil
}

// identityAction tries to print system identity
func identityAction(ctx *cli.Context) error {
	_, err := os.Stat(*certFile)

	if err != nil {
		return err
	}

	uuid, err := getConsumerUUID(*certFile)

	if err != nil {
		return err
	}

	fmt.Printf("system identity: %v\n", *uuid)

	return nil
}

func registerAction(ctx *cli.Context) error {
	return nil
}

func unregisterAction(ctx *cli.Context) error {
	return nil
}

func main() {
	app := &cli.App{
		Name:    "sub-man",
		Version: "0.0.1",
		Usage:   "Minimalistic CLI client for RHSM",
	}
	app.Commands = []*cli.Command{
		{
			Name: "register",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:    "username",
					Usage:   "register with `USERNAME`",
					Aliases: []string{"u"},
				},
				&cli.StringFlag{
					Name:    "password",
					Usage:   "register with `PASSWORD`",
					Aliases: []string{"p"},
				},
				&cli.StringFlag{
					Name:    "organization",
					Usage:   "register with `ID`",
					Aliases: []string{"o"},
				},
			},
			Usage:       "Register the system to " + serverHostname,
			UsageText:   fmt.Sprintf("%v register [command options]", app.Name),
			Description: fmt.Sprintf("The register command registers the system to Red Hat Subscription Management"),
			Action:      registerAction,
		},
		{
			Name:        "unregister",
			Usage:       "Unregister system",
			UsageText:   fmt.Sprintf("%v unregister", app.Name),
			Description: fmt.Sprintf("Unregister the system"),
			Action:      unregisterAction,
		},
		{
			Name:        "status",
			Usage:       "Print status",
			UsageText:   fmt.Sprintf("%v status", app.Name),
			Description: fmt.Sprintf("Print status of system"),
			Action:      statusAction,
		},
		{
			Name:        "identity",
			Usage:       "Print identity",
			UsageText:   fmt.Sprintf("%v identity", app.Name),
			Description: fmt.Sprintf("Print identity of system"),
			Action:      identityAction,
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
