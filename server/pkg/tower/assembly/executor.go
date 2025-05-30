// Package assembly executor provides execution logic for component assemblies
package assembly

import (
	"context"
	"fmt"
	"sync"
	"time"

	"defense-allies-server/pkg/tower/component"
)

// ExecutorConfig provides configuration for the assembly executor
type ExecutorConfig struct {
	MaxConcurrency    int           `json:"max_concurrency"`
	ExecutionTimeout  time.Duration `json:"execution_timeout"`
	EnablePipelining  bool          `json:"enable_pipelining"`
	EnableParallelism bool          `json:"enable_parallelism"`
	BufferSize        int           `json:"buffer_size"`
}

// DefaultExecutorConfig returns a default executor configuration
func DefaultExecutorConfig() ExecutorConfig {
	return ExecutorConfig{
		MaxConcurrency:    10,
		ExecutionTimeout:  time.Second * 5,
		EnablePipelining:  true,
		EnableParallelism: true,
		BufferSize:        100,
	}
}

// RealAssemblyExecutor provides the actual implementation of assembly execution
type RealAssemblyExecutor struct {
	config ExecutorConfig
	mutex  sync.RWMutex
}

// NewRealAssemblyExecutor creates a new real assembly executor
func NewRealAssemblyExecutor(config ExecutorConfig) *RealAssemblyExecutor {
	return &RealAssemblyExecutor{
		config: config,
	}
}

// ExecuteAssembly executes a component assembly with the given execution context
func (rae *RealAssemblyExecutor) ExecuteAssembly(ctx context.Context, assembly *ComponentAssembly, execCtx *component.ExecutionContext) (*AssemblyResult, error) {
	startTime := time.Now()
	
	// Create execution context with timeout
	execCtxWithTimeout, cancel := context.WithTimeout(ctx, rae.config.ExecutionTimeout)
	defer cancel()
	
	// Validate assembly before execution
	if !assembly.IsValid {
		return nil, fmt.Errorf("assembly is not valid")
	}
	
	if len(assembly.ExecutionOrder) == 0 {
		return nil, fmt.Errorf("assembly has no execution order")
	}
	
	// Initialize execution state
	executionState := &ExecutionState{
		Assembly:         assembly,
		Context:          execCtx,
		ComponentResults: make(map[string]*component.ComponentResult),
		DataFlow:         make(map[string]map[string]interface{}),
		Errors:           []string{},
		Warnings:         []string{},
	}
	
	// Execute components in topological order
	var finalResult *AssemblyResult
	var err error
	
	if rae.config.EnableParallelism {
		finalResult, err = rae.executeParallel(execCtxWithTimeout, executionState)
	} else {
		finalResult, err = rae.executeSequential(execCtxWithTimeout, executionState)
	}
	
	if err != nil {
		return nil, err
	}
	
	// Set execution time
	finalResult.ExecutionTime = time.Since(startTime)
	
	return finalResult, nil
}

// ExecutionState maintains state during assembly execution
type ExecutionState struct {
	Assembly         *ComponentAssembly
	Context          *component.ExecutionContext
	ComponentResults map[string]*component.ComponentResult
	DataFlow         map[string]map[string]interface{}
	Errors           []string
	Warnings         []string
	mutex            sync.RWMutex
}

// executeSequential executes components sequentially in topological order
func (rae *RealAssemblyExecutor) executeSequential(ctx context.Context, state *ExecutionState) (*AssemblyResult, error) {
	for _, componentID := range state.Assembly.ExecutionOrder {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
			if err := rae.executeComponent(ctx, state, componentID); err != nil {
				return nil, fmt.Errorf("failed to execute component %s: %w", componentID, err)
			}
		}
	}
	
	return rae.buildFinalResult(state), nil
}

// executeParallel executes components in parallel where possible
func (rae *RealAssemblyExecutor) executeParallel(ctx context.Context, state *ExecutionState) (*AssemblyResult, error) {
	// Build dependency levels for parallel execution
	levels := rae.buildDependencyLevels(state.Assembly)
	
	// Execute each level in parallel
	for _, level := range levels {
		if err := rae.executeLevel(ctx, state, level); err != nil {
			return nil, err
		}
	}
	
	return rae.buildFinalResult(state), nil
}

// buildDependencyLevels builds levels of components that can be executed in parallel
func (rae *RealAssemblyExecutor) buildDependencyLevels(assembly *ComponentAssembly) [][]string {
	// Build dependency graph
	dependencies := make(map[string][]string)
	dependents := make(map[string][]string)
	
	for componentID := range assembly.Components {
		dependencies[componentID] = []string{}
		dependents[componentID] = []string{}
	}
	
	for _, conn := range assembly.Connections {
		dependencies[conn.ToComponent] = append(dependencies[conn.ToComponent], conn.FromComponent)
		dependents[conn.FromComponent] = append(dependents[conn.FromComponent], conn.ToComponent)
	}
	
	// Build levels using modified Kahn's algorithm
	var levels [][]string
	remaining := make(map[string]bool)
	for componentID := range assembly.Components {
		remaining[componentID] = true
	}
	
	for len(remaining) > 0 {
		var currentLevel []string
		
		// Find components with no remaining dependencies
		for componentID := range remaining {
			canExecute := true
			for _, dep := range dependencies[componentID] {
				if remaining[dep] {
					canExecute = false
					break
				}
			}
			if canExecute {
				currentLevel = append(currentLevel, componentID)
			}
		}
		
		if len(currentLevel) == 0 {
			// This shouldn't happen if the assembly is valid
			break
		}
		
		// Remove current level components from remaining
		for _, componentID := range currentLevel {
			delete(remaining, componentID)
		}
		
		levels = append(levels, currentLevel)
	}
	
	return levels
}

