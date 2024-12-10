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

// WorkItemUpdate represents the payload to update a work item
type WorkItemUpdate struct {
	Op    string `json:"op"`
	Path  string `json:"path"`
	Value string `json:"value"`
}

func main() {
	// Check if the user passed any arguments
	if len(os.Args) < 2 {
		log.Fatal("Usage: main.exe <username>")
	}

	// Username to filter work items
	userName := os.Args[1]

	// Hardcoded values
	pat := "<P_A_T>"
	organization := "Olopo"
	project := "ERP"

	// Fetch work items assigned to the user
	workItems, err := listWorkItems(pat, organization, project, userName)
	if err != nil {
		log.Fatalf("Failed to list work items: %v", err)
	}

	// Log the total count of work items and the username
	fmt.Printf("Total work items assigned to %s: %d\n", userName, len(workItems))

	// Handle the case when no work items are found
	if len(workItems) == 0 {
		fmt.Println("No work items found for the specified user.")
		return
	}

	// Attempt to close each work item
	fmt.Println("Closing work items assigned to", userName, "...")
	for _, id := range workItems {
		err := closeWorkItem(pat, organization, project, id)
		if err != nil {
			log.Printf("Failed to close work item %d: %v\n", id, err)
		} else {
			fmt.Printf("Work item %d closed successfully.\n", id)
		}
	}
}

func closeWorkItem(pat, organization, project string, workItemID int) error {
	url := fmt.Sprintf("https://dev.azure.com/%s/%s/_apis/wit/workitems/%d?api-version=7.0", organization, project, workItemID)

	// Payload to set the work item state to "Closed"
	payload := []WorkItemUpdate{
		{
			Op:    "add",
			Path:  "/fields/System.State",
			Value: "Closed",
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequest(http.MethodPatch, url, bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json-patch+json")
	req.SetBasicAuth("", pat)

	// Execute the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}

func listWorkItems(pat, organization, project, userName string) ([]int, error) {
	url := fmt.Sprintf("https://dev.azure.com/%s/%s/_apis/wit/wiql?api-version=7.0", organization, project)

	// WIQL query to find work items assigned to the user
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

	// Parse response
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
