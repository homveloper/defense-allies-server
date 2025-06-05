package main

import (
	"bufio"
	"context"
	"defense-allies-server/pkg/cqrs/cqrsx/examples/01-basic-event-sourcing/demo"
	"defense-allies-server/pkg/cqrs/cqrsx/examples/01-basic-event-sourcing/domain"
	"defense-allies-server/pkg/cqrs/cqrsx/examples/01-basic-event-sourcing/infrastructure"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
)

func main() {
	fmt.Println("🚀 Basic Event Sourcing Example")
	fmt.Println("================================")

	// 인프라스트럭처 초기화
	config := infrastructure.GetDefaultConfig()
	infra, err := infrastructure.NewInfrastructure(config)
	if err != nil {
		log.Fatalf("Failed to initialize infrastructure: %v", err)
	}

	ctx := context.Background()
	defer func() {
		if err := infra.Close(ctx); err != nil {
			log.Printf("Error closing infrastructure: %v", err)
		}
	}()

	// 스키마 초기화
	err = infra.Initialize(ctx)
	if err != nil {
		log.Fatalf("Failed to initialize schema: %v", err)
	}

	// 헬스체크
	err = infra.HealthCheck(ctx)
	if err != nil {
		log.Fatalf("Infrastructure health check failed: %v", err)
	}

	fmt.Printf("✅ Connected to MongoDB: %s\n", config.MongoDB.Database)
	fmt.Printf("📊 Application: %s v%s (%s)\n",
		config.App.Name, config.App.Version, config.App.Environment)

	// 데모 시나리오 실행기 생성
	demoScenarios := demo.NewDemoScenarios(infra)

	// 대화형 모드 시작
	fmt.Println("\n" + strings.Repeat("=", 50))
	fmt.Println("Interactive Demo Mode")
	fmt.Println(strings.Repeat("=", 50))

	showHelp()

	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("\n> ")
		if !scanner.Scan() {
			break
		}

		input := strings.TrimSpace(scanner.Text())
		if input == "" {
			continue
		}

		parts := strings.Fields(input)
		command := strings.ToLower(parts[0])

		err := handleCommand(ctx, command, parts[1:], infra, demoScenarios)
		if err != nil {
			if err.Error() == "exit" {
				break
			}
			fmt.Printf("❌ Error: %v\n", err)
		}
	}

	fmt.Println("\n👋 Goodbye!")
}

func handleCommand(ctx context.Context, command string, args []string, infra *infrastructure.Infrastructure, demoScenarios *demo.DemoScenarios) error {
	switch command {
	case "help", "h":
		showHelp()
		return nil

	case "create":
		return handleCreateUser(ctx, args, infra)

	case "update":
		return handleUpdateUser(ctx, args, infra)

	case "delete":
		return handleDeleteUser(ctx, args, infra)

	case "activate":
		return handleActivateUser(ctx, args, infra)

	case "deactivate":
		return handleDeactivateUser(ctx, args, infra)

	case "get":
		return handleGetUser(ctx, args, infra)

	case "history":
		return handleGetHistory(ctx, args, infra)

	case "list":
		return handleListUsers(ctx, infra)

	case "stats":
		return handleGetStats(ctx, infra)

	case "clear":
		return handleClearData(ctx, infra)

	case "demo":
		return handleRunDemo(ctx, args, demoScenarios)

	case "metrics":
		return handleGetMetrics(infra)

	case "exit", "quit", "q":
		return fmt.Errorf("exit")

	default:
		return fmt.Errorf("unknown command: %s. Type 'help' for available commands", command)
	}
}

func handleCreateUser(ctx context.Context, args []string, infra *infrastructure.Infrastructure) error {
	if len(args) < 2 {
		return fmt.Errorf("usage: create <name> <email>")
	}

	name := args[0]
	email := args[1]
	userID := uuid.New().String()

	user := domain.NewUserWithID(userID)
	err := user.CreateUser(userID, name, email)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	err = infra.UserRepo.Save(ctx, user)
	if err != nil {
		return fmt.Errorf("failed to save user: %w", err)
	}

	fmt.Printf("✅ Created user: %s\n", user.String())
	return nil
}

