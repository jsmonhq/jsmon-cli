package handlers

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/jsmonhq/jsmon-cli/api"
)

// countItem represents a label-value pair for sorting
type countItem struct {
	label string
	value int
}

// HandleCount displays the total count analysis for a workspace
func HandleCount(workspaceID, apiKey string, headers map[string]string, runID string) {
	client := api.NewClient(apiKey, headers)

	countAnalysis, err := client.GetTotalCountAnalysis(workspaceID, runID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%sError fetching count analysis: %v%s\n", ColorRed, err, ColorReset)
		os.Exit(1)
	}

	// Display the counts in a formatted way
	fmt.Printf("\n%sWorkspace Count Analysis%s\n", ColorGreen, ColorReset)
	fmt.Println(strings.Repeat("=", 70))

	// Helper function to format a count line
	formatCount := func(label string, value int) {
		fmt.Printf("  %-42s %10d\n", label+":", value)
	}

	// Helper function to sort and display counts
	displaySortedCounts := func(items []countItem) {
		// Sort by value (descending - highest first)
		sort.Slice(items, func(i, j int) bool {
			return items[i].value > items[j].value
		})
		// Display sorted items
		for _, item := range items {
			formatCount(item.label, item.value)
		}
	}

	// 📊 General Counts (sorted by count)
	fmt.Printf("\n%s📊 General Counts%s\n", ColorGreen, ColorReset)
	displaySortedCounts([]countItem{
		{"Total Documents", countAnalysis.TotalDocuments},
		{"Total Domains", countAnalysis.TotalDomains},
		{"Total Emails", countAnalysis.TotalEmails},
		{"Total IP Addresses", countAnalysis.TotalIpAddresses},
		{"Total IPv4 Addresses", countAnalysis.TotalIpv4Addresses},
		{"Total JS URLs", countAnalysis.TotalJsUrls},
		{"Total URLs", countAnalysis.TotalUrls},
	})

	// 🔐 Security & Authentication (sorted by count)
	fmt.Printf("\n%s🔐 Security & Authentication%s\n", ColorGreen, ColorReset)
	displaySortedCounts([]countItem{
		{"Total API Paths", countAnalysis.TotalApiPaths},
		{"Total Extracted Parameters", countAnalysis.TotalExtractedParameters},
		{"Total JWT Tokens", countAnalysis.TotalJwtTokens},
	})

	// 📦 Dependencies & Modules (sorted by count)
	fmt.Printf("\n%s📦 Dependencies & Modules%s\n", ColorGreen, ColorReset)
	displaySortedCounts([]countItem{
		{"Total GUIDs", countAnalysis.TotalGuids},
		{"Total Node Modules", countAnalysis.TotalNodeModules},
		{"Total Valid Node Modules", countAnalysis.TotalValidNodeModules},
	})

	// 🌐 GraphQL (sorted by count)
	fmt.Printf("\n%s🌐 GraphQL%s\n", ColorGreen, ColorReset)
	displaySortedCounts([]countItem{
		{"Total GQL Fragments", countAnalysis.TotalGqlFragments},
		{"Total GQL Mutations", countAnalysis.TotalGqlMutations},
		{"Total GQL Queries", countAnalysis.TotalGqlQueries},
	})

	// ☁️ AWS Assets (sorted by count)
	fmt.Printf("\n%s☁️  AWS Assets%s\n", ColorGreen, ColorReset)
	displaySortedCounts([]countItem{
		{"Total Amplify Domains", countAnalysis.TotalAmplifyDomains},
		{"Total API Gateway Endpoints", countAnalysis.TotalApiGatewayEndpoints},
		{"Total AppSync Endpoints", countAnalysis.TotalAppSyncEndpoints},
		{"Total AWS Assets", countAnalysis.TotalAwsAssets},
		{"Total CloudFormation Endpoints", countAnalysis.TotalCloudFormationEndpoints},
		{"Total CloudFront Domains", countAnalysis.TotalCloudFrontDomains},
		{"Total CloudWatch Endpoints", countAnalysis.TotalCloudWatchEndpoints},
		{"Total Cognito Auth Domains", countAnalysis.TotalCognitoAuthDomains},
		{"Total Cognito Endpoints", countAnalysis.TotalCognitoEndpoints},
		{"Total Cognito Identity Pool IDs", countAnalysis.TotalCognitoIdentityPoolIDs},
		{"Total Cognito User Pool IDs", countAnalysis.TotalCognitoUserPoolIDs},
		{"Total Container Endpoints", countAnalysis.TotalContainerEndpoints},
		{"Total EC2 Instances", countAnalysis.TotalEc2Instances},
		{"Total ELB Endpoints", countAnalysis.TotalElbEndpoints},
		{"Total IoT Endpoints", countAnalysis.TotalIotEndpoints},
		{"Total Kinesis Endpoints", countAnalysis.TotalKinesisEndpoints},
		{"Total Lambda Functions", countAnalysis.TotalLambdaFunctions},
		{"Total OpenSearch Domains", countAnalysis.TotalOpenSearchDomains},
		{"Total Other AWS Endpoints", countAnalysis.TotalOtherAWSEndpoints},
		{"Total RDS Instances", countAnalysis.TotalRdsInstances},
		{"Total S3 Buckets", countAnalysis.TotalS3Buckets},
		{"Total S3 Domains", countAnalysis.TotalS3Domains},
		{"Total S3 Domains (Invalid)", countAnalysis.TotalS3DomainsInvalid},
		{"Total STS Endpoints", countAnalysis.TotalStsEndpoints},
		{"Total Transfer Endpoints", countAnalysis.TotalTransferEndpoints},
		{"Total Work Services", countAnalysis.TotalWorkServices},
	})

	// 🔗 URLs & Links (sorted by count)
	fmt.Printf("\n%s🔗 URLs & Links%s\n", ColorGreen, ColorReset)
	displaySortedCounts([]countItem{
		{"Total Extracted Domains Status", countAnalysis.TotalExtractedDomainsStatus},
		{"Total File Extension URLs", countAnalysis.TotalFileExtensionUrls},
		{"Total Filtered Port URLs", countAnalysis.TotalFilteredPortUrls},
		{"Total Localhost URLs", countAnalysis.TotalLocalhostUrls},
		{"Total Query Params URLs", countAnalysis.TotalQueryParamsUrls},
		{"Total Social Media URLs", countAnalysis.TotalSocialMediaUrls},
	})

	// ⚙️ Execution & Timing (sorted by count)
	fmt.Printf("\n%s⚙️  Execution & Timing%s\n", ColorGreen, ColorReset)
	displaySortedCounts([]countItem{
		{"Total Exec Data", countAnalysis.TotalExecData},
		{"Total setInterval Calls", countAnalysis.TotalSetIntervalCalls},
		{"Total setTimeout Calls", countAnalysis.TotalSetTimeoutCalls},
	})

	// 🛡️ Vulnerabilities (sorted by count)
	fmt.Printf("\n%s🛡️  Vulnerabilities%s\n", ColorGreen, ColorReset)
	displaySortedCounts([]countItem{
		{"Total Client-Side SQLi", countAnalysis.TotalClientSideSQLi},
		{"Total DOM-Based Ajax Header Manipulation", countAnalysis.TotalDomBasedAjaxHeaderManipulation},
		{"Total DOM-Based Cookie Manipulation", countAnalysis.TotalDomBasedCookieManipulation},
		{"Total DOM-Based DoS", countAnalysis.TotalDomBasedDOS},
		{"Total DOM-Based File Path Manipulation", countAnalysis.TotalDomBasedFilePathManipulation},
		{"Total DOM-Based JavaScript Injection", countAnalysis.TotalDomBasedJavaScriptInjection},
		{"Total DOM-Based Link Manipulation", countAnalysis.TotalDomBasedLinkManipulation},
		{"Total DOM-Based Open Redirection", countAnalysis.TotalDomBasedOpenRedirection},
		{"Total DOM XSS Potential", countAnalysis.TotalDomXssPotentialVulnerabilities},
		{"Total Vulnerabilities", countAnalysis.TotalVulnerabilities},
	})

	fmt.Println(strings.Repeat("=", 70))
	fmt.Println()
}
