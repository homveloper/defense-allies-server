package domain

import (
	"fmt"
	"time"
)

// Note: MineralType is already defined in mine.go

// MiningNode represents a mining location
type MiningNode struct {
	NodeID        string      `json:"node_id"`
	Name          string      `json:"name"`
	MineralType   MineralType `json:"mineral_type"`
	Capacity      int         `json:"capacity"`   // Maximum workers
	Difficulty    int         `json:"difficulty"` // Mining difficulty (1-10)
	YieldRate     float64     `json:"yield_rate"` // Minerals per hour per worker
	IsActive      bool        `json:"is_active"`
	RequiredLevel int         `json:"required_level"` // Minimum guild level to access
}

// MiningWorker represents a guild member working in a mine
type MiningWorker struct {
	UserID     string    `json:"user_id"`
	Username   string    `json:"username"`
	AssignedAt time.Time `json:"assigned_at"`
	Efficiency float64   `json:"efficiency"` // Worker efficiency multiplier (0.5 - 2.0)
	Experience int64     `json:"experience"` // Mining experience
	Level      int       `json:"level"`      // Worker mining level
}

// GetEfficiencyMultiplier calculates the efficiency based on worker level and experience
func (w *MiningWorker) GetEfficiencyMultiplier() float64 {
	baseEfficiency := 0.5 + (float64(w.Level) * 0.1)         // 0.5 + (level * 0.1)
	experienceBonus := float64(w.Experience) / 10000.0 * 0.5 // Experience bonus up to 0.5

	totalEfficiency := baseEfficiency + experienceBonus
	if totalEfficiency > 2.0 {
		totalEfficiency = 2.0 // Cap at 2.0x efficiency
	}

	return totalEfficiency
}

// MiningOperation represents an active mining operation
type MiningOperation struct {
	OperationID   string                   `json:"operation_id"`
	NodeID        string                   `json:"node_id"`
	Workers       map[string]*MiningWorker `json:"workers"` // userID -> MiningWorker
	StartedAt     time.Time                `json:"started_at"`
	LastHarvestAt time.Time                `json:"last_harvest_at"`
	TotalYield    map[MineralType]int64    `json:"total_yield"` // Total minerals mined
	Status        string                   `json:"status"`      // Active, Paused, Completed
}

// GetActiveWorkerCount returns the number of active workers
func (op *MiningOperation) GetActiveWorkerCount() int {
	return len(op.Workers)
}

// CalculateYield calculates the mineral yield for a given time period
func (op *MiningOperation) CalculateYield(node *MiningNode, duration time.Duration) int64 {
	if len(op.Workers) == 0 {
		return 0
	}

	hours := duration.Hours()
	totalEfficiency := 0.0

	for _, worker := range op.Workers {
		totalEfficiency += worker.GetEfficiencyMultiplier()
	}

	yield := hours * node.YieldRate * totalEfficiency
	return int64(yield)
}

// GuildMining represents the mining state of a guild
type GuildMining struct {
	GuildID          string                      `json:"guild_id"`
	AvailableNodes   map[string]*MiningNode      `json:"available_nodes"`   // nodeID -> MiningNode
	ActiveOperations map[string]*MiningOperation `json:"active_operations"` // operationID -> MiningOperation
	MineralInventory map[MineralType]int64       `json:"mineral_inventory"` // Total minerals in storage
	TotalProduction  map[MineralType]int64       `json:"total_production"`  // Lifetime production
	MiningLevel      int                         `json:"mining_level"`      // Guild mining level
	MiningExperience int64                       `json:"mining_experience"` // Guild mining experience
	LastUpdatedAt    time.Time                   `json:"last_updated_at"`
}

// NewGuildMining creates a new GuildMining instance
func NewGuildMining(guildID string) *GuildMining {
	return &GuildMining{
		GuildID:          guildID,
		AvailableNodes:   make(map[string]*MiningNode),
		ActiveOperations: make(map[string]*MiningOperation),
		MineralInventory: make(map[MineralType]int64),
		TotalProduction:  make(map[MineralType]int64),
		MiningLevel:      1,
		MiningExperience: 0,
		LastUpdatedAt:    time.Now(),
	}
}

