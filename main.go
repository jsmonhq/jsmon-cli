package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/jsmonhq/jsmon-cli/config"
	"github.com/jsmonhq/jsmon-cli/handlers"
)

func main() {
	// Store original args
	originalArgs := make([]string, len(os.Args))
	copy(originalArgs, os.Args)

	// Parse headers from command line arguments first (before flag.Parse)
	// This extracts -H flags and removes them from os.Args so flag.Parse() doesn't see them
	headers, filteredArgs := parseHeadersAndFilterArgs()

	// Extract --urls, --domains, --files, -secrets, -recon, -rsearch, and -filters flags and their values before flag.Parse() to prevent them from consuming -wksp
	urlsFlagProvided := false
	urlsValue := ""
	domainsFlagProvided := false
	domainsValue := ""
	filesFlagProvided := false
	filesValue := ""
	secretsFlagProvided := false
	secretsValue := ""
	reconFlagProvided := false
	reconValue := ""
	rsearchFlagProvided := false
	rsearchValue := ""
	filtersFlagProvided := false
	filtersValue := ""

	// First, check originalArgs to find --urls, --domains, --files, -secrets, -recon, -rsearch, and -filters and their values
	for i, arg := range originalArgs {
		if arg == "--urls" {
			urlsFlagProvided = true
			// Check if next argument is a value (starts with "page=")
			if i+1 < len(originalArgs) && strings.HasPrefix(originalArgs[i+1], "page=") {
				urlsValue = originalArgs[i+1]
			}
		}
		if arg == "--domains" {
			domainsFlagProvided = true
			// Check if next argument is a value (starts with "page=")
			if i+1 < len(originalArgs) && strings.HasPrefix(originalArgs[i+1], "page=") {
				domainsValue = originalArgs[i+1]
			}
		}
		if arg == "--files" {
			filesFlagProvided = true
			// Check if next argument is a value (starts with "page=")
			if i+1 < len(originalArgs) && strings.HasPrefix(originalArgs[i+1], "page=") {
				filesValue = originalArgs[i+1]
			}
		}
		if arg == "-secrets" {
			secretsFlagProvided = true
			// Check if next argument is a value (starts with "page=")
			if i+1 < len(originalArgs) && strings.HasPrefix(originalArgs[i+1], "page=") {
				secretsValue = originalArgs[i+1]
			}
		}
		if arg == "-recon" {
			reconFlagProvided = true
			// Check if next argument is a value (contains "field=" and "page=")
			if i+1 < len(originalArgs) {
				reconValue = originalArgs[i+1]
			}
		}
		if arg == "-rsearch" {
			rsearchFlagProvided = true
			// Check if next argument is a value (contains "=")
			if i+1 < len(originalArgs) {
				rsearchValue = originalArgs[i+1]
				// If the value contains newlines or special characters, it might be split across multiple args
				// Check if we need to join more arguments
				if strings.Contains(rsearchValue, "=") && !strings.Contains(rsearchValue, "\"") {
					// Value might continue in subsequent args if it contains spaces/newlines
					// For now, just use the next arg as-is
				}
			}
		}
		if arg == "-filters" {
			filtersFlagProvided = true
			// Check if next argument is a value (contains "fieldname=" and optionally "page=")
			if i+1 < len(originalArgs) {
				filtersValue = originalArgs[i+1]
			}
		}
	}

	// Now build newFilteredArgs without --scannedurls, --domains and their page= values
	// Check if filteredArgs is empty (no arguments provided)
	if len(filteredArgs) == 0 {
		// No arguments - show help (logo will be shown in showUsage if not silent)
		showUsage()
		os.Exit(0)
	}
	// filteredArgs doesn't include program name, so we start with an empty slice
	newFilteredArgs := []string{}

	for i := 0; i < len(filteredArgs); i++ {
		arg := filteredArgs[i]
		if arg == "--urls" {
			// Skip --urls flag
			// Also skip the next arg if it's the page= value
			if i+1 < len(filteredArgs) && strings.HasPrefix(filteredArgs[i+1], "page=") {
				i++ // Skip the page= value
			}
			continue
		}
		if arg == "--domains" {
			// Skip --domains flag
			// Also skip the next arg if it's the page= value
			if i+1 < len(filteredArgs) && strings.HasPrefix(filteredArgs[i+1], "page=") {
				i++ // Skip the page= value
			}
			continue
		}
		if arg == "--files" {
			// Skip --files flag
			// Also skip the next arg if it's the page= value
			if i+1 < len(filteredArgs) && strings.HasPrefix(filteredArgs[i+1], "page=") {
				i++ // Skip the page= value
			}
			continue
		}
		if arg == "-secrets" {
			// Skip -secrets flag
			// Also skip the next arg if it's the page= value
			if i+1 < len(filteredArgs) && strings.HasPrefix(filteredArgs[i+1], "page=") {
				i++ // Skip the page= value
			}
			continue
		}
		if arg == "-recon" {
			// Skip -recon flag
			// Also skip the next arg if it contains "field=" (the value is in quotes, so it's a single arg)
			if i+1 < len(filteredArgs) && strings.Contains(filteredArgs[i+1], "field=") {
				i++ // Skip the field=... page=... value
			}
			continue
		}
		if arg == "-rsearch" {
			// Skip -rsearch flag
			// Also skip the next arg if it contains "=" (the value is in quotes, so it's a single arg)
			if i+1 < len(filteredArgs) && strings.Contains(filteredArgs[i+1], "=") {
				i++ // Skip the field=value
			}
			continue
		}
		if arg == "-filters" {
			// Skip -filters flag
			// Also skip the next arg if it contains "=" (the value is in quotes, so it's a single arg)
			// Format: "fieldname=keyword" or "fieldname=keyword page=1"
			if i+1 < len(filteredArgs) && strings.Contains(filteredArgs[i+1], "=") {
				i++ // Skip the fieldname=... page=... value
			}
			continue
		}
		// Skip page= value if it's not immediately after --urls, --domains, --files, or -secrets (shouldn't happen, but just in case)
		if strings.HasPrefix(arg, "page=") && (urlsFlagProvided || domainsFlagProvided || filesFlagProvided || secretsFlagProvided) {
			continue
		}
		// Skip field=... page=... value if it's not immediately after -recon (shouldn't happen, but just in case)
		if strings.Contains(arg, "field=") && reconFlagProvided {
			continue
		}
		// Skip fieldname=... page=... value if it's not immediately after -filters (shouldn't happen, but just in case)
		if strings.Contains(arg, "fieldname=") && filtersFlagProvided {
			continue
		}
		newFilteredArgs = append(newFilteredArgs, arg)
	}

	// Temporarily replace os.Args with filtered version for flag.Parse()
	// Include program name + filtered args
	os.Args = append([]string{originalArgs[0]}, newFilteredArgs...)
	defer func() { os.Args = originalArgs }()

	// Define flags
	urlFlag := flag.String("u", "", "Input URL to scan")
	domainFlag := flag.String("d", "", "Input domain to scan")
	fileFlag := flag.String("f", "", "Input file of URLs to scan (one URL per line)")
	createWorkspaceFlag := flag.String("cw", "", "Create a new workspace (use --create-workspace as alternative)")
	createWorkspaceFlagAlt := flag.String("create-workspace", "", "Create a new workspace")
	apiKeyFlag := flag.String("key", "", "JSMon API key (or set in ~/.jsmon/credentials)")
	workspaceIDFlag := flag.String("wksp", "", "Workspace ID (or set in ~/.jsmon/credentials)")
	resumeFlag := flag.String("resume", "", "Resume from a previous scan using resume.cfg file")
	countFlag := flag.Bool("count", false, "Show count analysis for the workspace")
	runIDFlag := flag.String("runId", "", "Run ID for count analysis (optional)")
	workspacesFlag := flag.Bool("workspaces", false, "List all workspaces")
	_ = flag.Bool("silent", false, "Suppress logo output") // Flag is checked in showUsage()
	helpFlag := flag.Bool("h", false, "Show help message")
	helpFlagAlt := flag.Bool("help", false, "Show help message")

	// Check if silent flag is in args before parsing
	silentInArgs := false
	for _, arg := range originalArgs {
		if arg == "-silent" {
			silentInArgs = true
			break
		}
	}

	// Set custom usage function to show our custom help menu
	flag.Usage = func() {
		// Don't print logo here - showUsage() will handle it
		showUsage()
	}

	flag.Parse()

	// Check if help flag is set
	showHelp := *helpFlag || *helpFlagAlt

	// Handle help flag (showUsage will handle logo printing)
	if showHelp {
		showUsage()
		os.Exit(0)
	}

	// Show logo for regular commands unless silent flag is set (to stderr so it doesn't interfere with stdout output)
	// silentInArgs was already checked earlier, so reuse it
	if !silentInArgs {
		fmt.Fprint(os.Stderr, LogoColor)
	}

	// Get API key: flag > credentials file > env var
	apiKey := *apiKeyFlag
	if apiKey == "" {
		// Try to read from credentials file
		credAPIKey, err := config.ReadCredentials()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Could not read credentials file: %v\n", err)
		}
		if credAPIKey != "" {
			apiKey = credAPIKey
		}
	}
	if apiKey == "" {
		apiKey = os.Getenv("JSMON_API_KEY")
	}

	// Get workspace ID: flag > env var (must be provided, not in credentials file)
	workspaceID := *workspaceIDFlag
	if workspaceID == "" {
		workspaceID = os.Getenv("JSMON_WORKSPACE_ID")
	}

	// Handle create workspace flag (check both -cw and --create-workspace)
	workspaceName := *createWorkspaceFlag
	if workspaceName == "" {
		workspaceName = *createWorkspaceFlagAlt
	}

	// Route to appropriate handler
	if urlsFlagProvided {
		// Fetch scanned URLs
		if apiKey == "" {
			fmt.Fprintf(os.Stderr, "Error: API key is required. Use -key flag, add to ~/.jsmon/credentials, or set JSMON_API_KEY environment variable\n")
			os.Exit(1)
		}
		if workspaceID == "" {
			fmt.Fprintf(os.Stderr, "Error: Workspace ID is required. Use -wksp flag or set JSMON_WORKSPACE_ID environment variable\n")
			os.Exit(1)
		}
		// Parse page parameter (format: page=<number>), default to page=1 if not provided
		page := 1
		if urlsValue != "" {
			if strings.HasPrefix(urlsValue, "page=") {
				pageStr := strings.TrimPrefix(urlsValue, "page=")
				if parsedPage, err := strconv.Atoi(pageStr); err == nil && parsedPage > 0 {
					page = parsedPage
				} else {
					fmt.Fprintf(os.Stderr, "Error: Invalid page number format. Use --urls page=<number> (e.g., --urls page=2)\n\n")
					showUsage()
					os.Exit(1)
				}
			} else {
				fmt.Fprintf(os.Stderr, "Error: Invalid format. Use --urls page=<number> (e.g., --urls page=2)\n\n")
				showUsage()
				os.Exit(1)
			}
		}
		handlers.HandleJSURLs(workspaceID, apiKey, headers, page, "", "", "")
	} else if domainsFlagProvided {
		// Fetch domain scans
		if apiKey == "" {
			fmt.Fprintf(os.Stderr, "Error: API key is required. Use -key flag, add to ~/.jsmon/credentials, or set JSMON_API_KEY environment variable\n")
			os.Exit(1)
		}
		if workspaceID == "" {
			fmt.Fprintf(os.Stderr, "Error: Workspace ID is required. Use -wksp flag or set JSMON_WORKSPACE_ID environment variable\n")
			os.Exit(1)
		}
		// Parse page parameter (format: page=<number>), default to page=1 if not provided
		page := 1
		if domainsValue != "" {
			if strings.HasPrefix(domainsValue, "page=") {
				pageStr := strings.TrimPrefix(domainsValue, "page=")
				if parsedPage, err := strconv.Atoi(pageStr); err == nil && parsedPage > 0 {
					page = parsedPage
				} else {
					fmt.Fprintf(os.Stderr, "Error: Invalid page number format. Use --domains page=<number> (e.g., --domains page=2)\n\n")
					showUsage()
					os.Exit(1)
				}
			} else {
				fmt.Fprintf(os.Stderr, "Error: Invalid format. Use --domains page=<number> (e.g., --domains page=2)\n\n")
				showUsage()
				os.Exit(1)
			}
		}
		handlers.HandleDomains(workspaceID, apiKey, headers, page, "", "", "", "", "", "", "100", "")
	} else if filesFlagProvided {
		// Fetch file scans
		if apiKey == "" {
			fmt.Fprintf(os.Stderr, "Error: API key is required. Use -key flag, add to ~/.jsmon/credentials, or set JSMON_API_KEY environment variable\n")
			os.Exit(1)
		}
		if workspaceID == "" {
			fmt.Fprintf(os.Stderr, "Error: Workspace ID is required. Use -wksp flag or set JSMON_WORKSPACE_ID environment variable\n")
			os.Exit(1)
		}
		// Parse page parameter (format: page=<number>), default to page=1 if not provided
		page := 1
		if filesValue != "" {
			if strings.HasPrefix(filesValue, "page=") {
				pageStr := strings.TrimPrefix(filesValue, "page=")
				if parsedPage, err := strconv.Atoi(pageStr); err == nil && parsedPage > 0 {
					page = parsedPage
				} else {
					fmt.Fprintf(os.Stderr, "Error: Invalid page number format. Use --files page=<number> (e.g., --files page=2)\n\n")
					showUsage()
					os.Exit(1)
				}
			} else {
				fmt.Fprintf(os.Stderr, "Error: Invalid format. Use --files page=<number> (e.g., --files page=2)\n\n")
				showUsage()
				os.Exit(1)
			}
		}
		handlers.HandleFileScans(workspaceID, apiKey, headers, page, "", "", "", "", "", "", "100", "")
	} else if secretsFlagProvided {
		// Fetch secrets
		if apiKey == "" {
			fmt.Fprintf(os.Stderr, "Error: API key is required. Use -key flag, add to ~/.jsmon/credentials, or set JSMON_API_KEY environment variable\n")
			os.Exit(1)
		}
		if workspaceID == "" {
			fmt.Fprintf(os.Stderr, "Error: Workspace ID is required. Use -wksp flag or set JSMON_WORKSPACE_ID environment variable\n")
			os.Exit(1)
		}
		// Parse page parameter (format: page=<number>), default to page=1 if not provided
		page := 1
		if secretsValue != "" {
			if strings.HasPrefix(secretsValue, "page=") {
				pageStr := strings.TrimPrefix(secretsValue, "page=")
				if parsedPage, err := strconv.Atoi(pageStr); err == nil && parsedPage > 0 {
					page = parsedPage
				} else {
					fmt.Fprintf(os.Stderr, "Error: Invalid page number format. Use -secrets page=<number> (e.g., -secrets page=2)\n\n")
					showUsage()
					os.Exit(1)
				}
			} else {
				fmt.Fprintf(os.Stderr, "Error: Invalid format. Use -secrets page=<number> (e.g., -secrets page=2)\n\n")
				showUsage()
				os.Exit(1)
			}
		}
		handlers.HandleSecrets(workspaceID, apiKey, headers, page, "", "", "", "", "100", "")
	} else if reconFlagProvided {
		// Check if field parameter is provided
		if reconValue == "" {
			fmt.Fprintf(os.Stderr, "Error: Field name is required. Use -recon \"field=<name>\" or -recon \"field=<name> page=<number>\" (e.g., -recon \"field=emails\" or -recon \"field=emails page=3\")\n\n")
			showUsage()
			os.Exit(1)
		}

		// Parse field and page from reconValue (format: "field=emails page=3" or just "field=emails")
		field := ""
		page := 1

		parts := strings.Fields(reconValue)
		for _, part := range parts {
			if strings.HasPrefix(part, "field=") {
				field = strings.TrimPrefix(part, "field=")
			}
			if strings.HasPrefix(part, "page=") {
				pageStr := strings.TrimPrefix(part, "page=")
				if parsedPage, err := strconv.Atoi(pageStr); err == nil && parsedPage > 0 {
					page = parsedPage
				} else {
					fmt.Fprintf(os.Stderr, "Error: Invalid page number format. Use -recon \"field=<name> page=<number>\" (e.g., -recon \"field=emails page=3\")\n\n")
					showUsage()
					os.Exit(1)
				}
			}
		}

		if field == "" {
			fmt.Fprintf(os.Stderr, "Error: Field name is required. Use -recon \"field=<name>\" or -recon \"field=<name> page=<number>\" (e.g., -recon \"field=emails\" or -recon \"field=emails page=3\")\n\n")
			showUsage()
			os.Exit(1)
		}

		// Fetch reconnaissance data
		if apiKey == "" {
			fmt.Fprintf(os.Stderr, "Error: API key is required. Use -key flag, add to ~/.jsmon/credentials, or set JSMON_API_KEY environment variable\n")
			os.Exit(1)
		}
		if workspaceID == "" {
			fmt.Fprintf(os.Stderr, "Error: Workspace ID is required. Use -wksp flag or set JSMON_WORKSPACE_ID environment variable\n")
			os.Exit(1)
		}
		handlers.HandleJSIntelligence(workspaceID, apiKey, headers, field, page)
	} else if rsearchFlagProvided {
		// Perform reverse search
		if apiKey == "" {
			fmt.Fprintf(os.Stderr, "Error: API key is required. Use -key flag, add to ~/.jsmon/credentials, or set JSMON_API_KEY environment variable\n")
			os.Exit(1)
		}
		if workspaceID == "" {
			fmt.Fprintf(os.Stderr, "Error: Workspace ID is required. Use -wksp flag or set JSMON_WORKSPACE_ID environment variable\n")
			os.Exit(1)
		}
		if rsearchValue == "" {
			fmt.Fprintf(os.Stderr, "Error: Field and value are required. Use -rsearch \"<field>=<value>\" (e.g., -rsearch \"apipaths=@azure/msal-browser\")\n\n")
			showUsage()
			os.Exit(1)
		}

		// Parse field and value from rsearchValue (format: "field=value")
		parts := strings.SplitN(rsearchValue, "=", 2)
		if len(parts) != 2 {
			fmt.Fprintf(os.Stderr, "Error: Invalid format. Use -rsearch \"<field>=<value>\" (e.g., -rsearch \"apipaths=@azure/msal-browser\")\n\n")
			showUsage()
			os.Exit(1)
		}

		field := strings.TrimSpace(parts[0])
		searchValue := strings.TrimSpace(parts[1])

		if field == "" || searchValue == "" {
			fmt.Fprintf(os.Stderr, "Error: Field and value cannot be empty. Use -rsearch \"<field>=<value>\" (e.g., -rsearch \"apipaths=@azure/msal-browser\")\n\n")
			showUsage()
			os.Exit(1)
		}

		handlers.HandleReverseSearch(workspaceID, apiKey, headers, field, searchValue)
	} else if filtersFlagProvided {
		// Perform filter search
		if apiKey == "" {
			fmt.Fprintf(os.Stderr, "Error: API key is required. Use -key flag, add to ~/.jsmon/credentials, or set JSMON_API_KEY environment variable\n")
			os.Exit(1)
		}
		if workspaceID == "" {
			fmt.Fprintf(os.Stderr, "Error: Workspace ID is required. Use -wksp flag or set JSMON_WORKSPACE_ID environment variable\n")
			os.Exit(1)
		}
		if filtersValue == "" {
			fmt.Fprintf(os.Stderr, "Error: Field name and keyword are required. Use -filters \"fieldname=<keyword>\" or -filters \"fieldname=<keyword> page=<number>\" (e.g., -filters \"urls=github.com\" or -filters \"urls=github.com page=2\")\n\n")
			showUsage()
			os.Exit(1)
		}

		// Parse fieldname, keyword, and page from filtersValue (format: "urls=github.com page=1" or "urls=github.com")
		parts := strings.Fields(filtersValue)
		fieldname := ""
		keyword := ""
		page := 1

		for _, part := range parts {
			if strings.HasPrefix(part, "page=") {
				pageStr := strings.TrimPrefix(part, "page=")
				if parsedPage, err := strconv.Atoi(pageStr); err == nil && parsedPage > 0 {
					page = parsedPage
				} else {
					fmt.Fprintf(os.Stderr, "Error: Invalid page number format. Use -filters \"fieldname=<keyword> page=<number>\" (e.g., -filters \"urls=github.com page=2\")\n\n")
					showUsage()
					os.Exit(1)
				}
			} else if strings.Contains(part, "=") {
				// This should be fieldname=keyword
				fieldParts := strings.SplitN(part, "=", 2)
				if len(fieldParts) == 2 {
					fieldname = strings.TrimSpace(fieldParts[0])
					keyword = strings.TrimSpace(fieldParts[1])
				}
			}
		}

		if fieldname == "" || keyword == "" {
			fmt.Fprintf(os.Stderr, "Error: Field name and keyword are required. Use -filters \"fieldname=<keyword>\" or -filters \"fieldname=<keyword> page=<number>\" (e.g., -filters \"urls=github.com\" or -filters \"urls=github.com page=2\")\n\n")
			showUsage()
			os.Exit(1)
		}

		// Validate allowed fields
		allowedFields := map[string]bool{
			"jsurls":       true,
			"apipaths":     true,
			"urls":         true,
			"emails":       true,
			"gqlqueries":   true,
			"gqlmutaions":  true,
			"sqlfragments": true,
			"param":        true,
		}

		if !allowedFields[strings.ToLower(fieldname)] {
			fmt.Fprintf(os.Stderr, "Error: Invalid field name. Allowed fields: jsurls, apiPaths, urls, emails, gqlQueries, gqlMutaions, sqlFragments, param\n\n")
			showUsage()
			os.Exit(1)
		}

		handlers.HandleFilter(workspaceID, apiKey, headers, fieldname, keyword, page)
	} else if *workspacesFlag {
		// List all workspaces
		if apiKey == "" {
			fmt.Fprintf(os.Stderr, "Error: API key is required. Use -key flag, add to ~/.jsmon/credentials, or set JSMON_API_KEY environment variable\n")
			os.Exit(1)
		}
		handlers.HandleWorkspaces(apiKey, headers)
	} else if *countFlag {
		// Show count analysis
		if apiKey == "" {
			fmt.Fprintf(os.Stderr, "Error: API key is required. Use -key flag, add to ~/.jsmon/credentials, or set JSMON_API_KEY environment variable\n")
			os.Exit(1)
		}
		if workspaceID == "" {
			fmt.Fprintf(os.Stderr, "Error: Workspace ID is required. Use -wksp flag or set JSMON_WORKSPACE_ID environment variable\n")
			os.Exit(1)
		}
		handlers.HandleCount(workspaceID, apiKey, headers, *runIDFlag)
	} else if workspaceName != "" {
		// Create workspace
		if apiKey == "" {
			fmt.Fprintf(os.Stderr, "Error: API key is required. Use -key flag, add to ~/.jsmon/credentials, or set JSMON_API_KEY environment variable\n")
			os.Exit(1)
		}
		handlers.HandleCreateWorkspace(workspaceName, apiKey, headers)
	} else if *urlFlag != "" {
		// URL upload
		if apiKey == "" {
			fmt.Fprintf(os.Stderr, "Error: API key is required. Use -key flag, add to ~/.jsmon/credentials, or set JSMON_API_KEY environment variable\n")
			os.Exit(1)
		}
		if workspaceID == "" {
			fmt.Fprintf(os.Stderr, "Error: Workspace ID is required. Use -wksp flag or set JSMON_WORKSPACE_ID environment variable\n")
			os.Exit(1)
		}
		handlers.HandleURLUpload(*urlFlag, workspaceID, apiKey, headers)
	} else if *resumeFlag != "" {
		// Resume from previous scan (resume flag takes precedence)
		// API key and workspace ID must be provided via flags/env vars
		if apiKey == "" {
			fmt.Fprintf(os.Stderr, "Error: API key is required for resume. Use -key flag, add to ~/.jsmon/credentials, or set JSMON_API_KEY environment variable\n")
			os.Exit(1)
		}
		if workspaceID == "" {
			fmt.Fprintf(os.Stderr, "Error: Workspace ID is required for resume. Use -wksp flag or set JSMON_WORKSPACE_ID environment variable\n")
			os.Exit(1)
		}
		handlers.HandleResume(*resumeFlag, workspaceID, apiKey, headers)
	} else if *domainFlag != "" {
		// Domain scanning
		if apiKey == "" {
			fmt.Fprintf(os.Stderr, "Error: API key is required. Use -key flag, add to ~/.jsmon/credentials, or set JSMON_API_KEY environment variable\n")
			os.Exit(1)
		}
		if workspaceID == "" {
			fmt.Fprintf(os.Stderr, "Error: Workspace ID is required. Use -wksp flag or set JSMON_WORKSPACE_ID environment variable\n")
			os.Exit(1)
		}
		handlers.HandleDomainScan(*domainFlag, workspaceID, apiKey, "", headers)
	} else if *fileFlag != "" {
		// File upload
		if apiKey == "" {
			fmt.Fprintf(os.Stderr, "Error: API key is required. Use -key flag, add to ~/.jsmon/credentials, or set JSMON_API_KEY environment variable\n")
			os.Exit(1)
		}
		if workspaceID == "" {
			fmt.Fprintf(os.Stderr, "Error: Workspace ID is required. Use -wksp flag or set JSMON_WORKSPACE_ID environment variable\n")
			os.Exit(1)
		}
		handlers.HandleFileUpload(*fileFlag, workspaceID, apiKey, "", headers)
	} else {
		// No valid flag provided - show usage (showUsage will handle logo printing)
		showUsage()
		os.Exit(1)
	}
}