// executeLevel executes all components in a level in parallel
func (rae *RealAssemblyExecutor) executeLevel(ctx context.Context, state *ExecutionState, level []string) error {
	if len(level) == 1 {
		// Single component, execute directly
		return rae.executeComponent(ctx, state, level[0])
	}
	
	// Multiple components, execute in parallel
	errChan := make(chan error, len(level))
	var wg sync.WaitGroup
	
	for _, componentID := range level {
		wg.Add(1)
		go func(compID string) {
			defer wg.Done()
			if err := rae.executeComponent(ctx, state, compID); err != nil {
				errChan <- fmt.Errorf("component %s: %w", compID, err)
			}
		}(componentID)
	}
	
	wg.Wait()
	close(errChan)
	
	// Check for errors
	for err := range errChan {
		if err != nil {
			return err
		}
	}
	
	return nil
}

// executeComponent executes a single component
func (rae *RealAssemblyExecutor) executeComponent(ctx context.Context, state *ExecutionState, componentID string) error {
	comp, exists := state.Assembly.Components[componentID]
	if !exists {
		return fmt.Errorf("component %s not found", componentID)
	}
	
	// Prepare input data for this component
	inputData := rae.prepareInputData(state, componentID)
	
	// Create execution context for this component
	compExecCtx := &component.ExecutionContext{
		GameTime:     state.Context.GameTime,
		DeltaTime:    state.Context.DeltaTime,
		GameState:    state.Context.GameState,
		TowerID:      state.Context.TowerID,
		TowerPos:     state.Context.TowerPos,
		OwnerID:      state.Context.OwnerID,
		InputData:    inputData,
		Environment:  state.Context.Environment,
		PowerMatrix:  state.Context.PowerMatrix,
		ExecutionID:  state.Context.ExecutionID,
		TraceID:      state.Context.TraceID,
	}
	
	// Execute the component
	result, err := comp.Execute(ctx, compExecCtx)
	if err != nil {
		return fmt.Errorf("component execution failed: %w", err)
	}
	
	// Store the result
	state.mutex.Lock()
	state.ComponentResults[componentID] = result
	
	// Store output data for downstream components
	if state.DataFlow[componentID] == nil {
		state.DataFlow[componentID] = make(map[string]interface{})
	}
	for key, value := range result.Outputs {
		state.DataFlow[componentID][key] = value
	}
	state.mutex.Unlock()
	
	return nil
}

// prepareInputData prepares input data for a component based on its connections
func (rae *RealAssemblyExecutor) prepareInputData(state *ExecutionState, componentID string) map[string]interface{} {
	inputData := make(map[string]interface{})
	
	// Copy initial context data
	for key, value := range state.Context.InputData {
		inputData[key] = value
	}
	
	// Add data from connected components
	for _, conn := range state.Assembly.Connections {
		if conn.ToComponent == componentID {
			state.mutex.RLock()
			if fromData, exists := state.DataFlow[conn.FromComponent]; exists {
				if outputValue, exists := fromData[conn.FromOutput]; exists {
					inputData[conn.ToInput] = outputValue
				}
			}
			state.mutex.RUnlock()
		}
	}
	
	return inputData
}

// buildFinalResult builds the final assembly result from component results
func (rae *RealAssemblyExecutor) buildFinalResult(state *ExecutionState) *AssemblyResult {
	result := &AssemblyResult{
		Success:          true,
		Outputs:          make(map[string]interface{}),
		Effects:          []component.Effect{},
		Events:           []component.GameEvent{},
		Errors:           state.Errors,
		Warnings:         state.Warnings,
		ComponentResults: state.ComponentResults,
		Metadata:         make(map[string]interface{}),
	}
	
	// Collect outputs from exit point components
	for _, exitPointID := range state.Assembly.ExitPoints {
		if compResult, exists := state.ComponentResults[exitPointID]; exists {
			for key, value := range compResult.Outputs {
				result.Outputs[key] = value
			}
		}
	}
	
	// Collect all effects and events
	for _, compResult := range state.ComponentResults {
		if !compResult.Success {
			result.Success = false
			if compResult.Error != "" {
				result.Errors = append(result.Errors, compResult.Error)
			}
		}
		
		result.Effects = append(result.Effects, compResult.Effects...)
		result.Events = append(result.Events, compResult.Events...)
	}
	
	// Add execution metadata
	result.Metadata["component_count"] = len(state.ComponentResults)
	result.Metadata["connection_count"] = len(state.Assembly.Connections)
	result.Metadata["execution_order"] = state.Assembly.ExecutionOrder
	
	return result
}

// GetExecutorStats returns statistics about the executor
func (rae *RealAssemblyExecutor) GetExecutorStats() ExecutorStats {
	return ExecutorStats{
		MaxConcurrency:    rae.config.MaxConcurrency,
		ExecutionTimeout:  rae.config.ExecutionTimeout,
		EnablePipelining:  rae.config.EnablePipelining,
		EnableParallelism: rae.config.EnableParallelism,
		BufferSize:        rae.config.BufferSize,
	}
}

// ExecutorStats provides statistics about the executor
type ExecutorStats struct {
	MaxConcurrency    int           `json:"max_concurrency"`
	ExecutionTimeout  time.Duration `json:"execution_timeout"`
	EnablePipelining  bool          `json:"enable_pipelining"`
	EnableParallelism bool          `json:"enable_parallelism"`
	BufferSize        int           `json:"buffer_size"`
}
