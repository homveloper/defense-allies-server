package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"

	"cqrs"
	"defense-allies-server/examples/guild/application/commands"
	"defense-allies-server/examples/guild/application/handlers"
	"defense-allies-server/examples/guild/domain"
	"defense-allies-server/examples/guild/infrastructure/projections"
	"defense-allies-server/examples/guild/infrastructure/queries"
	"defense-allies-server/examples/guild/infrastructure/repositories"
)

func main() {
	fmt.Println("🚛 Defense Allies - Guild Transport System Example")
	fmt.Println("=================================================")

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

	// Run the guild transport example
	if err := runGuildTransportExample(ctx, commandDispatcher, queryDispatcher, repository); err != nil {
		log.Fatalf("Example failed: %v", err)
	}

	fmt.Println("\n🎉 Guild transport example completed successfully!")
}

func runGuildTransportExample(ctx context.Context, dispatcher cqrs.CommandDispatcher, queryDispatcher cqrs.QueryDispatcher, repository cqrs.EventSourcedRepository) error {
	// Generate IDs
	guildID := uuid.New().String()
	founderID := "founder123"
	founderUsername := "TransportMaster"
	transporter1ID := "transporter001"
	transporter1Username := "CargoHauler"
	transporter2ID := "transporter002"
	transporter2Username := "FastDelivery"
	transporter3ID := "transporter003"
	transporter3Username := "HeavyLifter"

	fmt.Printf("\n🚛 Creating transport guild with ID: %s\n", guildID)

	// Step 1: Create guild
	fmt.Println("\n1️⃣ Creating transport guild...")
	createCmd := commands.NewCreateGuildCommand(guildID, "Transport Alliance", "A guild specialized in mineral transport operations", founderID, founderUsername)
	result, err := dispatcher.Dispatch(ctx, createCmd)
	if err != nil {
		return fmt.Errorf("failed to create guild: %w", err)
	}
	fmt.Printf("   ✅ Transport guild created: %s\n", getMessageFromResult(result, "Guild created successfully"))

	// Step 2: Update guild settings to allow more members
	fmt.Println("\n2️⃣ Updating guild settings...")
	updateSettingsCmd := commands.NewUpdateGuildSettingsCommand(guildID, 20, 1, true, false, founderID)
	result, err = dispatcher.Dispatch(ctx, updateSettingsCmd)
	if err != nil {
		return fmt.Errorf("failed to update guild settings: %w", err)
	}
	fmt.Printf("   ✅ Guild settings updated: %s\n", getMessageFromResult(result, "Guild settings updated successfully"))

	// Step 3: Add transporters to the guild
	fmt.Println("\n3️⃣ Recruiting transporters...")

	// Invite and accept transporters
	transporters := []struct {
		ID       string
		Username string
	}{
		{transporter1ID, transporter1Username},
		{transporter2ID, transporter2Username},
		{transporter3ID, transporter3Username},
	}

	for _, transporter := range transporters {
		// Invite transporter
		inviteCmd := commands.NewInviteMemberCommand(guildID, transporter.ID, transporter.Username, founderID)
		result, err = dispatcher.Dispatch(ctx, inviteCmd)
		if err != nil {
			return fmt.Errorf("failed to invite transporter: %w", err)
		}
		fmt.Printf("   ✅ Transporter invited: %s\n", transporter.Username)

		// Accept invitation
		acceptCmd := commands.NewAcceptInvitationCommand(guildID, transporter.ID)
		result, err = dispatcher.Dispatch(ctx, acceptCmd)
		if err != nil {
			return fmt.Errorf("failed to accept invitation: %w", err)
		}
		fmt.Printf("   ✅ %s joined the guild\n", transporter.Username)
	}

	// Step 4: Load guild and create transport recruitment
	fmt.Println("\n4️⃣ Creating transport recruitment posting...")

	// Load guild aggregate
	guildAggregate, err := repository.GetByID(ctx, guildID)
	if err != nil {
		return fmt.Errorf("failed to load guild: %w", err)
	}

	guild, ok := guildAggregate.(*domain.GuildAggregate)
	if !ok {
		return fmt.Errorf("invalid aggregate type")
	}

	// Create transport recruitment
	recruitmentID := uuid.New().String()
	title := "Iron Ore Transport Mission"
	description := "Transport valuable iron ore from the northern mines to the guild treasury. High reward for participants!"
	maxParticipants := 3
	minParticipants := 2
	duration := 5 * time.Minute      // Recruitment open for 5 minutes
	transportTime := 3 * time.Minute // Transport takes 3 minutes

	// Define cargo to transport
	totalCargo := map[domain.MineralType]int64{
		domain.MineralIron: 100,
		domain.MineralGold: 10,
	}

	err = guild.CreateTransportRecruitment(recruitmentID, title, description,
		maxParticipants, minParticipants, duration, transportTime, totalCargo, founderID)
	if err != nil {
		return fmt.Errorf("failed to create transport recruitment: %w", err)
	}

	// Save the guild
	if err := repository.Save(ctx, guild, guild.OriginalVersion()); err != nil {
		return fmt.Errorf("failed to save guild after creating recruitment: %w", err)
	}

	fmt.Printf("   ✅ Transport recruitment created: %s\n", title)
	fmt.Printf("   📦 Cargo: %d Iron, %d Gold\n", totalCargo[domain.MineralIron], totalCargo[domain.MineralGold])
	fmt.Printf("   👥 Participants: %d-%d\n", minParticipants, maxParticipants)
	fmt.Printf("   ⏰ Duration: %v\n", duration)

	// Step 5: Transporters join the recruitment
	fmt.Println("\n5️⃣ Transporters joining recruitment...")

	// Reload guild to get latest state
	guildAggregate, err = repository.GetByID(ctx, guildID)
	if err != nil {
		return fmt.Errorf("failed to reload guild: %w", err)
	}
	guild = guildAggregate.(*domain.GuildAggregate)

	// First transporter joins
	err = guild.JoinTransportRecruitment(recruitmentID, transporter1ID)
	if err != nil {
		return fmt.Errorf("failed to join recruitment: %w", err)
	}
	fmt.Printf("   ✅ %s joined the transport mission\n", transporter1Username)

	// Second transporter joins
	err = guild.JoinTransportRecruitment(recruitmentID, transporter2ID)
	if err != nil {
		return fmt.Errorf("failed to join recruitment: %w", err)
	}
	fmt.Printf("   ✅ %s joined the transport mission\n", transporter2Username)

	// Save the guild
	if err := repository.Save(ctx, guild, guild.OriginalVersion()); err != nil {
		return fmt.Errorf("failed to save guild after joining recruitment: %w", err)
	}

	// Step 6: Display recruitment status
	fmt.Println("\n6️⃣ Recruitment status...")
	displayRecruitmentStatus(guild, recruitmentID)

	// Step 7: Start transport operation
	fmt.Println("\n7️⃣ Starting transport operation...")

	// Reload guild to get latest state
	guildAggregate, err = repository.GetByID(ctx, guildID)
	if err != nil {
		return fmt.Errorf("failed to reload guild: %w", err)
	}
	guild = guildAggregate.(*domain.GuildAggregate)

	transportID := uuid.New().String()
	err = guild.StartTransportFromRecruitment(recruitmentID, transportID, founderID)
	if err != nil {
		return fmt.Errorf("failed to start transport: %w", err)
	}

	// Save the guild
	if err := repository.Save(ctx, guild, guild.OriginalVersion()); err != nil {
		return fmt.Errorf("failed to save guild after starting transport: %w", err)
	}

	fmt.Printf("   ✅ Transport operation started with ID: %s\n", transportID)
	fmt.Printf("   🚛 Transport in progress...\n")

	// Step 8: Simulate transport time
	fmt.Println("\n8️⃣ Simulating transport progress...")
	fmt.Println("   ⏰ Waiting for transport to complete... (simulating 3 minutes)")

	// In a real system, this would be actual time passage
	// For demo purposes, we'll simulate by waiting briefly
	time.Sleep(3 * time.Second) // Brief pause for demo effect

	// Step 9: Complete transport and distribute rewards
	fmt.Println("\n9️⃣ Completing transport and distributing rewards...")

	// Reload guild to get latest state
	guildAggregate, err = repository.GetByID(ctx, guildID)
	if err != nil {
		return fmt.Errorf("failed to reload guild: %w", err)
	}
	guild = guildAggregate.(*domain.GuildAggregate)

	rewards, err := guild.ForceCompleteTransportRecruitment(recruitmentID, founderID)
	if err != nil {
		return fmt.Errorf("failed to complete transport: %w", err)
	}

	// Save the guild
	if err := repository.Save(ctx, guild, guild.OriginalVersion()); err != nil {
		return fmt.Errorf("failed to save guild after completing transport: %w", err)
	}

	fmt.Printf("   ✅ Transport completed successfully!\n")
	fmt.Printf("   💰 Rewards distributed:\n")
	for userID, userRewards := range rewards {
		// Get username from participants
		recruitment, _ := guild.GetTransportRecruitment(recruitmentID)
		username := userID
		if participant, exists := recruitment.Participants[userID]; exists {
			username = participant.Username
		}

		fmt.Printf("      - %s:\n", username)
		for mineralType, amount := range userRewards {
			value := amount * mineralType.GetValue()
			fmt.Printf("        • %s: %d units (value: %d gold)\n", mineralType.String(), amount, value)
		}
	}

	// Step 10: Final status
	fmt.Println("\n🔟 Final guild status...")
	if err := displayGuildStatus(ctx, queryDispatcher, guildID); err != nil {
		return fmt.Errorf("failed to display guild status: %w", err)
	}

	return nil
}

