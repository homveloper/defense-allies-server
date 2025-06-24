package pagit

// SortDirection represents sort direction
type SortDirection int

const (
	SortAsc SortDirection = iota
	SortDesc
)

func (s SortDirection) String() string {
	switch s {
	case SortAsc:
		return "ASC"
	case SortDesc:
		return "DESC"
	default:
		return "ASC"
	}
}

// SortField represents a single sort field
type SortField struct {
	Field     string
	Direction SortDirection
}

// SortConfig represents multiple sort fields
type SortConfig []SortField

// NewSort creates a new sort configuration
func NewSort() SortConfig {
	return SortConfig{}
}

// Asc adds an ascending sort field
func (s SortConfig) Asc(field string) SortConfig {
	return append(s, SortField{Field: field, Direction: SortAsc})
}

// Desc adds a descending sort field
func (s SortConfig) Desc(field string) SortConfig {
	return append(s, SortField{Field: field, Direction: SortDesc})
}

// Primary returns the primary sort field (first one)
func (s SortConfig) Primary() SortField {
	if len(s) == 0 {
		return SortField{Field: "id", Direction: SortDesc} // Default
	}
	return s[0]
}

// IsEmpty returns true if no sort fields are configured
func (s SortConfig) IsEmpty() bool {
	return len(s) == 0
}

// Default sort configurations
var (
	SortByIDDesc    = NewSort().Desc("id")
	SortByIDASC     = NewSort().Asc("id")
	SortByCreatedAt = NewSort().Desc("created_at")
	SortByUpdatedAt = NewSort().Desc("updated_at")
	SortByName      = NewSort().Asc("name")
)
