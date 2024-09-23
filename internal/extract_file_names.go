package internal

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"
)

const (
	baseUrl                  = "https://github.com/open-telemetry/opentelemetry-collector/blob/semconv/v0.109.0/semconv"
	somethingWrongBodyErrMsg = "Looks like something went wrong!"
	maxRetries               = 6
)

func extractAfterSlash(url string) string {
	lastSlashIndex := strings.LastIndex(url, "/")
	if lastSlashIndex != -1 {
		return url[lastSlashIndex+1:]
	}
	log.Fatalf("unable to extract after spash for %s", url)
	return ""
}

func tryFetchFileNames(url string) ([]string, error) {
	// Send a GET request to the URL
	response, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	// Check if the response status is OK
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch URL: %s, status: %s", url, response.Status)
	}

	// Read the response body
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	// Use regex to find file names in the HTML
	re := regexp.MustCompile(`href="([^"]+)"`)
	matches := re.FindAllStringSubmatch(string(body), -1)

	set := make(map[string]struct{})
	for _, match := range matches {
		if strings.HasSuffix(match[0], ".go") {
			set[extractAfterSlash(match[0])] = struct{}{}
		}
		if strings.HasSuffix(match[1], ".go") {
			set[extractAfterSlash(match[1])] = struct{}{}
		}
	}
	var fileNames []string
	for key := range set {
		fileNames = append(fileNames, key)
	}
	return fileNames, nil
}

func FetchFileNames(version string) ([]string, error) {
	url := fmt.Sprintf("%s/%s", baseUrl, version)
	var resp []string
	var err error
	for i := range maxRetries {
		resp, err = tryFetchFileNames(url)
		if len(resp) > 0 {
			return resp, err
		}
		sleepDuration := time.Duration((i+3)*2) * time.Second
		DebugPrintf("Something went wrong while requesting files. version: %s, err:%v, retries %d, retrying in %s...", version, err, i, sleepDuration)
		time.Sleep(sleepDuration)
	}
	return resp, err
}
