package main

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
)

func main() {
	content, err := os.ReadFile("fanqie_reader.html")
	if err != nil {
		panic(err)
	}
	htmlContent := string(content)

	// 原来的正则
	stateRe := regexp.MustCompile(`window\.__INITIAL_STATE__\s*=\s*(\{.*?\});`)
	matches := stateRe.FindStringSubmatch(htmlContent)

	if len(matches) < 2 {
		fmt.Println("No match found")
		return
	}

	fmt.Printf("Match found. Length: %d\n", len(matches[1]))
	fmt.Printf("Start: %s\n", matches[1][:50])
	fmt.Printf("Start Hex: %x\n", matches[1][:10])
	fmt.Printf("End: %s\n", matches[1][len(matches[1])-50:])

	// Check for undefined
	undefRe := regexp.MustCompile(`undefined`)
	loc := undefRe.FindStringIndex(matches[1])
	if loc != nil {
		fmt.Printf("Found 'undefined' at index %d\n", loc[0])
		start := loc[0] - 20
		if start < 0 {
			start = 0
		}
		end := loc[1] + 20
		if end > len(matches[1]) {
			end = len(matches[1])
		}
		fmt.Printf("Context: %s\n", matches[1][start:end])
	}

	jsonStr := matches[1]
	jsonStr = regexp.MustCompile(`undefined`).ReplaceAllString(jsonStr, "null")

	var state map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &state); err != nil {
		fmt.Printf("JSON Error: %v\n", err)
	} else {
		fmt.Println("JSON Parsed Successfully")
		// Check key content
		if reader, ok := state["reader"].(map[string]interface{}); ok {
			if chapterData, ok := reader["chapterData"].(map[string]interface{}); ok {
				fmt.Printf("Title: %s\n", chapterData["title"])
				content := chapterData["content"].(string)
				fmt.Printf("Content Length: %d\n", len(content))
			}
		}
	}
}