// AddMiningNode adds a new mining node
func (gm *GuildMining) AddMiningNode(node *MiningNode) error {
	if node.NodeID == "" {
		return fmt.Errorf("node ID cannot be empty")
	}

	if _, exists := gm.AvailableNodes[node.NodeID]; exists {
		return fmt.Errorf("mining node %s already exists", node.NodeID)
	}

	gm.AvailableNodes[node.NodeID] = node
	return nil
}

// StartMiningOperation starts a new mining operation
func (gm *GuildMining) StartMiningOperation(operationID, nodeID string, workers []*MiningWorker) error {
	// Validate node exists and is active
	node, exists := gm.AvailableNodes[nodeID]
	if !exists {
		return fmt.Errorf("mining node %s not found", nodeID)
	}

	if !node.IsActive {
		return fmt.Errorf("mining node %s is not active", nodeID)
	}

	// Check capacity
	if len(workers) > node.Capacity {
		return fmt.Errorf("too many workers: %d exceeds node capacity %d", len(workers), node.Capacity)
	}

	// Check if operation already exists
	if _, exists := gm.ActiveOperations[operationID]; exists {
		return fmt.Errorf("mining operation %s already exists", operationID)
	}

	// Create operation
	operation := &MiningOperation{
		OperationID:   operationID,
		NodeID:        nodeID,
		Workers:       make(map[string]*MiningWorker),
		StartedAt:     time.Now(),
		LastHarvestAt: time.Now(),
		TotalYield:    make(map[MineralType]int64),
		Status:        "Active",
	}

	// Add workers
	for _, worker := range workers {
		operation.Workers[worker.UserID] = worker
	}

	gm.ActiveOperations[operationID] = operation
	gm.LastUpdatedAt = time.Now()

	return nil
}

// HarvestMinerals harvests minerals from an active operation
func (gm *GuildMining) HarvestMinerals(operationID string) (map[MineralType]int64, error) {
	operation, exists := gm.ActiveOperations[operationID]
	if !exists {
		return nil, fmt.Errorf("mining operation %s not found", operationID)
	}

	if operation.Status != "Active" {
		return nil, fmt.Errorf("mining operation %s is not active", operationID)
	}

	node, exists := gm.AvailableNodes[operation.NodeID]
	if !exists {
		return nil, fmt.Errorf("mining node %s not found", operation.NodeID)
	}

	// Calculate yield since last harvest
	duration := time.Since(operation.LastHarvestAt)
	yield := operation.CalculateYield(node, duration)

	if yield <= 0 {
		return map[MineralType]int64{}, nil
	}

	// Add to inventories
	harvested := map[MineralType]int64{
		node.MineralType: yield,
	}

	gm.MineralInventory[node.MineralType] += yield
	gm.TotalProduction[node.MineralType] += yield
	operation.TotalYield[node.MineralType] += yield

	// Update timestamps
	operation.LastHarvestAt = time.Now()
	gm.LastUpdatedAt = time.Now()

	// Add mining experience
	gm.MiningExperience += yield / 10 // 1 exp per 10 minerals

	// Check for level up
	requiredExp := int64(gm.MiningLevel * 1000)
	if gm.MiningExperience >= requiredExp {
		gm.MiningLevel++
		gm.MiningExperience -= requiredExp
	}

	return harvested, nil
}

// StopMiningOperation stops an active mining operation
func (gm *GuildMining) StopMiningOperation(operationID string) error {
	operation, exists := gm.ActiveOperations[operationID]
	if !exists {
		return fmt.Errorf("mining operation %s not found", operationID)
	}

	operation.Status = "Completed"
	gm.LastUpdatedAt = time.Now()

	return nil
}

// GetTotalMineralValue calculates the total value of all minerals in inventory
func (gm *GuildMining) GetTotalMineralValue() int64 {
	totalValue := int64(0)
	for mineralType, amount := range gm.MineralInventory {
		totalValue += amount * mineralType.GetValue()
	}
	return totalValue
}

// GetActiveOperationsCount returns the number of active mining operations
func (gm *GuildMining) GetActiveOperationsCount() int {
	count := 0
	for _, operation := range gm.ActiveOperations {
		if operation.Status == "Active" {
			count++
		}
	}
	return count
}
