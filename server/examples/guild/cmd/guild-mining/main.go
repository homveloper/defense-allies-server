package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"

	"defense-allies-server/examples/guild/application/commands"
	"defense-allies-server/examples/guild/application/handlers"
	"defense-allies-server/examples/guild/domain"
	"defense-allies-server/examples/guild/infrastructure/projections"
	"defense-allies-server/examples/guild/infrastructure/queries"
	"defense-allies-server/examples/guild/infrastructure/repositories"
	"defense-allies-server/pkg/cqrs"
)

func main() {
	fmt.Println("⛏️ Defense Allies - Guild Mining System Example")
	fmt.Println("===============================================")

	// Initialize CQRS infrastructure
	ctx := context.Background()

	// Create in-memory read store for projections
	readStore := cqrs.NewInMemoryReadStore()

	// Create command dispatcher
	commandDispatcher := cqrs.NewInMemoryCommandDispatcher()

	// Create query dispatcher
	queryDispatcher := cqrs.NewInMemoryQueryDispatcher()

	// Create projection manager
	projectionManager := cqrs.NewInMemoryProjectionManager()

	// Create projections
	guildViewProjection := projections.NewGuildViewProjection(readStore)
	memberViewProjection := projections.NewMemberViewProjection(readStore)
	allProjections := []cqrs.Projection{guildViewProjection, memberViewProjection}

	// Create in-memory repository for this example (with projections)
	repository := repositories.NewInMemoryGuildRepository(allProjections)

	// Create and register command handler
	guildHandler := handlers.NewGuildCommandHandler(repository)

	// Create and register query handler
	guildQueryHandler := queries.NewGuildQueryHandler(readStore)

	// Register command handlers
	commandTypes := []string{
		commands.CreateGuildCommandType,
		commands.UpdateGuildInfoCommandType,
		commands.UpdateGuildSettingsCommandType,
		commands.InviteMemberCommandType,
		commands.AcceptInvitationCommandType,
		commands.KickMemberCommandType,
		commands.PromoteMemberCommandType,
	}
	for _, commandType := range commandTypes {
		if err := commandDispatcher.RegisterHandler(commandType, guildHandler); err != nil {
			log.Fatalf("Failed to register %s handler: %v", commandType, err)
		}
	}

	// Create event bus for projections
	eventBus := cqrs.NewInMemoryEventBus()
	if err := eventBus.Start(ctx); err != nil {
		log.Fatalf("Failed to start event bus: %v", err)
	}
	defer eventBus.Stop(ctx)

	// Register projections with projection manager
	if err := projectionManager.RegisterProjection(guildViewProjection); err != nil {
		log.Fatalf("Failed to register guild view projection: %v", err)
	}
	if err := projectionManager.RegisterProjection(memberViewProjection); err != nil {
		log.Fatalf("Failed to register member view projection: %v", err)
	}

	// Start projection manager
	if err := projectionManager.Start(ctx); err != nil {
		log.Fatalf("Failed to start projection manager: %v", err)
	}
	defer projectionManager.Stop(ctx)

	// Register query handlers
	queryTypes := []string{
		queries.GetGuildQueryType,
		queries.GetGuildMembersQueryType,
		queries.SearchGuildsQueryType,
	}
	for _, queryType := range queryTypes {
		if err := queryDispatcher.RegisterHandler(queryType, guildQueryHandler); err != nil {
			log.Fatalf("Failed to register %s handler: %v", queryType, err)
		}
	}

	fmt.Println("\n✅ CQRS Infrastructure initialized successfully")

	// Run the guild mining example
	if err := runGuildMiningExample(ctx, commandDispatcher, queryDispatcher, repository); err != nil {
		log.Fatalf("Example failed: %v", err)
	}

	fmt.Println("\n🎉 Guild mining example completed successfully!")
}

