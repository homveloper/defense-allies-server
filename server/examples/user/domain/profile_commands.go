package domain

import (
	"fmt"
	"strings"

	"cqrs"
)

// UpdateProfileCommand represents a command to update user profile
type UpdateProfileCommand struct {
	*cqrs.BaseCommand
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Bio       string `json:"bio"`
}

// NewUpdateProfileCommand creates a new UpdateProfileCommand
func NewUpdateProfileCommand(userID, firstName, lastName, bio string) *UpdateProfileCommand {
	return &UpdateProfileCommand{
		BaseCommand: cqrs.NewBaseCommand(
			UpdateProfileCommandType,
			userID,
			"User",
			map[string]interface{}{
				"first_name": firstName,
				"last_name":  lastName,
				"bio":        bio,
			},
		),
		FirstName: firstName,
		LastName:  lastName,
		Bio:       bio,
	}
}

// Validate validates the UpdateProfileCommand
func (c *UpdateProfileCommand) Validate() error {
	if err := c.BaseCommand.Validate(); err != nil {
		return err
	}

	if strings.TrimSpace(c.FirstName) == "" && strings.TrimSpace(c.LastName) == "" {
		return fmt.Errorf("at least first name or last name must be provided")
	}

	if len(c.Bio) > 500 {
		return fmt.Errorf("bio cannot exceed 500 characters")
	}

	return nil
}

// UpdateDisplayNameCommand represents a command to update display name
type UpdateDisplayNameCommand struct {
	*cqrs.BaseCommand
	DisplayName string `json:"display_name"`
}

// NewUpdateDisplayNameCommand creates a new UpdateDisplayNameCommand
func NewUpdateDisplayNameCommand(userID, displayName string) *UpdateDisplayNameCommand {
	return &UpdateDisplayNameCommand{
		BaseCommand: cqrs.NewBaseCommand(
			UpdateDisplayNameCommandType,
			userID,
			"User",
			map[string]interface{}{
				"display_name": displayName,
			},
		),
		DisplayName: displayName,
	}
}

// Validate validates the UpdateDisplayNameCommand
func (c *UpdateDisplayNameCommand) Validate() error {
	if err := c.BaseCommand.Validate(); err != nil {
		return err
	}

	if strings.TrimSpace(c.DisplayName) == "" {
		return fmt.Errorf("display name cannot be empty")
	}

	if len(c.DisplayName) > 50 {
		return fmt.Errorf("display name cannot exceed 50 characters")
	}

	return nil
}

// UpdateContactInfoCommand represents a command to update contact information
type UpdateContactInfoCommand struct {
	*cqrs.BaseCommand
	PhoneNumber string `json:"phone_number"`
	Address     string `json:"address"`
	City        string `json:"city"`
	Country     string `json:"country"`
	PostalCode  string `json:"postal_code"`
}

// NewUpdateContactInfoCommand creates a new UpdateContactInfoCommand
func NewUpdateContactInfoCommand(userID, phoneNumber, address, city, country, postalCode string) *UpdateContactInfoCommand {
	return &UpdateContactInfoCommand{
		BaseCommand: cqrs.NewBaseCommand(
			UpdateContactInfoCommandType,
			userID,
			"User",
			map[string]interface{}{
				"phone_number": phoneNumber,
				"address":      address,
				"city":         city,
				"country":      country,
				"postal_code":  postalCode,
			},
		),
		PhoneNumber: phoneNumber,
		Address:     address,
		City:        city,
		Country:     country,
		PostalCode:  postalCode,
	}
}

// Validate validates the UpdateContactInfoCommand
func (c *UpdateContactInfoCommand) Validate() error {
	if err := c.BaseCommand.Validate(); err != nil {
		return err
	}

	// Optional validation for phone number format, address length, etc.
	if len(c.Address) > 200 {
		return fmt.Errorf("address cannot exceed 200 characters")
	}

	if len(c.City) > 100 {
		return fmt.Errorf("city cannot exceed 100 characters")
	}

	if len(c.Country) > 100 {
		return fmt.Errorf("country cannot exceed 100 characters")
	}

	if len(c.PostalCode) > 20 {
		return fmt.Errorf("postal code cannot exceed 20 characters")
	}

	return nil
}

// SetAvatarCommand represents a command to set user avatar
type SetAvatarCommand struct {
	*cqrs.BaseCommand
	AvatarURL string `json:"avatar_url"`
}

// NewSetAvatarCommand creates a new SetAvatarCommand
func NewSetAvatarCommand(userID, avatarURL string) *SetAvatarCommand {
	return &SetAvatarCommand{
		BaseCommand: cqrs.NewBaseCommand(
			SetAvatarCommandType,
			userID,
			"User",
			map[string]interface{}{
				"avatar_url": avatarURL,
			},
		),
		AvatarURL: avatarURL,
	}
}

// Validate validates the SetAvatarCommand
func (c *SetAvatarCommand) Validate() error {
	if err := c.BaseCommand.Validate(); err != nil {
		return err
	}

	if c.AvatarURL == "" {
		return fmt.Errorf("avatar URL cannot be empty")
	}

	if len(c.AvatarURL) > 500 {
		return fmt.Errorf("avatar URL cannot exceed 500 characters")
	}

	return nil
}

// SetPreferenceCommand represents a command to set user preference
type SetPreferenceCommand struct {
	*cqrs.BaseCommand
	Key   string      `json:"key"`
	Value interface{} `json:"value"`
}

// NewSetPreferenceCommand creates a new SetPreferenceCommand
func NewSetPreferenceCommand(userID, key string, value interface{}) *SetPreferenceCommand {
	return &SetPreferenceCommand{
		BaseCommand: cqrs.NewBaseCommand(
			SetPreferenceCommandType,
			userID,
			"User",
			map[string]interface{}{
				"key":   key,
				"value": value,
			},
		),
		Key:   key,
		Value: value,
	}
}

// Validate validates the SetPreferenceCommand
func (c *SetPreferenceCommand) Validate() error {
	if err := c.BaseCommand.Validate(); err != nil {
		return err
	}

	if strings.TrimSpace(c.Key) == "" {
		return fmt.Errorf("preference key cannot be empty")
	}

	if len(c.Key) > 100 {
		return fmt.Errorf("preference key cannot exceed 100 characters")
	}

	return nil
}
