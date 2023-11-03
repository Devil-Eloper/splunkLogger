package main

import (
	"fmt"
	"github.azc.ext.hp.com/Devil-Eloper/splunkLogger/lib"
)

func main() {
	envErrors := lib.InitializeEnvironment()
	if envErrors != nil {
		return
	}
	// This would be declared at the start of lambda and logger would be passed to the lambda handler function
	logger := lib.New(lib.Environment[lib.SplunkToken], lib.Environment[lib.SplunkUrl], lib.Environment[lib.BatchSize])

	logger.Info("messageId", "et", "Info test message1")
	logger.Info("messageId2", "et2", "Info test message2")
	logger.Warn("messageId", "et", "Warn test message1")
	logger.Warn("messageId2", "et2", "Warn test message2")
	logger.Error("messageId", "et", "Error test message1")
	logger.Error("messageId2", "et2", "Error test message2")
	logger.Debug("messageId", "et", "Debug test message1")
	logger.Debug("messageId2", "et2", "Debug test message2")
	err := logger.SendBatch(true)
	if err != nil {
		fmt.Printf("Failed to send logs to Splunk: %s", err)
	}

}
