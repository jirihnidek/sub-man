package main

import (
	"fmt"
	"github.com/creasty/defaults"
	"gopkg.in/ini.v1"
	"reflect"
	"strconv"
	"strings"
)

const defaultRHSMConfFilePath = "/etc/rhsm/rhsm.conf"

// RHSMConf is structure intended for storing configuration
// that is typically read from /etc/rhsm/rhsm.conf. We try to
type RHSMConf struct {
	// not public attribute
	filePath string

	// Server represents section [server]
	Server struct {
		// Basic settings for connection to candlepin server
		Hostname string `ini:"hostname" default:"subscription.rhsm.redhat.com"`
		Prefix   string `ini:"prefix" default:"/subscription"`
		Port     string `ini:"port" default:"443"`
		Insecure bool   `ini:"insecure" default:"false"`
		Timeout  int64  `ini:"server_timeout" default:"180"`

		// Proxy settings
		ProxyHostname string `ini:"proxy_hostname" default:""`
		ProxyScheme   string `ini:"proxy_scheme" default:"http" allowedValues:"http,https"`
		ProxyPort     string `ini:"proxy_port" default:"3128"`
		ProxyUser     string `ini:"proxy_user" default:""`
		ProxyPassword string `ini:"proxy_password" default:""`

		// Comma separated list of hostnames, when connection should not go
		// through proxy server
		NoProxy string `ini:"no_proxy" default:""`
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
		ManageRepos          bool   `ini:"manage_repos" default:"true"`

		// Configuration options related to DNF plugins
		AutoEnableYumPlugins  bool `ini:"auto_enable_yum_plugins" default:"true"`
		PackageProfileOnTrans bool `ini:"package_profile_on_trans" default:"false"`
	} `ini:"rhsm"`

	// RHSMCertDaemon represents section [rhsmcertd]
	RHSMCertDaemon struct {
		AutoRegistration         bool  `ini:"auto_registration" default:"false"`
		AutoRegistrationInterval int64 `ini:"auto_registration_interval" default:"60"`
		Splay                    bool  `ini:"splay" default:"true"`
	} `ini:"rhsmcertd"`

	// Logging represents section [logging]
	Logging struct {
		DefaultLogLevel string `ini:"default_log_level" default:"INFO" allowedValues:"ERROR,WARN,INFO,DEBUG"`
	} `ini:"logging"`
}

// setDefaultValues tries to set default values specified in tags
func (rhsmConf *RHSMConf) setDefaultValues() error {
	err := defaults.Set(rhsmConf)
	if err != nil {
		return err
	}
	return nil
}

// load tries to load configuration file (usually /etc/rhsm/rhsm.conf)
func (rhsmConf *RHSMConf) load() error {
	cfg, err := ini.Load(rhsmConf.filePath)
	if err != nil {
		return err
	}

	// First set default values
	err = rhsmConf.setDefaultValues()
	if err != nil {
		return err
	}

	// Then try to load values from given configuration file.
	// Note that parsing errors are ignored and default values are used in case of some errors.
	err = cfg.MapTo(rhsmConf)
	if err != nil {
		return err
	}

	return nil
}

// isDefaultValue tries to say if given value is default value or not
func isDefaultValue(value *reflect.Value, defaultValue *string) (bool, error) {
	switch value.Kind() {
	case reflect.String:
		return value.String() == *defaultValue, nil
	case reflect.Int, reflect.Int64:
		intVal := value.Int()
		defaultIntVal, err := strconv.ParseInt(*defaultValue, 10, 64)
		if err != nil {
			return false, err
		}
		return intVal == defaultIntVal, nil
	case reflect.Bool:
		boolVal := value.Bool()
		defaultBoolVal, err := strconv.ParseBool(*defaultValue)
		if err != nil {
			return false, err
		}
		return boolVal == defaultBoolVal, nil
	default:
		return false, fmt.Errorf("unsupported type of value: %s", value.Kind())
	}
}