func handleUpdateUser(ctx context.Context, args []string, infra *infrastructure.Infrastructure) error {
	if len(args) < 3 {
		return fmt.Errorf("usage: update <id> <name> <email>")
	}

	userID := args[0]
	newName := args[1]
	newEmail := args[2]

	user, err := infra.UserRepo.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to load user: %w", err)
	}

	err = user.UpdateUser(newName, newEmail)
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	err = infra.UserRepo.Save(ctx, user)
	if err != nil {
		return fmt.Errorf("failed to save user: %w", err)
	}

	fmt.Printf("✅ Updated user: %s\n", user.String())
	return nil
}

func handleDeleteUser(ctx context.Context, args []string, infra *infrastructure.Infrastructure) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: delete <id> [reason]")
	}

	userID := args[0]
	reason := "User requested deletion"
	if len(args) > 1 {
		reason = strings.Join(args[1:], " ")
	}

	user, err := infra.UserRepo.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to load user: %w", err)
	}

	err = user.DeleteUser(reason)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	err = infra.UserRepo.Save(ctx, user)
	if err != nil {
		return fmt.Errorf("failed to save user: %w", err)
	}

	fmt.Printf("✅ Deleted user: %s\n", user.String())
	return nil
}

func handleActivateUser(ctx context.Context, args []string, infra *infrastructure.Infrastructure) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: activate <id> [activated_by]")
	}

	userID := args[0]
	activatedBy := "admin"
	if len(args) > 1 {
		activatedBy = args[1]
	}

	user, err := infra.UserRepo.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to load user: %w", err)
	}

	err = user.ActivateUser(activatedBy)
	if err != nil {
		return fmt.Errorf("failed to activate user: %w", err)
	}

	err = infra.UserRepo.Save(ctx, user)
	if err != nil {
		return fmt.Errorf("failed to save user: %w", err)
	}

	fmt.Printf("✅ Activated user: %s\n", user.String())
	return nil
}

func handleDeactivateUser(ctx context.Context, args []string, infra *infrastructure.Infrastructure) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: deactivate <id> [deactivated_by] [reason]")
	}

	userID := args[0]
	deactivatedBy := "admin"
	reason := "Administrative action"

	if len(args) > 1 {
		deactivatedBy = args[1]
	}
	if len(args) > 2 {
		reason = strings.Join(args[2:], " ")
	}

	user, err := infra.UserRepo.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to load user: %w", err)
	}

	err = user.DeactivateUser(deactivatedBy, reason)
	if err != nil {
		return fmt.Errorf("failed to deactivate user: %w", err)
	}

	err = infra.UserRepo.Save(ctx, user)
	if err != nil {
		return fmt.Errorf("failed to save user: %w", err)
	}

	fmt.Printf("✅ Deactivated user: %s\n", user.String())
	return nil
}

func handleGetUser(ctx context.Context, args []string, infra *infrastructure.Infrastructure) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: get <id>")
	}

	userID := args[0]
	user, err := infra.UserRepo.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to load user: %w", err)
	}

	fmt.Printf("👤 User Details:\n")
	fmt.Printf("   ID: %s\n", user.ID())
	fmt.Printf("   Name: %s\n", user.Name())
	fmt.Printf("   Email: %s\n", user.Email())
	fmt.Printf("   Active: %t\n", user.IsActive())
	fmt.Printf("   Deleted: %t\n", user.IsDeleted())
	fmt.Printf("   Version: %d\n", user.Version())
	fmt.Printf("   Created: %s\n", user.CreatedAt().Format(time.RFC3339))
	fmt.Printf("   Updated: %s\n", user.UpdatedAt().Format(time.RFC3339))
	if user.DeletedAt() != nil {
		fmt.Printf("   Deleted: %s\n", user.DeletedAt().Format(time.RFC3339))
	}

	return nil
}

func handleGetHistory(ctx context.Context, args []string, infra *infrastructure.Infrastructure) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: history <id>")
	}

	userID := args[0]
	events, err := infra.UserRepo.GetEventHistory(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get event history: %w", err)
	}

	fmt.Printf("📜 Event History for User %s (%d events):\n", userID, len(events))
	for i, event := range events {
		fmt.Printf("   %d. %s (v%d) at %s\n",
			i+1,
			event.EventType(),
			event.Version(),
			event.Timestamp().Format("2006-01-02 15:04:05"))
	}

	return nil
}

