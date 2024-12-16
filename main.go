package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"time"

	"github.com/m3db/prometheus_remote_client_golang/promremote"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
)

// Constants for event types
const (
	EventCreated = "SupportRequestTicketCreated"
	EventUpdated = "SupportRequestTicketUpdated"
)

// Struct for SupportRequestTicket
type SupportRequestTicket struct {
	CreatedDate      time.Time  `json:"CreatedDate"`
	Severity         string     `json:"Severity"`
	Status           string     `json:"Status"`
	IssueTypeId      string     `json:"IssueTypeId"`
	LastModifiedDate *time.Time `json:"LastModifiedDate"`
	EventType        string     `json:"eventType"`
	PartitionKey     string     `json:"partitionKey"`
}

// Global variables for metrics
var (
	totalTickets           int
	highSeverityTickets    int
	mediumSeverityTickets  int
	lowSeverityTickets     int
	closedTickets          int
	openTickets            int
	duplicateTickets       int
	unrespondedTicketCount int
	totalCloseTime         time.Duration
	averageCloseTime       time.Duration

	weeklyTickets        = make(map[string]int)
	monthlyTickets       = make(map[string]int)
	weeklyTicketsByType  = make(map[string]map[string]int) // Week -> IssueTypeId -> Count
	monthlyTicketsByType = make(map[string]map[string]int) // Month -> IssueTypeId -> Count

	ticketsByIssueType   = map[string]int{}
	createdPartitionKeys = []string{}
	updatedPartitionKeys = []string{}
)

// Check if slice contains a value
func contains(slice []string, value string) bool {
	for _, v := range slice {
		if v == value {
			return true
		}
	}
	return false
}

// Process JSON data and calculate metrics
func processJSONData(jsonData []byte) {
	var tickets []SupportRequestTicket
	if err := json.Unmarshal(jsonData, &tickets); err != nil {
		log.Println("Error unmarshalling JSON:", err)
		return
	}

	// Track partition keys for created and updated tickets
	for _, ticket := range tickets {
		if ticket.EventType == EventCreated {
			createdPartitionKeys = append(createdPartitionKeys, ticket.PartitionKey)
		} else if ticket.EventType == EventUpdated {
			updatedPartitionKeys = append(updatedPartitionKeys, ticket.PartitionKey)
		}
	}

	// Calculate unresponded ticket count
	unrespondedTicketCount = 0
	for _, partitionKey := range createdPartitionKeys {
		if !contains(updatedPartitionKeys, partitionKey) {
			unrespondedTicketCount++
		}
	}

	// Process tickets for other metrics
	for _, ticket := range tickets {
		totalTickets++

		// Calculate weekly and monthly tickets
		year, week := ticket.CreatedDate.ISOWeek()
		weekKey := fmt.Sprintf("%d-W%d", year, week)
		weeklyTickets[weekKey]++

		monthKey := ticket.CreatedDate.Format("2006-01")
		monthlyTickets[monthKey]++

		// Update weekly and monthly counts based on IssueTypeId
		if _, ok := weeklyTicketsByType[weekKey]; !ok {
			weeklyTicketsByType[weekKey] = make(map[string]int)
		}
		weeklyTicketsByType[weekKey][ticket.IssueTypeId]++

		if _, ok := monthlyTicketsByType[monthKey]; !ok {
			monthlyTicketsByType[monthKey] = make(map[string]int)
		}
		monthlyTicketsByType[monthKey][ticket.IssueTypeId]++

		// Categorize severity and status
		switch ticket.Severity {
		case "High":
			highSeverityTickets++
		case "Medium":
			mediumSeverityTickets++
		case "Low":
			lowSeverityTickets++
		}
		switch ticket.Status {
		case "Closed", "close":
			closedTickets++
			if ticket.LastModifiedDate != nil {
				closeTime := ticket.LastModifiedDate.Sub(ticket.CreatedDate)
				totalCloseTime += closeTime
			}
		case "Open", "open":
			openTickets++
		default:
			duplicateTickets++
		}

		// Increment tickets by IssueTypeId
		ticketsByIssueType[ticket.IssueTypeId]++
	}

	if closedTickets > 0 {
		averageCloseTime = totalCloseTime / time.Duration(closedTickets)
	}
}

