package domain

import (
	"fmt"
	"time"
)

// MineralType represents different types of minerals
type MineralType int

const (
	// MineralIron represents iron ore
	MineralIron MineralType = iota
	// MineralGold represents gold ore
	MineralGold
	// MineralSilver represents silver ore
	MineralSilver
	// MineralCopper represents copper ore
	MineralCopper
	// MineralDiamond represents diamond
	MineralDiamond
	// MineralMithril represents mithril (rare)
	MineralMithril
)

// String returns the string representation of the mineral type
func (m MineralType) String() string {
	switch m {
	case MineralIron:
		return "Iron"
	case MineralGold:
		return "Gold"
	case MineralSilver:
		return "Silver"
	case MineralCopper:
		return "Copper"
	case MineralDiamond:
		return "Diamond"
	case MineralMithril:
		return "Mithril"
	default:
		return "Unknown"
	}
}

// GetValue returns the base value of the mineral per unit
func (m MineralType) GetValue() int64 {
	switch m {
	case MineralIron:
		return 10
	case MineralCopper:
		return 15
	case MineralSilver:
		return 50
	case MineralGold:
		return 100
	case MineralDiamond:
		return 500
	case MineralMithril:
		return 1000
	default:
		return 1
	}
}

// MineralDeposit represents a deposit of a specific mineral in a mine
type MineralDeposit struct {
	Type      MineralType `json:"type"`
	Amount    int64       `json:"amount"`    // Total amount available
	Extracted int64       `json:"extracted"` // Amount already extracted
	Quality   float64     `json:"quality"`   // Quality multiplier (0.5 - 2.0)
}

// GetRemainingAmount returns the remaining amount to be extracted
func (d *MineralDeposit) GetRemainingAmount() int64 {
	return d.Amount - d.Extracted
}

// GetValue returns the total value of the remaining deposit
func (d *MineralDeposit) GetValue() int64 {
	remaining := d.GetRemainingAmount()
	baseValue := d.Type.GetValue()
	return int64(float64(remaining*baseValue) * d.Quality)
}

// Extract extracts a specified amount from the deposit
func (d *MineralDeposit) Extract(amount int64) (int64, error) {
	remaining := d.GetRemainingAmount()
	if amount > remaining {
		amount = remaining
	}
	if amount <= 0 {
		return 0, fmt.Errorf("no minerals to extract")
	}
	d.Extracted += amount
	return amount, nil
}

// WorkerType represents different types of workers
type WorkerType int

const (
	// WorkerBasic represents a basic worker
	WorkerBasic WorkerType = iota
	// WorkerSkilled represents a skilled worker
	WorkerSkilled
	// WorkerExpert represents an expert worker
	WorkerExpert
	// WorkerMaster represents a master worker
	WorkerMaster
)

// String returns the string representation of the worker type
func (w WorkerType) String() string {
	switch w {
	case WorkerBasic:
		return "Basic"
	case WorkerSkilled:
		return "Skilled"
	case WorkerExpert:
		return "Expert"
	case WorkerMaster:
		return "Master"
	default:
		return "Unknown"
	}
}

// GetEfficiency returns the efficiency multiplier for the worker type
func (w WorkerType) GetEfficiency() float64 {
	switch w {
	case WorkerBasic:
		return 1.0
	case WorkerSkilled:
		return 1.5
	case WorkerExpert:
		return 2.0
	case WorkerMaster:
		return 3.0
	default:
		return 1.0
	}
}

// GetCost returns the cost per hour for the worker type
func (w WorkerType) GetCost() int64 {
	switch w {
	case WorkerBasic:
		return 10
	case WorkerSkilled:
		return 20
	case WorkerExpert:
		return 40
	case WorkerMaster:
		return 80
	default:
		return 10
	}
}

// Worker represents a worker assigned to a mine
type Worker struct {
	ID           string     `json:"id"`
	Type         WorkerType `json:"type"`
	AssignedAt   time.Time  `json:"assigned_at"`
	AssignedBy   string     `json:"assigned_by"`
	IsActive     bool       `json:"is_active"`
	TotalExtracted int64    `json:"total_extracted"` // Total amount extracted by this worker
}

// NewWorker creates a new worker
func NewWorker(id string, workerType WorkerType, assignedBy string) *Worker {
	return &Worker{
		ID:           id,
		Type:         workerType,
		AssignedAt:   time.Now(),
		AssignedBy:   assignedBy,
		IsActive:     true,
		TotalExtracted: 0,
	}
}

