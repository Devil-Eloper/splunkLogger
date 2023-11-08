package lib

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"
)

type Logger struct {
	httpClient        *http.Client
	splunkHECToken    string
	splunkHECEndpoint string
	batchSize         int
	logs              []StructuredLog
	mutex             sync.Mutex
}

type StructuredLog struct {
	Time      string `json:"time"`
	MessageId string `json:"messageId"`
	Et        string `json:"et"`
	Message   string `json:"message"`
	Level     string `json:"level"`
}

type SplunkEvents struct {
	Event      string `json:"event"`
	Sourcetype string `json:"sourcetype"`
}

func New(splunkToken, splunkEndpoint, batchSize string) *Logger {
	batchLimit, _ := strconv.Atoi(batchSize)
	log.Print("Splunk 1")
	return &Logger{
		httpClient:        &http.Client{Timeout: 10 * time.Second}, // This can be passed from the caller so that throughout the lifecycle we maintain only 1 httpClient
		splunkHECToken:    splunkToken,
		splunkHECEndpoint: splunkEndpoint,
		logs:              []StructuredLog{},
		batchSize:         batchLimit,
	}
}

func (logger *Logger) Info(messageId, et, message string) {

	logger.addLogs(messageId, et, message, Info)

}

func (logger *Logger) Warn(messageId, et, message string) {

	logger.addLogs(messageId, et, message, Warn)

}

func (logger *Logger) Error(messageId, et, message string) {

	logger.addLogs(messageId, et, message, Error)

}

func (logger *Logger) Debug(messageId, et, message string) {

	logger.addLogs(messageId, et, message, Debug)

}

func (logger *Logger) SendBatch(batchSend bool) error {
	if !batchSend {
		return nil
	}
	log.Print("Splunk 2")
	logger.mutex.Lock()
	logs := logger.logs
	logger.logs = nil
	logger.mutex.Unlock()
	// Buffer to store the complete payload
	var buffer bytes.Buffer
	log.Print("Splunk 3")
	// Iterate over the log entries and marshal them individually
	for _, event := range logs {
		// Wrap each log entry in an object with the "event" field

		eventWrapper := map[string]interface{}{
			Event: event,
		}
		eventJSON, err := json.Marshal(eventWrapper)
		if err != nil {
			return err
		}

		// Write the JSON to the buffer, followed by a newline if needed
		buffer.Write(eventJSON)
		buffer.WriteString(Newline) // Splunk expects newline-separated events

	}
	log.Print("Splunk 4")
	request, err := http.NewRequest(Post, logger.splunkHECEndpoint, &buffer)
	if err != nil {
		log.Print("Splunk 4.1", err.Error())
		return err
	}

	request.Header.Set(Authorization, Splunk+logger.splunkHECToken)
	request.Header.Set(ContentType, ApplicationJson)

	//response, err := logger.httpClient.Do(request)
	client := &http.Client{Timeout: 10 * time.Second}

	//If we are working with a trial account the certificate will be self-signed, so we want to ignore certificate verification
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client.Transport = tr

	response, err := client.Do(request)
	if err != nil {
		log.Print("Splunk 5", err.Error())
		return err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Print("Splunk 6", err.Error())
		}
	}(response.Body)

	log.Print("Splunk 7", http.StatusOK)
	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("splunk error: %d", response.StatusCode)
	}

	return nil
}
func (logger *Logger) addLogs(messageId, et, message, level string) {
	structuredLog := StructuredLog{
		Time:      time.Now().Format(time.RFC3339),
		MessageId: messageId,
		Et:        et,
		Message:   message,
		Level:     level,
	}
	logger.mutex.Lock()
	logger.logs = append(logger.logs, structuredLog)
	batchSend := len(logger.logs) >= logger.batchSize
	logger.mutex.Unlock()
	log.Print("Splunk 8")
	if batchSend {
		err := logger.SendBatch(batchSend)
		if err != nil {
			log.Print("Splunk 8.1")
		} // Executes separately. Not handling the error yet.
	}
}