// Create Prometheus time series data
func createTimeSeries() []promremote.TimeSeries {
	addTs := func(name string, value float64, extraLabels ...promremote.Label) promremote.TimeSeries {
		labels := append([]promremote.Label{
			{Name: "__name__", Value: name},
			{Name: "product_id", Value: "admiral"},
			{Name: "region", Value: "westeurope"},
			{Name: "provider", Value: "azure"},
			{Name: "env", Value: "dev"},
			{Name: "app", Value: "admiral"},
		}, extraLabels...)

		return promremote.TimeSeries{
			Labels: labels,
			Datapoint: promremote.Datapoint{
				Timestamp: time.Now(),
				Value:     value,
			},
		}
	}

	// Prepare time series for metrics
	series := []promremote.TimeSeries{
		addTs("total_tickets", float64(totalTickets)),
		addTs("tickets_by_severity", float64(highSeverityTickets), promremote.Label{Name: "severity", Value: "High"}),
		addTs("tickets_by_severity", float64(mediumSeverityTickets), promremote.Label{Name: "severity", Value: "Medium"}),
		addTs("tickets_by_severity", float64(lowSeverityTickets), promremote.Label{Name: "severity", Value: "Low"}),
		addTs("closed_tickets", float64(closedTickets)),
		addTs("open_tickets", float64(openTickets)),
		addTs("duplicate_tickets", float64(duplicateTickets)),
		addTs("average_close_time_seconds", averageCloseTime.Seconds()),
		addTs("unresponded_tickets", float64(unrespondedTicketCount)),
	}

	// Add weekly ticket counts
	for week, count := range weeklyTickets {
		series = append(series, addTs("tickets_per_week", float64(count), promremote.Label{Name: "week", Value: week}))
	}

	// Add monthly ticket counts
	for month, count := range monthlyTickets {
		series = append(series, addTs("tickets_per_month", float64(count), promremote.Label{Name: "month", Value: month}))
	}

	// Add tickets by IssueTypeId
	for issueType, count := range ticketsByIssueType {
		series = append(series, addTs("tickets_by_issuetype", float64(count), promremote.Label{Name: "issue_type", Value: issueType}))
	}

	// Add weekly tickets per IssueTypeId
	for week, issueCounts := range weeklyTicketsByType {
		for issueType, count := range issueCounts {
			series = append(series, addTs("tickets_per_week_by_issuetype", float64(count), promremote.Label{Name: "week", Value: week}, promremote.Label{Name: "issue_type", Value: issueType}))
		}
	}

	// Add monthly tickets per IssueTypeId
	for month, issueCounts := range monthlyTicketsByType {
		for issueType, count := range issueCounts {
			series = append(series, addTs("tickets_per_month_by_issuetype", float64(count), promremote.Label{Name: "month", Value: month}, promremote.Label{Name: "issue_type", Value: issueType}))
		}
	}

	return series
}

// Send metrics to Prometheus
func sendMetrics() {
	oauthConfig := &clientcredentials.Config{
		ClientID:     os.Getenv("AZURE_CLIENT_ID"),
		ClientSecret: os.Getenv("AZURE_CLIENT_SECRET"),
		Scopes:       []string{"api://ingestion.pensieve/.default"},
		TokenURL:     "https://login.microsoftonline.com/maersk.com/oauth2/v2.0/token",
		AuthStyle:    oauth2.AuthStyleInParams,
	}

	httpClient := oauthConfig.Client(context.TODO())

	promClient, err := promremote.NewClient(promremote.Config{
		WriteURL:          "https://telemetry.pensieve.maersk-digital.net/api/v1/push",
		HTTPClientTimeout: 30,
		HTTPClient:        httpClient,
		UserAgent:         "Metrics Writer 1.0",
	})

	if err != nil {
		log.Println("Error creating Prometheus push client:", err)
		return
	}

	timeSeriesList := createTimeSeries()

	result, err := promClient.WriteTimeSeries(context.TODO(), timeSeriesList, promremote.WriteOptions{})
	log.Println(result, err)
}

func main() {
	// Load and process JSON data
	jsonFile, err := os.Open("source_data.json")
	if err != nil {
		log.Println("Error opening JSON file:", err)
		return
	}
	defer jsonFile.Close()

	jsonData, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		log.Println("Error reading JSON file:", err)
		return
	}

	processJSONData(jsonData)
	sendMetrics()
}