// GetWorkingHours returns the number of hours the worker has been working
func (w *Worker) GetWorkingHours() float64 {
	return time.Since(w.AssignedAt).Hours()
}

// GetTotalCost returns the total cost of the worker
func (w *Worker) GetTotalCost() int64 {
	hours := int64(w.GetWorkingHours())
	return hours * w.Type.GetCost()
}

// Mine represents a mining location owned by a guild
type Mine struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Location    string            `json:"location"`
	Deposits    []*MineralDeposit `json:"deposits"`
	Workers     []*Worker         `json:"workers"`
	DiscoveredAt time.Time        `json:"discovered_at"`
	DiscoveredBy string           `json:"discovered_by"`
	IsActive    bool              `json:"is_active"`
	MaxWorkers  int               `json:"max_workers"`
}

// NewMine creates a new mine
func NewMine(id, name, location, discoveredBy string, deposits []*MineralDeposit) *Mine {
	return &Mine{
		ID:          id,
		Name:        name,
		Location:    location,
		Deposits:    deposits,
		Workers:     make([]*Worker, 0),
		DiscoveredAt: time.Now(),
		DiscoveredBy: discoveredBy,
		IsActive:    true,
		MaxWorkers:  10, // Default max workers
	}
}

// AssignWorker assigns a worker to the mine
func (m *Mine) AssignWorker(worker *Worker) error {
	if !m.IsActive {
		return fmt.Errorf("mine is not active")
	}
	if len(m.Workers) >= m.MaxWorkers {
		return fmt.Errorf("mine has reached maximum worker capacity (%d)", m.MaxWorkers)
	}
	
	// Check if worker is already assigned
	for _, w := range m.Workers {
		if w.ID == worker.ID {
			return fmt.Errorf("worker %s is already assigned to this mine", worker.ID)
		}
	}
	
	m.Workers = append(m.Workers, worker)
	return nil
}

// RemoveWorker removes a worker from the mine
func (m *Mine) RemoveWorker(workerID string) error {
	for i, worker := range m.Workers {
		if worker.ID == workerID {
			worker.IsActive = false
			m.Workers = append(m.Workers[:i], m.Workers[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("worker %s not found in mine", workerID)
}

// GetActiveWorkers returns all active workers
func (m *Mine) GetActiveWorkers() []*Worker {
	activeWorkers := make([]*Worker, 0)
	for _, worker := range m.Workers {
		if worker.IsActive {
			activeWorkers = append(activeWorkers, worker)
		}
	}
	return activeWorkers
}

// GetTotalEfficiency returns the total efficiency of all active workers
func (m *Mine) GetTotalEfficiency() float64 {
	totalEfficiency := 0.0
	for _, worker := range m.GetActiveWorkers() {
		totalEfficiency += worker.Type.GetEfficiency()
	}
	return totalEfficiency
}

// GetTotalValue returns the total value of all deposits in the mine
func (m *Mine) GetTotalValue() int64 {
	totalValue := int64(0)
	for _, deposit := range m.Deposits {
		totalValue += deposit.GetValue()
	}
	return totalValue
}

// GetDepositByType returns a deposit of the specified type
func (m *Mine) GetDepositByType(mineralType MineralType) *MineralDeposit {
	for _, deposit := range m.Deposits {
		if deposit.Type == mineralType {
			return deposit
		}
	}
	return nil
}

// CanExtract checks if the mine can extract minerals
func (m *Mine) CanExtract() bool {
	if !m.IsActive {
		return false
	}
	if len(m.GetActiveWorkers()) == 0 {
		return false
	}
	for _, deposit := range m.Deposits {
		if deposit.GetRemainingAmount() > 0 {
			return true
		}
	}
	return false
}

// Validate validates the mine data
func (m *Mine) Validate() error {
	if m.ID == "" {
		return fmt.Errorf("mine ID cannot be empty")
	}
	if m.Name == "" {
		return fmt.Errorf("mine name cannot be empty")
	}
	if m.Location == "" {
		return fmt.Errorf("mine location cannot be empty")
	}
	if len(m.Deposits) == 0 {
		return fmt.Errorf("mine must have at least one deposit")
	}
	if m.MaxWorkers <= 0 {
		return fmt.Errorf("max workers must be positive")
	}
	return nil
}
