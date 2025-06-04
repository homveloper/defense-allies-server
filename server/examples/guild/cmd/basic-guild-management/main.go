package main

import (
	"context"
	"fmt"
	"log"

	"github.com/google/uuid"

	"defense-allies-server/examples/guild/application/commands"
	"defense-allies-server/examples/guild/application/handlers"
	"defense-allies-server/examples/guild/infrastructure/projections"
	"defense-allies-server/examples/guild/infrastructure/queries"
	"defense-allies-server/examples/guild/infrastructure/repositories"
	"defense-allies-server/pkg/cqrs"
)

func main() {
	fmt.Println("🏰 Defense Allies - Guild Management System Example")
	fmt.Println("==================================================")

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

	// Run the guild management example
	if err := runGuildExample(ctx, commandDispatcher, queryDispatcher, projectionManager); err != nil {
		log.Fatalf("Example failed: %v", err)
	}

	fmt.Println("\n🎉 Guild management example completed successfully!")
}

func runGuildExample(ctx context.Context, dispatcher cqrs.CommandDispatcher, queryDispatcher cqrs.QueryDispatcher, projectionManager cqrs.ProjectionManager) error {
	// Generate IDs
	guildID := uuid.New().String()
	founderID := "founder123"
	founderUsername := "GuildMaster"
	member1ID := "member001"
	member1Username := "Warrior"
	member2ID := "member002"
	member2Username := "Mage"

	fmt.Printf("\n🏰 Creating guild with ID: %s\n", guildID)

	// Step 1: Create guild
	fmt.Println("\n1️⃣ Creating guild...")
	createCmd := commands.NewCreateGuildCommand(guildID, "Elite Warriors", "A guild for elite warriors", founderID, founderUsername)
	result, err := dispatcher.Dispatch(ctx, createCmd)
	if err != nil {
		return fmt.Errorf("failed to create guild: %w", err)
	}
	fmt.Printf("   ✅ Guild created: %s\n", getMessageFromResult(result, "Guild created successfully"))

	// Step 2: Update guild info
	fmt.Println("\n2️⃣ Updating guild info...")
	updateInfoCmd := commands.NewUpdateGuildInfoCommand(guildID, "Elite Warriors Guild", "A prestigious guild for elite warriors and adventurers", "Welcome to our guild! Check the rules.", "[EW]", founderID)
	result, err = dispatcher.Dispatch(ctx, updateInfoCmd)
	if err != nil {
		return fmt.Errorf("failed to update guild info: %w", err)
	}
	fmt.Printf("   ✅ Guild info updated: %s\n", getMessageFromResult(result, "Guild info updated successfully"))

	// Step 3: Update guild settings
	fmt.Println("\n3️⃣ Updating guild settings...")
	updateSettingsCmd := commands.NewUpdateGuildSettingsCommand(guildID, 100, 5, true, false, founderID)
	result, err = dispatcher.Dispatch(ctx, updateSettingsCmd)
	if err != nil {
		return fmt.Errorf("failed to update guild settings: %w", err)
	}
	fmt.Printf("   ✅ Guild settings updated: %s\n", getMessageFromResult(result, "Guild settings updated successfully"))

	// Step 4: Invite first member
	fmt.Println("\n4️⃣ Inviting first member...")
	inviteCmd1 := commands.NewInviteMemberCommand(guildID, member1ID, member1Username, founderID)
	result, err = dispatcher.Dispatch(ctx, inviteCmd1)
	if err != nil {
		return fmt.Errorf("failed to invite member: %w", err)
	}
	fmt.Printf("   ✅ Member invited: %s\n", getMessageFromResult(result, "Member invited successfully"))
	fmt.Printf("   👤 Invited: %s (%s)\n", member1Username, member1ID)

	// Step 5: Accept first invitation
	fmt.Println("\n5️⃣ Accepting first invitation...")
	acceptCmd1 := commands.NewAcceptInvitationCommand(guildID, member1ID)
	result, err = dispatcher.Dispatch(ctx, acceptCmd1)
	if err != nil {
		return fmt.Errorf("failed to accept invitation: %w", err)
	}
	fmt.Printf("   ✅ Invitation accepted: %s\n", getMessageFromResult(result, "Invitation accepted successfully"))

	// Step 6: Invite second member
	fmt.Println("\n6️⃣ Inviting second member...")
	inviteCmd2 := commands.NewInviteMemberCommand(guildID, member2ID, member2Username, founderID)
	result, err = dispatcher.Dispatch(ctx, inviteCmd2)
	if err != nil {
		return fmt.Errorf("failed to invite second member: %w", err)
	}
	fmt.Printf("   ✅ Member invited: %s\n", getMessageFromResult(result, "Member invited successfully"))
	fmt.Printf("   👤 Invited: %s (%s)\n", member2Username, member2ID)

	// Step 7: Accept second invitation
	fmt.Println("\n7️⃣ Accepting second invitation...")
	acceptCmd2 := commands.NewAcceptInvitationCommand(guildID, member2ID)
	result, err = dispatcher.Dispatch(ctx, acceptCmd2)
	if err != nil {
		return fmt.Errorf("failed to accept second invitation: %w", err)
	}
	fmt.Printf("   ✅ Second invitation accepted: %s\n", getMessageFromResult(result, "Invitation accepted successfully"))

	// Step 8: Promote first member to officer
	fmt.Println("\n8️⃣ Promoting first member to officer...")
	promoteCmd := commands.NewPromoteMemberCommand(guildID, member1ID, "Officer", founderID)
	result, err = dispatcher.Dispatch(ctx, promoteCmd)
	if err != nil {
		return fmt.Errorf("failed to promote member: %w", err)
	}
	fmt.Printf("   ✅ Member promoted: %s\n", getMessageFromResult(result, "Member promoted successfully"))
	fmt.Printf("   👑 %s is now an Officer\n", member1Username)

	// Step 9: Display guild status
	fmt.Println("\n9️⃣ Displaying guild status...")
	if err := displayGuildStatus(ctx, queryDispatcher, guildID); err != nil {
		return fmt.Errorf("failed to display guild status: %w", err)
	}

	// Step 10: Kick second member
	fmt.Println("\n🔟 Kicking second member...")
	kickCmd := commands.NewKickMemberCommand(guildID, member2ID, founderID, "Inactive player")
	result, err = dispatcher.Dispatch(ctx, kickCmd)
	if err != nil {
		return fmt.Errorf("failed to kick member: %w", err)
	}
	fmt.Printf("   ✅ Member kicked: %s\n", getMessageFromResult(result, "Member kicked successfully"))
	fmt.Printf("   👢 %s has been removed from the guild\n", member2Username)

	// Final status
	fmt.Println("\n📊 Final guild status...")
	if err := displayGuildStatus(ctx, queryDispatcher, guildID); err != nil {
		return fmt.Errorf("failed to display final guild status: %w", err)
	}

	return nil
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
	fmt.Printf("   📊 Status: %s\n", guild.Status)
	fmt.Printf("   👥 Members: %d/%d (%.1f%%)\n",
		guild.ActiveMemberCount, guild.MaxMembers, guild.GetMemberCapacityPercentage())
	fmt.Printf("   💰 Treasury: %d\n", guild.Treasury)
	fmt.Printf("   ⭐ Level: %d\n", guild.Level)
	fmt.Printf("   🏛️ Founded: %s\n", guild.FoundedAt.Format("2006-01-02"))
	fmt.Printf("   👑 Founder: %s\n", guild.FounderUsername)

	if guild.Notice != "" {
		fmt.Printf("   📢 Notice: %s\n", guild.Notice)
	}

	// Query guild members
	membersQuery := queries.NewGetGuildMembersQuery(guildID).WithStatus("Active")
	membersResult, err := queryDispatcher.Dispatch(ctx, membersQuery)
	if err != nil {
		fmt.Printf("   ⚠️  Failed to load members: %v\n", err)
		return nil
	}

	if membersResult.Success {
		if memberData, ok := membersResult.Data.(*queries.GuildQueryResult); ok && len(memberData.Members) > 0 {
			fmt.Println("   👥 Active Members:")
			for _, member := range memberData.Members {
				fmt.Printf("      - %s (%s) - Role: %s, Days: %d\n",
					member.Username, member.UserID, member.Role, member.DaysInGuild)
			}
		}
	}

	return nil
}