func handleListUsers(ctx context.Context, infra *infrastructure.Infrastructure) error {
	users, err := infra.UserRepo.ListAllUsers(ctx)
	if err != nil {
		return fmt.Errorf("failed to list users: %w", err)
	}

	fmt.Printf("👥 All Users (%d total):\n", len(users))
	for i, user := range users {
		status := "inactive"
		if user.IsDeleted() {
			status = "deleted"
		} else if user.IsActive() {
			status = "active"
		}

		fmt.Printf("   %d. %s (%s) - %s [%s]\n",
			i+1,
			user.Name(),
			user.Email(),
			status,
			user.ID()[:8]+"...")
	}

	return nil
}

func handleGetStats(ctx context.Context, infra *infrastructure.Infrastructure) error {
	stats, err := infra.UserRepo.GetUserStats(ctx)
	if err != nil {
		return fmt.Errorf("failed to get stats: %w", err)
	}

	fmt.Printf("📊 User Statistics:\n")
	for key, value := range stats {
		fmt.Printf("   %s: %v\n", strings.Title(strings.ReplaceAll(key, "_", " ")), value)
	}

	return nil
}

func handleClearData(ctx context.Context, infra *infrastructure.Infrastructure) error {
	fmt.Print("⚠️  Are you sure you want to clear all data? (yes/no): ")
	scanner := bufio.NewScanner(os.Stdin)
	if scanner.Scan() {
		response := strings.ToLower(strings.TrimSpace(scanner.Text()))
		if response == "yes" || response == "y" {
			err := infra.ClearData(ctx)
			if err != nil {
				return fmt.Errorf("failed to clear data: %w", err)
			}
			fmt.Println("✅ All data cleared successfully")
		} else {
			fmt.Println("❌ Operation cancelled")
		}
	}
	return nil
}

func handleRunDemo(ctx context.Context, args []string, demoScenarios *demo.DemoScenarios) error {
	if len(args) == 0 {
		return demoScenarios.RunAllScenarios(ctx)
	}

	scenario := strings.ToLower(args[0])
	switch scenario {
	case "basic", "crud":
		return demoScenarios.RunBasicCRUDScenario(ctx)
	case "restoration", "restore":
		return demoScenarios.RunEventRestorationScenario(ctx)
	case "concurrency", "concurrent":
		return demoScenarios.RunConcurrencyScenario(ctx)
	case "performance", "perf":
		return demoScenarios.RunPerformanceScenario(ctx)
	case "all":
		return demoScenarios.RunAllScenarios(ctx)
	default:
		return fmt.Errorf("unknown demo scenario: %s. Available: basic, restoration, concurrency, performance, all", scenario)
	}
}

func handleGetMetrics(infra *infrastructure.Infrastructure) error {
	metrics := infra.GetMetrics()
	fmt.Printf("📈 Infrastructure Metrics:\n")

	for category, data := range metrics {
		fmt.Printf("   %s:\n", strings.Title(category))
		if dataMap, ok := data.(map[string]interface{}); ok {
			for key, value := range dataMap {
				fmt.Printf("     %s: %v\n", key, value)
			}
		} else {
			fmt.Printf("     %v\n", data)
		}
	}

	return nil
}

func showHelp() {
	fmt.Println(`
📖 Available Commands:

User Management:
  create <name> <email>              - Create a new user
  update <id> <name> <email>         - Update user information
  delete <id> [reason]               - Delete a user
  activate <id> [activated_by]       - Activate a user
  deactivate <id> [by] [reason]      - Deactivate a user
  get <id>                           - Get user details
  history <id>                       - Get user event history
  list                               - List all users
  stats                              - Show user statistics

Demo Scenarios:
  demo [scenario]                    - Run demo scenarios
    - basic/crud                     - Basic CRUD operations
    - restoration/restore            - Event restoration demo
    - concurrency/concurrent         - Concurrency handling demo
    - performance/perf               - Performance test demo
    - all                            - Run all scenarios

System:
  clear                              - Clear all data (with confirmation)
  metrics                            - Show infrastructure metrics
  help/h                             - Show this help
  exit/quit/q                        - Exit the program

💡 Tips:
  - User IDs are auto-generated UUIDs
  - All operations are logged to the event store
  - Use 'demo all' to see comprehensive examples
  - Use 'history <id>' to see how events build up state`)
}
