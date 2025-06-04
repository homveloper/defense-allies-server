package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

func main() {
	fmt.Println("🏰 Defense Allies - Guild System Examples")
	fmt.Println("=========================================")
	fmt.Println()
	fmt.Println("Available examples:")
	fmt.Println("1. Basic Guild Management - Create, manage members, roles")
	fmt.Println("2. Guild Mining System - Mining operations, resource management")
	fmt.Println("3. Guild Transport System - Transport recruitment, cargo delivery")
	fmt.Println("4. Exit")
	fmt.Println()

	for {
		fmt.Print("Select an example (1-4): ")
		var choice string
		fmt.Scanln(&choice)

		switch choice {
		case "1":
			runExample("basic-guild-management")
		case "2":
			runExample("guild-mining")
		case "3":
			runExample("guild-transport")
		case "4":
			fmt.Println("👋 Goodbye!")
			return
		default:
			fmt.Println("❌ Invalid choice. Please select 1, 2, 3, or 4.")
		}

		fmt.Println()
		fmt.Println("Press Enter to continue...")
		fmt.Scanln()
		fmt.Println()
	}
}

func runExample(exampleName string) {
	fmt.Printf("\n🚀 Running %s example...\n", exampleName)
	fmt.Println("=" + fmt.Sprintf("%*s", len(exampleName)+20, "="))

	// Get the current working directory
	wd, err := os.Getwd()
	if err != nil {
		fmt.Printf("❌ Error getting working directory: %v\n", err)
		return
	}

	// Build the path to the example
	examplePath := filepath.Join(wd, "cmd", exampleName)

	// Check if the example directory exists
	if _, err := os.Stat(examplePath); os.IsNotExist(err) {
		fmt.Printf("❌ Example directory not found: %s\n", examplePath)
		return
	}

	// Run the example
	cmd := exec.Command("go", "run", "main.go")
	cmd.Dir = examplePath
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		fmt.Printf("❌ Error running example: %v\n", err)
		return
	}

	fmt.Printf("\n✅ %s example completed!\n", exampleName)
}