// isValueAllowed tries to say if given value is allowed or not. The allowedValues
// is string with comma separated values
func isValueAllowed(value *reflect.Value, allowedValues *string) (bool, error) {
	allowedValuesSlice := strings.Split(*allowedValues, ",")

	// Only strings are supported at this moment, because I cannot imagine another use case
	// with another e.g. allowed integer values. When it will be necessary, then it is easy to add
	// support for another type
	switch value.Kind() {
	case reflect.String:
		val := value.String()
		for _, allowedValue := range allowedValuesSlice {
			if val == allowedValue {
				return true, nil
			}
		}
		return false, nil
	default:
		return false, fmt.Errorf("unsupported type of allowed value: %s", value.Kind())
	}
}

const colorRed = "\033[0;91m"
const colorGreen = "\033[0;92m"
const noColor = "\033[0m"

// prettyPrintOption print option key and value with some colors. Green color means that
// the value differs from default value and red color means that the value is not allowed
// ot there was some problem with parsing value
func prettyPrintOption(optValue reflect.Value, valType reflect.StructField) {
	tag := valType.Tag
	value := optValue.Interface()

	// We care only about files field with "ini" tag
	tagIniValue, ok := tag.Lookup("ini")
	if !ok {
		return
	}

	// All configuration options have to have some default tag
	tagDefaultValue, ok := tag.Lookup("default")
	if !ok {
		return
	}

	// When value is default, then print it with white color
	isDefault, err := isDefaultValue(&optValue, &tagDefaultValue)
	if err != nil {
		fmt.Printf("%s    %s = %v (error: %s) %s\n", colorRed, tagIniValue, value, err, noColor)
		return
	}
	if isDefault {
		fmt.Printf("    %s = %v\n", tagIniValue, value)
		return
	}

	// When value is not default, then check if field contains tag "allowedValues". When value is not
	// allowed, then print the field with red value
	tagAllowedValues, ok := tag.Lookup("allowedValues")
	if ok {
		isAllowed, err := isValueAllowed(&optValue, &tagAllowedValues)
		if err != nil {
			fmt.Printf("%s    %s = %v (error: %s) %s\n", colorRed, tagIniValue, value, err, noColor)
			return
		}
		if !isAllowed {
			fmt.Printf("%s    %s = %v %s\n", colorRed, tagIniValue, value, noColor)
			return
		}
	}

	// When value is not default, and it is allowed, then print it with green color
	fmt.Printf("%s    %s = %v %s\n", colorGreen, tagIniValue, value, noColor)
	return
}

// prettyPrintOptions tries to print keys and values of one configuration
// section like [server] section
func (rhsmConf *RHSMConf) prettyPrintOptions(section *reflect.Value) error {
	values := section.Interface()
	valuesOfSections := reflect.ValueOf(values)
	typesOfOptions := valuesOfSections.Type()

	for i := 0; i < valuesOfSections.NumField(); i++ {
		if typesOfOptions.Field(i).IsExported() {
			prettyPrintOption(valuesOfSections.Field(i), typesOfOptions.Field(i))
		}
	}
	return nil
}

// prettyPrint tries to pretty print structure of configuration
func (rhsmConf *RHSMConf) prettyPrint() error {
	valuesOfRHSMConf := reflect.ValueOf(*rhsmConf)
	typesOfRHSMConf := valuesOfRHSMConf.Type()

	for i := 0; i < valuesOfRHSMConf.NumField(); i++ {
		kind := valuesOfRHSMConf.Field(i).Kind()
		tag := typesOfRHSMConf.Field(i).Tag
		if kind == reflect.Struct {
			section := valuesOfRHSMConf.Field(i)
			tagIniValue, ok := tag.Lookup("ini")
			if ok {
				fmt.Printf("[%s]\n", tagIniValue)
				_ = rhsmConf.prettyPrintOptions(&section)
				fmt.Printf("\n")
			}
		} else {
			tagValue, ok := tag.Lookup("ini")
			if ok {
				fmt.Printf("%v\n", tagValue)
			}
		}
	}
	return nil
}

// loadRHSMConf tries to load given configuration file to
// RHSMConf structure
func loadRHSMConf(confFilePath *string) (*RHSMConf, error) {
	rhsmConf := &RHSMConf{filePath: *confFilePath}

	err := rhsmConf.load()

	if err != nil {
		return nil, err
	}

	return rhsmConf, nil
}
