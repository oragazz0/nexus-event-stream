package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"github.com/oragazz0/nexus-event-stream/data-plane/internal/client"
	"github.com/oragazz0/nexus-event-stream/data-plane/internal/domain"
)

const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorBold   = "\033[1m"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	apiURL := envOrDefault("API_URL", "http://localhost:8081")
	dataPlane := client.New(apiURL)

	switch os.Args[1] {
	case "list":
		runList(dataPlane)
	case "get":
		runGet(dataPlane)
	case "health":
		runHealth(dataPlane)
	default:
		printUsage()
		os.Exit(1)
	}
}

func runList(dataPlane client.DataPlane) {
	flags := flag.NewFlagSet("list", flag.ExitOnError)
	priority := flags.String("priority", "", "Filter by priority (Low, Medium, High)")
	flags.Parse(os.Args[2:])

	signals, err := dataPlane.ListSignals(*priority)
	if err != nil {
		exitWithError(err)
	}

	if len(signals) == 0 {
		fmt.Println("No signals found.")
		return
	}
	printSignalTable(signals)
}

func runGet(dataPlane client.DataPlane) {
	flags := flag.NewFlagSet("get", flag.ExitOnError)
	flags.Parse(os.Args[2:])

	args := flags.Args()
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "Error: signal ID is required")
		fmt.Fprintln(os.Stderr, "Usage: nexus-cli get <signal-id>")
		os.Exit(1)
	}

	signal, err := dataPlane.GetSignal(args[0])
	if errors.Is(err, client.ErrNotFound) {
		fmt.Fprintf(os.Stderr, "Signal %q not found.\n", args[0])
		os.Exit(1)
	}
	if err != nil {
		exitWithError(err)
	}
	printSignalDetail(signal)
}

func runHealth(dataPlane client.DataPlane) {
	err := dataPlane.Health()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s✗ Data Plane is unreachable: %v%s\n", colorRed, err, colorReset)
		os.Exit(1)
	}
	fmt.Printf("%s✓ Data Plane is healthy%s\n", colorGreen, colorReset)
}

func printSignalTable(signals []domain.Signal) {
	writer := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintf(writer, "%sID\tPRIORITY\tAUTHOR\tTITLE\tCREATED%s\n", colorBold, colorReset)

	for _, signal := range signals {
		color := priorityColor(signal.Priority)
		fmt.Fprintf(writer, "%s\t%s%s%s\t%s\t%s\t%s\n",
			signal.ID,
			color, signal.Priority, colorReset,
			signal.Author,
			truncate(signal.Title, 40),
			formatTime(signal.CreatedAt),
		)
	}
	writer.Flush()
}

func printSignalDetail(signal domain.Signal) {
	color := priorityColor(signal.Priority)
	fmt.Printf("%sID:%s        %s\n", colorBold, colorReset, signal.ID)
	fmt.Printf("%sTitle:%s     %s\n", colorBold, colorReset, signal.Title)
	fmt.Printf("%sContent:%s   %s\n", colorBold, colorReset, signal.Content)
	fmt.Printf("%sPriority:%s  %s%s%s\n", colorBold, colorReset, color, signal.Priority, colorReset)
	fmt.Printf("%sAuthor:%s    %s\n", colorBold, colorReset, signal.Author)
	fmt.Printf("%sCreated:%s   %s\n", colorBold, colorReset, signal.CreatedAt)
	fmt.Printf("%sUpdated:%s   %s\n", colorBold, colorReset, signal.UpdatedAt)
}

func printUsage() {
	fmt.Printf("%snexus-cli%s — Nexus Data Plane client\n\n", colorBold, colorReset)
	fmt.Println("Usage: nexus-cli <command> [flags]")
	fmt.Println()
	fmt.Printf("%sCommands:%s\n", colorBold, colorReset)
	fmt.Println("  list      List signals")
	fmt.Println("  get       Get a signal by ID")
	fmt.Println("  health    Check data-plane health")
	fmt.Println()
	fmt.Printf("%sExamples:%s\n", colorBold, colorReset)
	fmt.Println("  nexus-cli list")
	fmt.Println("  nexus-cli list -priority High")
	fmt.Println("  nexus-cli get 550e8400-e29b-41d4-a716-446655440000")
	fmt.Println("  nexus-cli health")
	fmt.Println()
	fmt.Printf("%sEnvironment:%s\n", colorBold, colorReset)
	fmt.Println("  API_URL   Data plane base URL (default: http://localhost:8081)")
}

func priorityColor(priority string) string {
	colors := map[string]string{
		"High":   colorRed,
		"Medium": colorYellow,
		"Low":    colorGreen,
	}
	color, ok := colors[priority]
	if !ok {
		return colorReset
	}
	return color
}

func formatTime(isoTime string) string {
	parsed, err := time.Parse(time.RFC3339, isoTime)
	if err != nil {
		return isoTime
	}
	return parsed.Format("2006-01-02 15:04")
}

func truncate(text string, maxLength int) string {
	if len(text) <= maxLength {
		return text
	}
	return text[:maxLength-1] + "…"
}

func exitWithError(err error) {
	fmt.Fprintf(os.Stderr, "%sError:%s %v\n", colorRed, colorReset, err)
	os.Exit(1)
}

func envOrDefault(key, fallback string) string {
	value := os.Getenv(key)
	if value != "" {
		return value
	}
	return fallback
}
