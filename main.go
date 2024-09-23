package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"

	"semconvdiff/internal"
)

const (
	FILE_BASE_URL = "https://raw.githubusercontent.com/open-telemetry/opentelemetry-collector/semconv/v0.109.0/semconv"
)

var (
	singleLineConstPattern = regexp.MustCompile(`const\s+(\w+)\s+= \s+"([^"]+)"`)
	blockConstPattern      = regexp.MustCompile(`const\s*\((.*?)\)`)
	extraConstPattern      = regexp.MustCompile(`(\w+)\s*= \s*"([^"]+)"`)
)

func fetchFile(url string) string {
	resp, err := http.Get(url)
	if err != nil {
		internal.ErrorPrintf("Failed to fetch %s: %v", url, err)
		return ""
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		internal.ErrorPrintf("Failed to fetch %s with status code %d", url, resp.StatusCode)
		return ""
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		internal.ErrorPrintf("Error reading response from %s: %v", url, err)
		return ""
	}

	return string(body)
}

func extractSingleLineConstants(content string) map[string]string {
	constants := make(map[string]string)
	matches := singleLineConstPattern.FindAllStringSubmatch(content, -1)

	for _, match := range matches {
		constName := match[1]
		constValue := match[2]
		constants[constName] = constValue
	}

	return constants
}

func extractBlockConstants(content string) map[string]string {
	constants := make(map[string]string)
	blockMatches := blockConstPattern.FindAllStringSubmatch(content, -1)

	for _, block := range blockMatches {
		blockContent := block[1]
		blockConstants := extraConstPattern.FindAllStringSubmatch(blockContent, -1)

		for _, match := range blockConstants {
			constName := match[1]
			constValue := match[2]
			constants[constName] = constValue
		}
	}

	return constants
}

func fetchConstants(url string) map[string]string {
	content := fetchFile(url)
	constants := extractSingleLineConstants(content)
	blockConstants := extractBlockConstants(content)

	for constName, constValue := range blockConstants {
		constants[constName] = constValue
	}

	extraConstants := extraConstPattern.FindAllStringSubmatch(content, -1)
	for _, match := range extraConstants {
		constName := match[1]
		constValue := match[2]
		if _, exists := constants[constName]; !exists {
			constants[constName] = constValue
		}
	}

	return constants
}

func fetchConstantsFromVersion(version string) map[string]string {
	allConstants := make(map[string]string)
	files, err := internal.FetchFileNames(version)
	if err != nil {
		internal.DebugPrintf("Failed to fetch files %s: %v", version, err)
		return map[string]string{}
	}
	internal.InfoPrint(version, "found files:", files)
	for _, file := range files {
		url := fmt.Sprintf("%s/%s/%s", FILE_BASE_URL, version, file)
		internal.InfoPrintf("Fetching constants from %s...", url)
		constants := fetchConstants(url)
		for k, v := range constants {
			allConstants[k] = v
		}

		if len(constants) == 0 {
			internal.DebugPrint(version, "no constant found on", file)
		}
	}

	return allConstants
}

func compareConstants(constantsV1, constantsV2 map[string]string, v1, v2 string) {
	fmt.Printf("%-30s | %-50s | %-50s\n", "Constant Name", v1, v2)
	fmt.Println(strings.Repeat("=", 140))

	for constName, valueV1 := range constantsV1 {
		if valueV2, exists := constantsV2[constName]; exists && valueV1 != valueV2 {
			fmt.Printf("%-30s | %-50s | %-50s\n", constName, valueV1, valueV2)
		}
	}
}

func main() {
	version1 := flag.String("v1", "v1", "Version 1 to compare")
	version2 := flag.String("v2", "v2", "Version 2 to compare")
	flag.Parse()

	constantsV1 := fetchConstantsFromVersion(*version1)
	constantsV2 := fetchConstantsFromVersion(*version2)

	internal.InfoPrint(*version1, "has", len(constantsV1), "constants")
	internal.InfoPrint(*version2, "has", len(constantsV2), "constants")

	compareConstants(constantsV1, constantsV2, *version1, *version2)
}
