// Example: Loading plugins with the SDK
//
// This example demonstrates loading and using plugins with the Claude SDK.
// Plugins extend Claude's capabilities by adding custom commands and tools.
//
// Plugin Directory Structure:
// A plugin is a directory containing configuration and command definitions:
//
//	my-plugin/
//	├── .claude-plugin/
//	│   └── plugin.json      # Plugin metadata: {"name": "...", "description": "..."}
//	└── commands/
//	    └── greet.md         # Custom slash command definition
//
// The plugin.json file defines the plugin:
//
//	{
//	    "name": "my-plugin",
//	    "description": "A sample plugin for demonstration"
//	}
//
// Each command file (e.g., greet.md) defines a slash command:
//
//	---
//	name: greet
//	description: Greets the user with a friendly message
//	---
//
//	When the user uses /greet, respond with a friendly greeting
//	including their name if provided.
//
// Run with: go run main.go
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	claude "github.com/afsharalex/claude-agent-sdk-go"
)

func main() {
	// Create a context that cancels on interrupt
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle interrupt signals
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigCh
		fmt.Println("\nInterrupted, shutting down...")
		cancel()
	}()

	// Example 1: Using WithLocalPlugin for a local plugin path
	fmt.Println("=== Example 1: Local Plugin ===")
	localPluginExample(ctx)

	// Example 2: Using WithPlugin for more control
	fmt.Println("\n=== Example 2: Plugin with SdkPluginConfig ===")
	pluginConfigExample(ctx)
}

func localPluginExample(ctx context.Context) {
	// WithLocalPlugin is a convenience function for loading plugins from a local path.
	// The path should point to a directory containing .claude-plugin/plugin.json.
	//
	// Note: This example uses a placeholder path. Replace with an actual plugin path
	// to see the plugin in action.
	pluginPath := "./my-plugin" // Replace with actual plugin path

	client := claude.NewClient(
		// Load a local plugin by path
		claude.WithLocalPlugin(pluginPath),
	)

	fmt.Printf("Configured plugin from: %s\n", pluginPath)
	fmt.Println("Note: This example uses a placeholder path.")
	fmt.Println("To test with a real plugin, create a plugin directory with the structure shown above.")

	// Connect to Claude Code
	fmt.Println("\nConnecting to Claude Code...")
	if err := client.Connect(ctx); err != nil {
		// This may fail if the plugin path doesn't exist or is invalid
		log.Printf("Connection note: %v", err)
		fmt.Println("(This is expected if the plugin path doesn't exist)")
		return
	}
	defer func() { _ = client.Close() }()

	// Query to list available commands (plugins add slash commands)
	if err := client.Query(ctx, "What commands are available?"); err != nil {
		log.Printf("Query failed: %v", err)
		return
	}

	processResponse(ctx, client)
}

func pluginConfigExample(ctx context.Context) {
	// WithPlugin provides more control over plugin configuration.
	// Currently only "local" type is supported.
	//
	// SdkPluginConfig fields:
	// - Type: Plugin type ("local")
	// - Path: Path to the plugin directory

	pluginConfigs := []claude.SdkPluginConfig{
		{
			Type: claude.SdkPluginTypeLocal,
			Path: "./plugin-one",
		},
		{
			Type: claude.SdkPluginTypeLocal,
			Path: "./plugin-two",
		},
	}

	// You can add multiple plugins using WithPlugins
	client := claude.NewClient(
		claude.WithPlugins(pluginConfigs),
	)

	fmt.Println("Configured multiple plugins:")
	for _, cfg := range pluginConfigs {
		fmt.Printf("  - Type: %s, Path: %s\n", cfg.Type, cfg.Path)
	}

	// Or add plugins one at a time using WithPlugin
	client2 := claude.NewClient(
		claude.WithPlugin(claude.SdkPluginConfig{
			Type: claude.SdkPluginTypeLocal,
			Path: "./my-custom-plugin",
		}),
	)
	_ = client2 // Shown for documentation

	fmt.Println("\nNote: Plugin paths in this example are placeholders.")
	fmt.Println("To test with real plugins, create plugin directories with the structure shown in the header.")

	// Demonstrate checking SystemMessage for plugin info
	fmt.Println("\nWhen a client connects with plugins, the SystemMessage.Data may contain plugin info.")
	fmt.Println("You can inspect it to verify plugins loaded correctly.")
	fmt.Println("See the processResponse function in this example for a demonstration.")

	// Connect to show the pattern (may fail with placeholder paths)
	fmt.Println("\nAttempting connection with placeholder plugins...")
	if err := client.Connect(ctx); err != nil {
		log.Printf("Connection note: %v", err)
		fmt.Println("(This is expected if plugin paths don't exist)")
		return
	}
	defer func() { _ = client.Close() }()

	// Simple query
	if err := client.Query(ctx, "Hello!"); err != nil {
		log.Printf("Query failed: %v", err)
		return
	}

	processResponse(ctx, client)
}

func processResponse(ctx context.Context, client *claude.Client) {
	for msg := range client.ReceiveResponse(ctx) {
		switch m := msg.(type) {
		case *claude.SystemMessage:
			// Check for plugin-related information
			if m.Subtype == "init" {
				fmt.Println("\n[System Init Message]")
				if plugins, ok := m.Data["plugins"]; ok {
					fmt.Printf("Loaded plugins: %v\n", plugins)
				}
			}

		case *claude.AssistantMessage:
			for _, block := range m.Content {
				if text, ok := block.(claude.TextBlock); ok {
					fmt.Printf("Claude: %s\n", text.Text)
				}
			}

		case *claude.ResultMessage:
			if m.TotalCostUSD != nil {
				fmt.Printf("\nCost: $%.4f\n", *m.TotalCostUSD)
			}
		}
	}
}
