package main

import (
	"fmt"
	"os"
)

// writePemFile Tries to write content of PEM (cert or key) to file
func writePemFile(filePath *string, pemFileContent *string) error {
	file, err := os.Create(*filePath)

	if err != nil {
		return err
	}

	defer file.Close()

	// Print content of cert using Fprint(), because
	// the string contains formatting sequences like \n
	_, err = fmt.Fprint(file, *pemFileContent)

	if err != nil {
		return err
	}

	return nil
}
