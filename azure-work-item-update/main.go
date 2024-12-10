package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

// WiqlQuery represents the structure of a WIQL query payload
type WiqlQuery struct {
	Query string `json:"query"`
}

// WiqlResponse represents the response structure for a WIQL query
type WiqlResponse struct {
	WorkItems []struct {
		ID int `json:"id"`
	} `json:"workItems"`
}

func main() {
	// Replace with actual values or fetch from environment variables
	pat := os.Getenv("AZURE_DEVOPS_PAT")
	organization := os.Getenv("AZURE_DEVOPS_ORG")
	project := os.Getenv("AZURE_DEVOPS_PROJECT")
	userName := os.Args[1] // Username passed as a command-line argument

	if pat == "" || organization == "" || project == "" {
		log.Fatal("Environment variables AZURE_DEVOPS_PAT, AZURE_DEVOPS_ORG, and AZURE_DEVOPS_PROJECT must be set")
	}

	workItems, err := listWorkItems(pat, organization, project, userName)
	if err != nil {
		log.Fatalf("Failed to list work items: %v", err)
	}

	if len(workItems) == 0 {
		fmt.Println("No work items found for the specified user.")
		return
	}

	fmt.Println("Work Items assigned to", userName, ":")
	for _, id := range workItems {
		fmt.Printf("Work Item ID: %d\n", id)
	}
}

func listWorkItems(pat, organization, project, userName string) ([]int, error) {
	url := fmt.Sprintf("https://dev.azure.com/%s/%s/_apis/wit/wiql?api-version=7.0", organization, project)

	// WIQL query to fetch work items assigned to the specific user
	query := WiqlQuery{
		Query: fmt.Sprintf(`SELECT [System.Id] FROM workitems WHERE [System.AssignedTo] CONTAINS '%s' AND [System.State] <> 'Closed'`, userName),
	}

	body, err := json.Marshal(query)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal WIQL query: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth("", pat)

	// Execute the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// Parse the response
	responseBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var wiqlResponse WiqlResponse
	err = json.Unmarshal(responseBody, &wiqlResponse)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	// Extract work item IDs
	var workItemIDs []int
	for _, item := range wiqlResponse.WorkItems {
		workItemIDs = append(workItemIDs, item.ID)
	}

	return workItemIDs, nil
}