func runGuildMiningExample(ctx context.Context, dispatcher cqrs.CommandDispatcher, queryDispatcher cqrs.QueryDispatcher, repository cqrs.EventSourcedRepository) error {
	// Generate IDs
	guildID := uuid.New().String()
	founderID := "founder123"
	founderUsername := "MiningMaster"
	miner1ID := "miner001"
	miner1Username := "IronDigger"
	miner2ID := "miner002"
	miner2Username := "GoldSeeker"

	fmt.Printf("\n⛏️ Creating mining guild with ID: %s\n", guildID)

	// Step 1: Create guild
	fmt.Println("\n1️⃣ Creating mining guild...")
	createCmd := commands.NewCreateGuildCommand(guildID, "Mining Consortium", "A guild dedicated to mining operations", founderID, founderUsername)
	result, err := dispatcher.Dispatch(ctx, createCmd)
	if err != nil {
		return fmt.Errorf("failed to create guild: %w", err)
	}
	fmt.Printf("   ✅ Mining guild created: %s\n", getMessageFromResult(result, "Guild created successfully"))

	// Step 2: Update guild settings to allow more members
	fmt.Println("\n2️⃣ Updating guild settings...")
	updateSettingsCmd := commands.NewUpdateGuildSettingsCommand(guildID, 50, 1, true, false, founderID)
	result, err = dispatcher.Dispatch(ctx, updateSettingsCmd)
	if err != nil {
		return fmt.Errorf("failed to update guild settings: %w", err)
	}
	fmt.Printf("   ✅ Guild settings updated: %s\n", getMessageFromResult(result, "Guild settings updated successfully"))

	// Step 3: Add miners to the guild
	fmt.Println("\n3️⃣ Recruiting miners...")

	// Invite first miner
	inviteCmd1 := commands.NewInviteMemberCommand(guildID, miner1ID, miner1Username, founderID)
	result, err = dispatcher.Dispatch(ctx, inviteCmd1)
	if err != nil {
		return fmt.Errorf("failed to invite miner: %w", err)
	}
	fmt.Printf("   ✅ Miner invited: %s\n", miner1Username)

	// Accept first invitation
	acceptCmd1 := commands.NewAcceptInvitationCommand(guildID, miner1ID)
	result, err = dispatcher.Dispatch(ctx, acceptCmd1)
	if err != nil {
		return fmt.Errorf("failed to accept invitation: %w", err)
	}
	fmt.Printf("   ✅ %s joined the guild\n", miner1Username)

	// Invite second miner
	inviteCmd2 := commands.NewInviteMemberCommand(guildID, miner2ID, miner2Username, founderID)
	result, err = dispatcher.Dispatch(ctx, inviteCmd2)
	if err != nil {
		return fmt.Errorf("failed to invite second miner: %w", err)
	}
	fmt.Printf("   ✅ Miner invited: %s\n", miner2Username)

	// Accept second invitation
	acceptCmd2 := commands.NewAcceptInvitationCommand(guildID, miner2ID)
	result, err = dispatcher.Dispatch(ctx, acceptCmd2)
	if err != nil {
		return fmt.Errorf("failed to accept second invitation: %w", err)
	}
	fmt.Printf("   ✅ %s joined the guild\n", miner2Username)

	// Step 4: Load guild and set up mining operations
	fmt.Println("\n4️⃣ Setting up mining operations...")

	// Load guild aggregate
	guildAggregate, err := repository.GetByID(ctx, guildID)
	if err != nil {
		return fmt.Errorf("failed to load guild: %w", err)
	}

	guild, ok := guildAggregate.(*domain.GuildAggregate)
	if !ok {
		return fmt.Errorf("invalid aggregate type")
	}

	// Initialize mining system and add mining nodes
	mining := guild.GetMining()

	// Add some mining nodes
	ironNode := &domain.MiningNode{
		NodeID:        "iron_mine_01",
		Name:          "Iron Ore Deposit",
		MineralType:   domain.MineralIron,
		Capacity:      5,
		Difficulty:    2,
		YieldRate:     10.0, // 10 iron per hour per worker
		IsActive:      true,
		RequiredLevel: 1,
	}

	goldNode := &domain.MiningNode{
		NodeID:        "gold_mine_01",
		Name:          "Gold Vein",
		MineralType:   domain.MineralGold,
		Capacity:      3,
		Difficulty:    5,
		YieldRate:     2.0, // 2 gold per hour per worker
		IsActive:      true,
		RequiredLevel: 3,
	}

	if err := mining.AddMiningNode(ironNode); err != nil {
		return fmt.Errorf("failed to add iron node: %w", err)
	}

	if err := mining.AddMiningNode(goldNode); err != nil {
		return fmt.Errorf("failed to add gold node: %w", err)
	}

	fmt.Printf("   ✅ Added mining nodes: %s, %s\n", ironNode.Name, goldNode.Name)

	// Step 5: Start mining operations
	fmt.Println("\n5️⃣ Starting mining operations...")

	// Start iron mining operation with both miners
	operationID1 := uuid.New().String()
	workerIDs := []string{miner1ID, miner2ID}

	err = guild.StartMiningOperation(operationID1, ironNode.NodeID, workerIDs, founderID)
	if err != nil {
		return fmt.Errorf("failed to start iron mining: %w", err)
	}

	// Save the guild after starting mining operation
	if err := repository.Save(ctx, guild, guild.OriginalVersion()); err != nil {
		return fmt.Errorf("failed to save guild after starting mining: %w", err)
	}

	fmt.Printf("   ✅ Started iron mining operation with %d workers\n", len(workerIDs))
	fmt.Printf("   ⛏️ Workers: %s, %s\n", miner1Username, miner2Username)

	// Step 5: Simulate time passing and harvest minerals
	fmt.Println("\n5️⃣ Simulating mining progress...")
	fmt.Println("   ⏰ Waiting for mining progress... (simulating 2 hours)")

	// In a real system, this would be actual time passage
	// For demo purposes, we'll simulate by directly updating the last harvest time
	time.Sleep(2 * time.Second) // Brief pause for demo effect

	// Harvest minerals
	harvested, err := guild.HarvestMinerals(operationID1, founderID)
	if err != nil {
		return fmt.Errorf("failed to harvest minerals: %w", err)
	}

	// Save the guild after harvesting
	if err := repository.Save(ctx, guild, guild.OriginalVersion()); err != nil {
		return fmt.Errorf("failed to save guild after harvesting: %w", err)
	}

	fmt.Printf("   ✅ Harvested minerals:\n")
	for mineralType, amount := range harvested {
		value := amount * mineralType.GetValue()
		fmt.Printf("      - %s: %d units (value: %d gold)\n", mineralType.String(), amount, value)
	}

	// Step 6: Display mining status
	fmt.Println("\n6️⃣ Mining status...")
	displayMiningStatus(guild)

	// Step 7: Start gold mining operation
	fmt.Println("\n7️⃣ Starting gold mining operation...")

	operationID2 := uuid.New().String()
	goldWorkers := []string{founderID} // Only founder for gold mining

	err = guild.StartMiningOperation(operationID2, goldNode.NodeID, goldWorkers, founderID)
	if err != nil {
		return fmt.Errorf("failed to start gold mining: %w", err)
	}

	// Save the guild
	if err := repository.Save(ctx, guild, guild.OriginalVersion()); err != nil {
		return fmt.Errorf("failed to save guild after starting gold mining: %w", err)
	}

	fmt.Printf("   ✅ Started gold mining operation\n")
	fmt.Printf("   ⛏️ Worker: %s\n", founderUsername)

	// Step 8: Final status
	fmt.Println("\n8️⃣ Final guild status...")
	if err := displayGuildStatus(ctx, queryDispatcher, guildID); err != nil {
		return fmt.Errorf("failed to display guild status: %w", err)
	}

	// Display final mining status
	fmt.Println("\n⛏️ Final mining status...")
	displayMiningStatus(guild)

	return nil
}

