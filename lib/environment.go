package lib

import (
	"fmt"
	"os"
)

// variableInformation defines a holder for environment variable details
type variableInformation struct {
	required     bool
	defaultValue string
}

// Environment environment stores a map of environment values
var Environment = map[string]string{}

// environmentInformation stores a map of the required flag and default value of environment variables
var environmentInformation = map[string]variableInformation{
	SplunkUrl:   {required: true},
	SplunkToken: {required: true},
	BatchSize:   {required: true},
}

// InitializeEnvironment initializeEnvironment initializes the environment map and ensures all required values are set
func InitializeEnvironment() error {
	for key, reqType := range environmentInformation {

		// Look up the key, and error out if it's not present
		value, present := os.LookupEnv(key)
		if !present && reqType.required {
			return fmt.Errorf("%s is a required environment variable", key)
		}

		// Store the value if any or the default one otherwise
		if value == "" {
			Environment[key] = reqType.defaultValue
		} else {
			Environment[key] = value
		}
	}

	return nil
}
