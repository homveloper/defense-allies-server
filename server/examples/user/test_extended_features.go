package main

import (
	"fmt"
	"log"
	"time"

	"defense-allies-server/examples/user/domain"

	"github.com/google/uuid"
)

func main() {
	fmt.Println("🚀 Testing Extended User Features")
	fmt.Println("=================================")

	// Create a new user
	userID := uuid.New().String()
	user, err := domain.NewUser(userID, "john.doe@example.com", "John Doe")
	if err != nil {
		log.Fatalf("Failed to create user: %v", err)
	}

	fmt.Printf("✅ Created user: %s (%s)\n", user.Name(), user.Email())

	// Test role management
	fmt.Println("\n🎭 Testing Role Management")
	fmt.Println("==========================")

	// Check initial roles
	roles := user.GetRoles()
	fmt.Printf("📋 Initial roles: %d\n", len(roles))
	for _, role := range roles {
		fmt.Printf("   - %s (assigned by: %s)\n", role.Name, role.AssignedBy)
	}

	// Assign admin role
	err = user.AssignRole(domain.RoleTypeAdmin, "system-admin")
	if err != nil {
		log.Fatalf("Failed to assign admin role: %v", err)
	}
	fmt.Printf("✅ Assigned admin role\n")

	// Assign beta tester role with expiry
	expiryTime := time.Now().Add(30 * 24 * time.Hour) // 30 days
	err = user.AssignRoleWithExpiry(domain.RoleTypeBetaTester, "system-admin", expiryTime)
	if err != nil {
		log.Fatalf("Failed to assign beta tester role: %v", err)
	}
	fmt.Printf("✅ Assigned beta tester role (expires: %s)\n", expiryTime.Format("2006-01-02"))

	// Check updated roles
	roles = user.GetRoles()
	fmt.Printf("📋 Updated roles: %d\n", len(roles))
	for _, role := range roles {
		expiry := "never"
		if role.ExpiresAt != nil {
			expiry = role.ExpiresAt.Format("2006-01-02")
		}
		fmt.Printf("   - %s (assigned by: %s, expires: %s)\n", role.Name, role.AssignedBy, expiry)
	}

	// Check permissions
	permissions := user.GetPermissions()
	fmt.Printf("🔐 Total permissions: %d\n", len(permissions))
	fmt.Printf("   Has admin permission: %t\n", user.HasPermission("user.*"))
	fmt.Printf("   Has game test permission: %t\n", user.HasPermission("game.test"))

	// Test profile management
	fmt.Println("\n👤 Testing Profile Management")
	fmt.Println("=============================")

	// Update basic profile
	err = user.UpdateProfile("John", "Doe", "Software developer passionate about gaming")
	if err != nil {
		log.Fatalf("Failed to update profile: %v", err)
	}
	fmt.Printf("✅ Updated basic profile\n")

	// Update display name
	err = user.UpdateDisplayName("JohnD")
	if err != nil {
		log.Fatalf("Failed to update display name: %v", err)
	}
	fmt.Printf("✅ Updated display name\n")

	// Update contact info
	err = user.UpdateContactInfo("+1-555-0123", "123 Main St", "San Francisco", "USA", "94102")
	if err != nil {
		log.Fatalf("Failed to update contact info: %v", err)
	}
	fmt.Printf("✅ Updated contact info\n")

	// Set avatar
	err = user.SetAvatar("https://example.com/avatars/john-doe.jpg")
	if err != nil {
		log.Fatalf("Failed to set avatar: %v", err)
	}
	fmt.Printf("✅ Set avatar\n")

	// Set preferences
	err = user.SetPreference("theme", "dark")
	if err != nil {
		log.Fatalf("Failed to set theme preference: %v", err)
	}
	err = user.SetPreference("notifications", true)
	if err != nil {
		log.Fatalf("Failed to set notifications preference: %v", err)
	}
	fmt.Printf("✅ Set preferences\n")

	// Display profile information
	profile := user.GetProfile()
	fmt.Printf("📋 Profile Information:\n")
	fmt.Printf("   Name: %s %s\n", profile.FirstName, profile.LastName)
	fmt.Printf("   Display Name: %s\n", profile.DisplayName)
	fmt.Printf("   Bio: %s\n", profile.Bio)
	fmt.Printf("   Phone: %s\n", profile.PhoneNumber)
	fmt.Printf("   Address: %s, %s, %s %s\n", profile.Address, profile.City, profile.Country, profile.PostalCode)
	fmt.Printf("   Avatar: %s\n", profile.Avatar)

	if theme, exists := profile.GetPreference("theme"); exists {
		fmt.Printf("   Theme: %v\n", theme)
	}
	if notifications, exists := profile.GetPreference("notifications"); exists {
		fmt.Printf("   Notifications: %v\n", notifications)
	}

	// Test role revocation
	fmt.Println("\n🚫 Testing Role Revocation")
	fmt.Println("==========================")

	err = user.RevokeRole(domain.RoleTypeBetaTester, "system-admin")
	if err != nil {
		log.Fatalf("Failed to revoke beta tester role: %v", err)
	}
	fmt.Printf("✅ Revoked beta tester role\n")

	// Check final roles
	roles = user.GetRoles()
	fmt.Printf("📋 Final roles: %d\n", len(roles))
	for _, role := range roles {
		fmt.Printf("   - %s (assigned by: %s)\n", role.Name, role.AssignedBy)
	}

	// Test event history
	fmt.Println("\n📜 Testing Event History")
	fmt.Println("========================")

	changes := user.GetChanges()
	fmt.Printf("📋 Total events generated: %d\n", len(changes))
	for i, event := range changes {
		fmt.Printf("   %d. %s (version: %d)\n", i+1, event.EventType(), event.Version())
	}

	// Test aggregate validation
	fmt.Println("\n✅ Testing Validation")
	fmt.Println("=====================")

	err = user.Validate()
	if err != nil {
		log.Fatalf("User validation failed: %v", err)
	}
	fmt.Printf("✅ User aggregate is valid\n")

	err = profile.Validate()
	if err != nil {
		log.Fatalf("Profile validation failed: %v", err)
	}
	fmt.Printf("✅ Profile is valid\n")

	fmt.Println("\n🎉 All extended features tested successfully!")
}