func displayRecruitmentStatus(guild *domain.GuildAggregate, recruitmentID string) {
	recruitment, exists := guild.GetTransportRecruitment(recruitmentID)
	if !exists {
		fmt.Printf("   ❌ Recruitment %s not found\n", recruitmentID)
		return
	}

	fmt.Printf("   🚛 Transport Mission: %s\n", recruitment.Title)
	fmt.Printf("   📝 Description: %s\n", recruitment.Description)
	fmt.Printf("   📊 Status: %s\n", recruitment.Status.String())
	fmt.Printf("   👥 Participants: %d/%d (min: %d)\n",
		recruitment.GetParticipantCount(), recruitment.MaxParticipants, recruitment.MinParticipants)

	if recruitment.GetParticipantCount() > 0 {
		fmt.Println("   👤 Current participants:")
		for _, participant := range recruitment.Participants {
			fmt.Printf("      - %s (joined: %s)\n",
				participant.Username, participant.JoinedAt.Format("15:04:05"))
		}
	}

	fmt.Printf("   📦 Total cargo value: %d gold\n", calculateCargoValue(recruitment.TotalCargo))
	fmt.Printf("   💰 Reward per person: %d gold\n", calculateCargoValue(recruitment.RewardPerPerson))

	if recruitment.Status == domain.RecruitmentStatusOpen {
		fmt.Printf("   ⏰ Time remaining: %v\n", recruitment.GetRemainingTime())
		if recruitment.CanStart() {
			fmt.Printf("   🟢 Ready to start transport!\n")
		} else {
			fmt.Printf("   🟡 Waiting for more participants...\n")
		}
	}
}

func calculateCargoValue(cargo map[domain.MineralType]int64) int64 {
	total := int64(0)
	for mineralType, amount := range cargo {
		total += amount * mineralType.GetValue()
	}
	return total
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
