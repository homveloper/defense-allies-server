package domain

import (
	"fmt"
	"strings"
	"time"
)

// UserProfile represents detailed user profile information
type UserProfile struct {
	// Basic information
	FirstName   string    `json:"first_name"`
	LastName    string    `json:"last_name"`
	DisplayName string    `json:"display_name"`
	Bio         string    `json:"bio"`
	Avatar      string    `json:"avatar"`
	
	// Contact information
	PhoneNumber string `json:"phone_number,omitempty"`
	Address     string `json:"address,omitempty"`
	City        string `json:"city,omitempty"`
	Country     string `json:"country,omitempty"`
	PostalCode  string `json:"postal_code,omitempty"`
	
	// Personal information
	DateOfBirth *time.Time `json:"date_of_birth,omitempty"`
	Gender      string     `json:"gender,omitempty"`
	Language    string     `json:"language"`
	Timezone    string     `json:"timezone"`
	
	// Preferences
	Preferences map[string]interface{} `json:"preferences"`
	
	// Social links
	SocialLinks map[string]string `json:"social_links,omitempty"`
	
	// Metadata
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// NewUserProfile creates a new UserProfile with default values
func NewUserProfile(firstName, lastName string) *UserProfile {
	displayName := strings.TrimSpace(firstName + " " + lastName)
	if displayName == "" {
		displayName = "User"
	}
	
	return &UserProfile{
		FirstName:   firstName,
		LastName:    lastName,
		DisplayName: displayName,
		Language:    "en",
		Timezone:    "UTC",
		Preferences: make(map[string]interface{}),
		SocialLinks: make(map[string]string),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
}

// UpdateBasicInfo updates basic profile information
func (p *UserProfile) UpdateBasicInfo(firstName, lastName, bio string) {
	p.FirstName = firstName
	p.LastName = lastName
	p.Bio = bio
	
	// Update display name
	displayName := strings.TrimSpace(firstName + " " + lastName)
	if displayName != "" {
		p.DisplayName = displayName
	}
	
	p.UpdatedAt = time.Now()
}

// UpdateDisplayName updates the display name
func (p *UserProfile) UpdateDisplayName(displayName string) error {
	if strings.TrimSpace(displayName) == "" {
		return fmt.Errorf("display name cannot be empty")
	}
	
	p.DisplayName = strings.TrimSpace(displayName)
	p.UpdatedAt = time.Now()
	return nil
}

// UpdateContactInfo updates contact information
func (p *UserProfile) UpdateContactInfo(phoneNumber, address, city, country, postalCode string) {
	p.PhoneNumber = phoneNumber
	p.Address = address
	p.City = city
	p.Country = country
	p.PostalCode = postalCode
	p.UpdatedAt = time.Now()
}

// UpdatePersonalInfo updates personal information
func (p *UserProfile) UpdatePersonalInfo(dateOfBirth *time.Time, gender, language, timezone string) {
	p.DateOfBirth = dateOfBirth
	p.Gender = gender
	
	if language != "" {
		p.Language = language
	}
	if timezone != "" {
		p.Timezone = timezone
	}
	
	p.UpdatedAt = time.Now()
}

// SetAvatar sets the avatar URL
func (p *UserProfile) SetAvatar(avatarURL string) {
	p.Avatar = avatarURL
	p.UpdatedAt = time.Now()
}

// SetPreference sets a user preference
func (p *UserProfile) SetPreference(key string, value interface{}) {
	if p.Preferences == nil {
		p.Preferences = make(map[string]interface{})
	}
	p.Preferences[key] = value
	p.UpdatedAt = time.Now()
}

// GetPreference gets a user preference
func (p *UserProfile) GetPreference(key string) (interface{}, bool) {
	if p.Preferences == nil {
		return nil, false
	}
	value, exists := p.Preferences[key]
	return value, exists
}

// RemovePreference removes a user preference
func (p *UserProfile) RemovePreference(key string) {
	if p.Preferences != nil {
		delete(p.Preferences, key)
		p.UpdatedAt = time.Now()
	}
}

// SetSocialLink sets a social media link
func (p *UserProfile) SetSocialLink(platform, url string) {
	if p.SocialLinks == nil {
		p.SocialLinks = make(map[string]string)
	}
	p.SocialLinks[platform] = url
	p.UpdatedAt = time.Now()
}

// GetSocialLink gets a social media link
func (p *UserProfile) GetSocialLink(platform string) (string, bool) {
	if p.SocialLinks == nil {
		return "", false
	}
	url, exists := p.SocialLinks[platform]
	return url, exists
}

// RemoveSocialLink removes a social media link
func (p *UserProfile) RemoveSocialLink(platform string) {
	if p.SocialLinks != nil {
		delete(p.SocialLinks, platform)
		p.UpdatedAt = time.Now()
	}
}

// GetFullName returns the full name
func (p *UserProfile) GetFullName() string {
	return strings.TrimSpace(p.FirstName + " " + p.LastName)
}

// GetAge calculates age from date of birth
func (p *UserProfile) GetAge() *int {
	if p.DateOfBirth == nil {
		return nil
	}
	
	now := time.Now()
	age := now.Year() - p.DateOfBirth.Year()
	
	// Adjust if birthday hasn't occurred this year
	if now.YearDay() < p.DateOfBirth.YearDay() {
		age--
	}
	
	return &age
}

// Validate validates the profile data
func (p *UserProfile) Validate() error {
	if strings.TrimSpace(p.DisplayName) == "" {
		return fmt.Errorf("display name cannot be empty")
	}
	
	if p.Language == "" {
		return fmt.Errorf("language cannot be empty")
	}
	
	if p.Timezone == "" {
		return fmt.Errorf("timezone cannot be empty")
	}
	
	// Validate date of birth is not in the future
	if p.DateOfBirth != nil && p.DateOfBirth.After(time.Now()) {
		return fmt.Errorf("date of birth cannot be in the future")
	}
	
	return nil
}

// Clone creates a deep copy of the profile
func (p *UserProfile) Clone() *UserProfile {
	clone := &UserProfile{
		FirstName:   p.FirstName,
		LastName:    p.LastName,
		DisplayName: p.DisplayName,
		Bio:         p.Bio,
		Avatar:      p.Avatar,
		PhoneNumber: p.PhoneNumber,
		Address:     p.Address,
		City:        p.City,
		Country:     p.Country,
		PostalCode:  p.PostalCode,
		Gender:      p.Gender,
		Language:    p.Language,
		Timezone:    p.Timezone,
		CreatedAt:   p.CreatedAt,
		UpdatedAt:   p.UpdatedAt,
	}
	
	// Copy date of birth
	if p.DateOfBirth != nil {
		dob := *p.DateOfBirth
		clone.DateOfBirth = &dob
	}
	
	// Copy preferences
	clone.Preferences = make(map[string]interface{})
	for k, v := range p.Preferences {
		clone.Preferences[k] = v
	}
	
	// Copy social links
	clone.SocialLinks = make(map[string]string)
	for k, v := range p.SocialLinks {
		clone.SocialLinks[k] = v
	}
	
	return clone
}
