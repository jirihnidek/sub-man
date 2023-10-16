package main

import (
	"fmt"
	"github.com/jirihnidek/rhsm2"
	"reflect"
)

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
	isDefault, err := rhsm2.IsDefaultValue(&optValue, &tagDefaultValue)
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
		isAllowed, err := rhsm2.IsValueAllowed(&optValue, &tagAllowedValues)
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
}

// prettyPrintOptions tries to print keys and values of one configuration
// section like [server] section
func prettyPrintOptions(section *reflect.Value) error {
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
func prettyPrint(rhsmConf *rhsm2.RHSMConf) error {
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
				_ = prettyPrintOptions(&section)
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
