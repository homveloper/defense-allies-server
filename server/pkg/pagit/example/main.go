package main

import (
	"fmt"
	"strings"
)

func main() {
	fmt.Println("Running Pagination Examples")
	fmt.Println("==========================")

	// Run offset-based pagination example
	runOffsetExample()

	fmt.Println("\n" + strings.Repeat("=", 50) + "\n")

	// Run cursor-based pagination example
	runCursorExample()
	
	fmt.Println("\n" + strings.Repeat("=", 50) + "\n")
	
	// Run sorting examples
	runSortExample()
}