func showUsage() {
	// Check if -silent is in os.Args
	silentInArgs := false
	for _, arg := range os.Args {
		if arg == "-silent" {
			silentInArgs = true
			break
		}
	}

	// Show logo unless silent flag is set
	if !silentInArgs {
		fmt.Fprint(os.Stderr, LogoColor)
	}

	fmt.Fprintf(os.Stderr, "\nUsage: jsmon-cli [OPTIONS]\n\n")

	fmt.Fprintf(os.Stderr, "Input:\n")
	fmt.Fprintf(os.Stderr, "  -u <input>                                  Input URL to scan\n")
	fmt.Fprintf(os.Stderr, "  -d <input>                                  Input domain to scan\n")
	fmt.Fprintf(os.Stderr, "  -f <input>                                  Input file of URLs to scan (one URL per line)\n")
	fmt.Fprintf(os.Stderr, "  -cw <input> | --create-workspace <input>    Create a new workspace\n\n")

	fmt.Fprintf(os.Stderr, "Configuration:\n")
	fmt.Fprintf(os.Stderr, "  -key <input>                                API key (or add the API key to ~/.jsmon/credentials)\n")
	fmt.Fprintf(os.Stderr, "  -wksp <wksp id>                             Workspace ID to scan the target\n")
	fmt.Fprintf(os.Stderr, "  -H <input>                                  Custom HTTP headers to send along with request to scan\n")
	fmt.Fprintf(os.Stderr, "  -resume                                     Resume scan using resume.config\n")
	fmt.Fprintf(os.Stderr, "                                              (resumes from last scan failed due to force stop or API limits)\n")
	fmt.Fprintf(os.Stderr, "  -silent                                     Silent the logo\n\n")

	fmt.Fprintf(os.Stderr, "Scans:\n")
	fmt.Fprintf(os.Stderr, "  -count                                      Show the counts of reconnaissance data and secrets count\n")
	fmt.Fprintf(os.Stderr, "  --urls page=<page number>                   Fetch all scanned URLs (default: page=1)\n")
	fmt.Fprintf(os.Stderr, "  --domains page=<page number>                Fetch all scanned domains (default: page=1)\n")
	fmt.Fprintf(os.Stderr, "  --files page=<page number>                  Fetch all scanned files (default: page=1)\n\n")

	fmt.Fprintf(os.Stderr, "Data:\n")
	fmt.Fprintf(os.Stderr, "  -workspaces                                 Fetch all workspaces\n")
	fmt.Fprintf(os.Stderr, "  -secrets page=<number>                      Fetch all secrets for a workspace (default: page=1)\n")
	fmt.Fprintf(os.Stderr, "  -recon \"field=<name> page=<number>\"         Fetch the reconnaissance data\n")
	fmt.Fprintf(os.Stderr, "                                              Example: -recon \"field=emails page=3\"\n\n")

	fmt.Fprintf(os.Stderr, "Reverse Search:\n")
	fmt.Fprintf(os.Stderr, "  -rsearch \"<field name>=<value>\"             Search the source of the result where it comes from\n")
	fmt.Fprintf(os.Stderr, "                                              Example: -rsearch \"apipaths=@azure/msal-browser\"\n\n")

	fmt.Fprintf(os.Stderr, "Filter:\n")
	fmt.Fprintf(os.Stderr, "  -filters \"<fieldname>=<keyword> page=<number>\"    Match keywords in the field data in reconnaissance results\n")
	fmt.Fprintf(os.Stderr, "                                                    (default: page=1)\n")
	fmt.Fprintf(os.Stderr, "                                                    Example: -filters \"urls=github.com page=2\"\n\n")

	fmt.Fprintf(os.Stderr, "Help:\n")
	fmt.Fprintf(os.Stderr, "  -h, --help                                  Show this help message\n\n")

	fmt.Fprintf(os.Stderr, "Field Names:\n")
	fmt.Fprintf(os.Stderr, "  -recon, -rsearch:\n")
	fmt.Fprintf(os.Stderr, "    apiPaths, urls, extractedDomains, ip, emails, s3Buckets, s3takeovers,gqlQueries, gqlMutaions, sqlFragments, param (extracted parameter),\n")
	fmt.Fprintf(os.Stderr, "    npmPackages, npmConfusion, guids, localhost, activeDomains,inactiveDomains, allAwsAssets, queryParameters, socialUrls,\n")
	fmt.Fprintf(os.Stderr, "    portUrls, extensionUrls\n\n")
	fmt.Fprintf(os.Stderr, "  -filters:\n")
	fmt.Fprintf(os.Stderr, "    jsurls, apiPaths, urls, emails, gqlQueries, gqlMutaions,sqlFragments, param (extracted parameter)\n")
}

// parseHeadersAndFilterArgs parses headers from command line arguments and returns
// both the headers map and a filtered args list without -H flags
// Supports multiple -H flags like: -H "Header1: value1" -H "Header2: value2"
func parseHeadersAndFilterArgs() (map[string]string, []string) {
	headers := make(map[string]string)
	args := os.Args[1:]
	filteredArgs := make([]string, 0, len(args))

	for i := 0; i < len(args); i++ {
		if args[i] == "-H" || args[i] == "--header" {
			if i+1 < len(args) {
				headerValue := args[i+1]
				// Parse header in format "Header-Name: value"
				parts := strings.SplitN(headerValue, ":", 2)
				if len(parts) == 2 {
					key := strings.TrimSpace(parts[0])
					value := strings.TrimSpace(parts[1])
					if key != "" {
						headers[key] = value
					}
				}
				i++ // Skip the next argument as it's the header value
			}
			// Don't add -H and its value to filteredArgs
		} else {
			filteredArgs = append(filteredArgs, args[i])
		}
	}

	return headers, filteredArgs
}
