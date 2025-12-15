/*
 * Copyright 2025 Magnus Gille <mgille@gmail.com>
 */

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"time"
)

func main() {
	hostPtr := flag.String("host", "localhost:8080", "CloudDB host (e.g. localhost:8080)")
	secretPtr := flag.String("secret", "", "Basic Auth Secret")

	// We need to parse flags before subcommands to get host/secret
	// However, standard flag package stops at the first non-flag argument (the subcommand).
	// So we can just parse commonly.
	flag.Parse()

	if len(flag.Args()) < 1 {
		printUsage()
		os.Exit(1)
	}

	client := NewCloudClient(*hostPtr, *secretPtr)
	subcommand := flag.Arg(0)

	switch subcommand {
	case "gchart":
		handleGChart(client, flag.Args()[1:])
	case "version":
		handleVersion(client, flag.Args()[1:])
	case "telemetry":
		handleTelemetry(client, flag.Args()[1:])
	default:
		fmt.Printf("Unknown subcommand: %s\n", subcommand)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Usage: cloud-cli [flags] <subcommand> [args]")
	fmt.Println("Flags:")
	flag.PrintDefaults()
	fmt.Println("\nSubcommands:")
	fmt.Println("  gchart    <get|search|create> [args]")
	fmt.Println("  version   <get|latest|create> [args]")
	fmt.Println("  telemetry <list|upsert> [args]")
}

func handleGChart(client *CloudClient, args []string) {
	if len(args) < 1 {
		fmt.Println("Usage: gchart <get|search|create> ...")
		return
	}

	cmd := args[0]
	switch cmd {
	case "get":
		if len(args) < 2 {
			fmt.Println("Usage: gchart get <id>")
			return
		}
		id := args[1]
		chart, err := client.GetGChart(id)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}
		printJSON(chart)

	case "search":
		// Usage: gchart search [flags]
		// Flags: -date <RFC3339> -curated

		searchCmd := flag.NewFlagSet("search", flag.ExitOnError)
		dateFromPtr := searchCmd.String("date", "", "Date from (RFC3339)")
		curatedPtr := searchCmd.Bool("curated", false, "Filter for curated charts only")

		searchCmd.Parse(args[1:])

		headers, err := client.SearchGChartHeaders(*dateFromPtr)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}

		if *curatedPtr {
			var filtered GChartAPIv1HeaderOnlyList
			for _, h := range headers {
				if h.Header.Curated {
					filtered = append(filtered, h)
				}
			}
			headers = filtered
		}

		printJSON(headers)

	case "create":
		// Usage: gchart create <json_file>
		if len(args) < 2 {
			fmt.Println("Usage: gchart create <json_file>")
			return
		}
		data, err := os.ReadFile(args[1])
		if err != nil {
			fmt.Printf("Error reading file: %v\n", err)
			return
		}
		var chart GChartPostAPIv1
		if err := json.Unmarshal(data, &chart); err != nil {
			fmt.Printf("Error parsing JSON: %v\n", err)
			return
		}
		id, err := client.CreateGChart(chart)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}
		fmt.Printf("Created GChart with ID: %s\n", id)

	default:
		fmt.Printf("Unknown gchart command: %s\n", cmd)
	}
}

func handleVersion(client *CloudClient, args []string) {
	if len(args) < 1 {
		fmt.Println("Usage: version <get|latest|create> ...")
		return
	}

	cmd := args[0]
	switch cmd {
	case "get":
		// Usage: version get [minVersion]
		minVer := ""
		if len(args) > 1 {
			minVer = args[1]
		}
		vers, err := client.GetVersions(minVer)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}
		printJSON(vers)

	case "latest":
		ver, err := client.GetLatestVersion()
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}
		printJSON(ver)

	case "create":
		// Usage: version create <json_file>
		if len(args) < 2 {
			fmt.Println("Usage: version create <json_file>")
			return
		}
		data, err := os.ReadFile(args[1])
		if err != nil {
			fmt.Printf("Error reading file: %v\n", err)
			return
		}
		var ver VersionEntityPostAPIv1
		if err := json.Unmarshal(data, &ver); err != nil {
			fmt.Printf("Error parsing JSON: %v\n", err)
			return
		}
		id, err := client.CreateVersion(ver)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}
		fmt.Printf("Created Version with ID: %s\n", id)

	default:
		fmt.Printf("Unknown version command: %s\n", cmd)
	}
}

func handleTelemetry(client *CloudClient, args []string) {
	if len(args) < 1 {
		fmt.Println("Usage: telemetry <list|upsert> ...")
		return
	}

	cmd := args[0]
	switch cmd {
	case "list":
		// Usage: telemetry list [createdAfter] [updatedAfter] [os] [version]
		// This argument parsing is simple/brittle for simplicity
		// Let's just say we support positional args or just one filter
		// Or assume the user passes specific flags? But we are already deep in subcommands.
		// Let's implement simple positional: list <createdAfter> <updatedAfter> <os> <version>
		// Pass "_" to skip.

		createdAfter, updatedAfter, osVal, verVal := "", "", "", ""
		if len(args) > 1 && args[1] != "_" {
			createdAfter = args[1]
		}
		if len(args) > 2 && args[2] != "_" {
			updatedAfter = args[2]
		}
		if len(args) > 3 && args[3] != "_" {
			osVal = args[3]
		}
		if len(args) > 4 && args[4] != "_" {
			verVal = args[4]
		}

		list, err := client.GetTelemetry(createdAfter, updatedAfter, osVal, verVal)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}
		printJSON(list)

	case "upsert":
		// Usage: telemetry upsert <key> [os] [version] [increment]
		if len(args) < 2 {
			fmt.Println("Usage: telemetry upsert <key> [os] [version] [increment=1]")
			return
		}
		key := args[1]
		osVal := "Linux"
		if len(args) > 2 {
			osVal = args[2]
		}
		verVal := "3.6"
		if len(args) > 3 {
			verVal = args[3]
		}
		inc := int64(1)
		// skip parsing increment complexity for now or...

		tel := TelemetryEntityPostAPIv1{
			UserKey:    key,
			OS:         osVal,
			GCVersion:  verVal,
			Increment:  inc,
			LastChange: time.Now().Format(time.RFC3339),
		}

		res, err := client.UpsertTelemetry(tel)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}
		printJSON(res)

	default:
		fmt.Printf("Unknown telemetry command: %s\n", cmd)
	}
}

func printJSON(v interface{}) {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		fmt.Printf("Error marshaling JSON: %v\n", err)
		return
	}
	fmt.Println(string(b))
}
