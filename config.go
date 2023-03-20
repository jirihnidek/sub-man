package main

import (
	"encoding/json"
	"fmt"
	"github.com/creasty/defaults"
	"gopkg.in/ini.v1"
)

const defaultRHSMConfFilePath = "/etc/rhsm/rhsm.conf"

// RHSMConf is structure intended for storing configuration
// that is typically read from /etc/rhsm/rhsm.conf. We try to
type RHSMConf struct {
	FilePath string

	// Server represents section [server]
	Server struct {
		// Basic settings for connection to candlepin server
		Hostname string `ini:"hostname" default:"subscription.rhsm.redhat.com"`
		Prefix   string `ini:"prefix" default:"/subscription"`
		Port     string `ini:"port" default:"443"`
		Insecure bool   `ini:"insecure" default:"false"`
		Timeout  int32  `ini:"server_timeout" default:"180"`

		// Proxy settings
		ProxyHostname string `ini:"proxy_hostname" default:""`
		ProxyScheme   string `ini:"proxy_scheme" default:"http"`
		ProxyPort     string `ini:"proxy_port" default:"3128"`
		ProxyUser     string `ini:"proxy_user" default:""`
		ProxyPassword string `ini:"proxy_password" default:""`

		// List of hostnames, when connection should not go through proxy server
		NoProxy []string `ini:"no_proxy" default:"[]"`
	} `ini:"server"`

	// RHSM represents section [rhsm]
	RHSM struct {
		// Directories used for certificates
		CACertDir          string `ini:"ca_cert_dir" default:"/etc/rhsm/ca/"`
		ConsumerCertDir    string `ini:"consumercertdir" default:"/etc/pki/consumer"`
		EntitlementCertDir string `ini:"entitlementcertdir" default:"/etc/pki/entitlement"`
		ProductCertDir     string `ini:"productcertdir" default:"/etc/pki/product"`

		// Configuration options related to RPMs and repositories
		BaseURL              string `ini:"baseurl" default:"https://cdn.redhat.com"`
		ReportPackageProfile bool   `ini:"report_package_profile" default:"true"`
		RepoCACertificate    string `ini:"repo_ca_cert" default:"/etc/rhsm/ca/redhat-uep.pem"`
	} `ini:"rhsm"`

	// RHSMCertDaemon represents section [rhsmcertd]
	RHSMCertDaemon struct {
		AutoRegistration         bool  `ini:"auto_registration" default:"false"`
		AutoRegistrationInterval int32 `ini:"auto_registration_interval" default:"60"`
		Splay                    bool  `ini:"splay" default:"tru"`
	} `ini:"rhsmcertd"`

	// Logging represents section [logging]
	Logging struct {
		DefaultLogLevel string `ini:"default_log_level"`
	} `ini:"logging"`
}

func (rhsmConf *RHSMConf) setDefaultValues() error {
	err := defaults.Set(rhsmConf)
	if err != nil {
		return err
	}
	return nil
}

// load tries to load configuration file (usually /etc/rhsm/rhsm.conf)
func (rhsmConf *RHSMConf) load() error {
	cfg, err := ini.Load(rhsmConf.FilePath)
	if err != nil {
		return err
	}

	// First set default values
	err = rhsmConf.setDefaultValues()
	if err != nil {
		return err
	}

	// Then try to load values from given configuration file
	err = cfg.MapTo(rhsmConf)
	if err != nil {
		return err
	}

	return nil
}

// prettyPrint tries to print structure with configuration
func (rhsmConf *RHSMConf) prettyPrint() error {
	s, err := json.MarshalIndent(rhsmConf, "", "    ")
	if err != nil {
		return err
	}
	fmt.Printf("%v\n", string(s))
	return nil
}

// loadRHSMConf tries to load given configuration file to
// RHSMConf structure
func loadRHSMConf(confFilePath *string) (*RHSMConf, error) {
	rhsmConf := &RHSMConf{FilePath: *confFilePath}

	err := rhsmConf.load()

	if err != nil {
		return nil, err
	}

	return rhsmConf, nil
}
