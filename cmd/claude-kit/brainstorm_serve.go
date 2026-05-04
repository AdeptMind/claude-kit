package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"github.com/spf13/cobra"

	"github.com/AdeptMind/infra-tool/claude-cli/internal/visualbs"
)

var brainstormPort int

var brainstormServeCmd = &cobra.Command{
	Use:   "brainstorm-serve",
	Short: "Start the visual brainstorming server",
	Long: `Launch a local HTTP server for visual brainstorming.

The server displays interactive mockups, wireframes, and option cards in the browser.
Agents push HTML content via POST /push, and user choices come back via GET /events.

The server auto-shuts down after 30 minutes of inactivity.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runBrainstormServe()
	},
}

func init() {
	brainstormServeCmd.Flags().IntVar(&brainstormPort, "port", 0, "Port to listen on (default: auto)")
	rootCmd.AddCommand(brainstormServeCmd)
}

func runBrainstormServe() error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	srv := visualbs.NewServer(brainstormPort)
	url, _, err := srv.Start(ctx)
	if err != nil {
		return err
	}

	fmt.Println(banner())
	fmt.Println(subtitleStyle.Render("  Visual Brainstorming Server"))
	fmt.Println()
	fmt.Println(fmt.Sprintf("  %s %s", arrow, accentStyle.Render(url)))
	fmt.Println()
	fmt.Println(dimStyle.Render("  Agents push content:  curl -X POST " + url + "/push -d '{\"type\":\"html\",\"content\":\"<div>...</div>\"}'"))
	fmt.Println(dimStyle.Render("  Agents read choices:  curl " + url + "/events"))
	fmt.Println(dimStyle.Render("  Auto-shutdown after 30 min idle. Press Ctrl+C to stop."))
	fmt.Println()

	<-ctx.Done()
	fmt.Println("\n" + dimStyle.Render("  Server stopped."))
	return nil
}