func displayMiningStatus(guild *domain.GuildAggregate) {
	mining := guild.GetMining()

	fmt.Printf("   🏭 Mining Level: %d (Experience: %d)\n", mining.MiningLevel, mining.MiningExperience)
	fmt.Printf("   💎 Total Mineral Value: %d gold\n", mining.GetTotalMineralValue())
	fmt.Printf("   🔄 Active Operations: %d\n", mining.GetActiveOperationsCount())

	fmt.Println("   📦 Mineral Inventory:")
	for mineralType, amount := range mining.MineralInventory {
		if amount > 0 {
			value := amount * mineralType.GetValue()
			fmt.Printf("      - %s: %d units (value: %d gold)\n", mineralType.String(), amount, value)
		}
	}

	fmt.Println("   🏗️ Available Mining Nodes:")
	for _, node := range mining.AvailableNodes {
		status := "🔴 Inactive"
		if node.IsActive {
			status = "🟢 Active"
		}
		fmt.Printf("      - %s (%s): %s, Capacity: %d, Yield: %.1f/hour\n",
			node.Name, node.MineralType.String(), status, node.Capacity, node.YieldRate)
	}
}

// getMessageFromResult extracts message from CommandResult
func getMessageFromResult(result *cqrs.CommandResult, defaultMessage string) string {
	if result.Data != nil {
		if data, ok := result.Data.(map[string]interface{}); ok {
			if msg, exists := data["message"]; exists {
				if msgStr, ok := msg.(string); ok {
					return msgStr
				}
			}
		}
	}
	return defaultMessage
}

func displayGuildStatus(ctx context.Context, queryDispatcher cqrs.QueryDispatcher, guildID string) error {
	// Create and execute guild query
	guildQuery := queries.NewGetGuildQuery(guildID)
	result, err := queryDispatcher.Dispatch(ctx, guildQuery)
	if err != nil {
		return fmt.Errorf("failed to query guild: %w", err)
	}

	if !result.Success {
		return fmt.Errorf("guild query failed: %v", result.Error)
	}

	// Extract guild view from result
	guildResult, ok := result.Data.(*queries.GuildQueryResult)
	if !ok {
		return fmt.Errorf("invalid query result type: expected *GuildQueryResult, got %T", result.Data)
	}

	if guildResult.Guild == nil {
		return fmt.Errorf("guild not found")
	}

	guild := guildResult.Guild

	fmt.Printf("   🏰 Guild: %s\n", guild.GetDisplayName())
	fmt.Printf("   📝 Description: %s\n", guild.Description)
	fmt.Printf("   👥 Members: %d/%d\n", guild.ActiveMemberCount, guild.MaxMembers)
	fmt.Printf("   💰 Treasury: %d gold\n", guild.Treasury)
	fmt.Printf("   ⭐ Level: %d\n", guild.Level)

	return nil
}
