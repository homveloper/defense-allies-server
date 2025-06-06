package domain

import (
	"context"
	"defense-allies-server/pkg/cqrs"
	"fmt"
	"time"

	"github.com/shopspring/decimal"
)

// Product represents a product aggregate in the system
type Product struct {
	*cqrs.BaseAggregate
	name        string
	price       decimal.Decimal
	category    string
	description string
	isActive    bool
	createdAt   time.Time
	updatedAt   time.Time
}

// NewProduct creates a new Product aggregate
func NewProduct(id, name string, price decimal.Decimal, category, description string) *Product {
	product := &Product{
		BaseAggregate: cqrs.NewBaseAggregate(id, "Product"),
		name:          name,
		price:         price,
		category:      category,
		description:   description,
		isActive:      true,
		createdAt:     time.Now(),
		updatedAt:     time.Now(),
	}

	// Apply creation event
	event := NewProductCreated(id, name, price, category, description)
	product.ApplyEvent(event)

	return product
}

// LoadProductFromHistory loads a Product aggregate from event history
func LoadProductFromHistory(id string, events []cqrs.EventMessage) (*Product, error) {
	product := &Product{
		BaseAggregate: cqrs.NewBaseAggregate(id, "Product"),
		price:         decimal.Zero,
	}

	for _, event := range events {
		if err := product.applyDomainEvent(event); err != nil {
			return nil, fmt.Errorf("failed to apply event: %w", err)
		}
	}

	return product, nil
}

// Business Methods

// UpdateProduct updates product information
func (p *Product) UpdateProduct(name string, price decimal.Decimal, category, description string) error {
	if name == "" {
		return fmt.Errorf("product name cannot be empty")
	}
	if price.IsNegative() {
		return fmt.Errorf("product price cannot be negative")
	}
	if category == "" {
		return fmt.Errorf("product category cannot be empty")
	}

	event := NewProductUpdated(p.ID(), name, price, category, description)
	p.ApplyEvent(event)

	return nil
}

// Activate activates the product
func (p *Product) Activate() {
	if !p.isActive {
		p.isActive = true
		p.updatedAt = time.Now()
	}
}

// Deactivate deactivates the product
func (p *Product) Deactivate() {
	if p.isActive {
		p.isActive = false
		p.updatedAt = time.Now()
	}
}

// Getters

func (p *Product) GetName() string {
	return p.name
}

func (p *Product) GetPrice() decimal.Decimal {
	return p.price
}

func (p *Product) GetCategory() string {
	return p.category
}

func (p *Product) GetDescription() string {
	return p.description
}

func (p *Product) IsActive() bool {
	return p.isActive
}

func (p *Product) GetCreatedAt() time.Time {
	return p.createdAt
}

func (p *Product) GetUpdatedAt() time.Time {
	return p.updatedAt
}

// Event Application

// applyDomainEvent applies domain events to the aggregate
func (p *Product) applyDomainEvent(event cqrs.EventMessage) error {
	switch e := event.EventData().(type) {
	case *ProductCreated:
		return p.applyProductCreated(e)
	case *ProductUpdated:
		return p.applyProductUpdated(e)
	default:
		// Ignore unknown events
		return nil
	}
}

// applyProductCreated applies ProductCreated event
func (p *Product) applyProductCreated(event *ProductCreated) error {
	p.name = event.Name
	p.price = event.Price
	p.category = event.Category
	p.description = event.Description
	p.isActive = true
	p.createdAt = event.Timestamp()
	p.updatedAt = event.Timestamp()
	return nil
}

// applyProductUpdated applies ProductUpdated event
func (p *Product) applyProductUpdated(event *ProductUpdated) error {
	p.name = event.Name
	p.price = event.Price
	p.category = event.Category
	p.description = event.Description
	p.updatedAt = event.Timestamp()
	return nil
}

// Validation

// Validate validates the product aggregate state
func (p *Product) Validate() error {
	if p.ID() == "" {
		return fmt.Errorf("product ID cannot be empty")
	}
	if p.name == "" {
		return fmt.Errorf("product name cannot be empty")
	}
	if p.price.IsNegative() {
		return fmt.Errorf("product price cannot be negative")
	}
	if p.category == "" {
		return fmt.Errorf("product category cannot be empty")
	}
	return nil
}

// Repository Interface

// ProductRepository defines the interface for product persistence
type ProductRepository interface {
	cqrs.EventSourcedRepository

	// FindByCategory finds products by category
	FindByCategory(ctx context.Context, category string) ([]*Product, error)

	// FindActiveProducts finds all active products
	FindActiveProducts(ctx context.Context) ([]*Product, error)

	// GetProductStats gets product statistics
	GetProductStats(ctx context.Context, productID string) (*ProductStats, error)
}

// ProductStats represents product statistics
type ProductStats struct {
	ProductID    string          `json:"product_id"`
	Name         string          `json:"name"`
	TotalSold    int             `json:"total_sold"`
	TotalRevenue decimal.Decimal `json:"total_revenue"`
	LastSoldAt   *time.Time      `json:"last_sold_at,omitempty"`
}

// Commands

// CreateProductCommand represents a command to create a product
type CreateProductCommand struct {
	*cqrs.BaseCommand
	Name        string          `json:"name"`
	Price       decimal.Decimal `json:"price"`
	Category    string          `json:"category"`
	Description string          `json:"description"`
}

// NewCreateProductCommand creates a new CreateProductCommand
func NewCreateProductCommand(productID, name string, price decimal.Decimal, category, description string) *CreateProductCommand {
	return &CreateProductCommand{
		BaseCommand: cqrs.NewBaseCommand("CreateProduct", productID, "Product", nil),
		Name:        name,
		Price:       price,
		Category:    category,
		Description: description,
	}
}

// UpdateProductCommand represents a command to update a product
type UpdateProductCommand struct {
	*cqrs.BaseCommand
	Name        string          `json:"name"`
	Price       decimal.Decimal `json:"price"`
	Category    string          `json:"category"`
	Description string          `json:"description"`
}

// NewUpdateProductCommand creates a new UpdateProductCommand
func NewUpdateProductCommand(productID, name string, price decimal.Decimal, category, description string) *UpdateProductCommand {
	return &UpdateProductCommand{
		BaseCommand: cqrs.NewBaseCommand("UpdateProduct", productID, "Product", nil),
		Name:        name,
		Price:       price,
		Category:    category,
		Description: description,
	}
}

// Command Handlers

// ProductCommandHandler handles product-related commands
type ProductCommandHandler struct {
	repository ProductRepository
}

// NewProductCommandHandler creates a new ProductCommandHandler
func NewProductCommandHandler(repository ProductRepository) *ProductCommandHandler {
	return &ProductCommandHandler{
		repository: repository,
	}
}

// Handle handles product commands
func (h *ProductCommandHandler) Handle(ctx context.Context, command cqrs.Command) (interface{}, error) {
	switch cmd := command.(type) {
	case *CreateProductCommand:
		return h.handleCreateProduct(ctx, cmd)
	case *UpdateProductCommand:
		return h.handleUpdateProduct(ctx, cmd)
	default:
		return nil, fmt.Errorf("unknown command type: %T", command)
	}
}

// handleCreateProduct handles CreateProductCommand
func (h *ProductCommandHandler) handleCreateProduct(ctx context.Context, cmd *CreateProductCommand) (*Product, error) {
	// Create new product
	product := NewProduct(cmd.ID(), cmd.Name, cmd.Price, cmd.Category, cmd.Description)

	// Validate
	if err := product.Validate(); err != nil {
		return nil, fmt.Errorf("product validation failed: %w", err)
	}

	// Save if repository is available (버전 관리 자동화)
	if h.repository != nil {
		if err := h.repository.Save(ctx, product, 0); err != nil {
			return nil, fmt.Errorf("failed to save product: %w", err)
		}
	}

	return product, nil
}

// handleUpdateProduct handles UpdateProductCommand
func (h *ProductCommandHandler) handleUpdateProduct(ctx context.Context, cmd *UpdateProductCommand) (*Product, error) {
	// Check if repository is available
	if h.repository == nil {
		return nil, fmt.Errorf("repository not available")
	}

	// Load product
	product, err := h.repository.GetByID(ctx, cmd.ID())
	if err != nil {
		return nil, fmt.Errorf("failed to load product: %w", err)
	}

	productAggregate, ok := product.(*Product)
	if !ok {
		return nil, fmt.Errorf("invalid aggregate type")
	}

	// Update product
	if err := productAggregate.UpdateProduct(cmd.Name, cmd.Price, cmd.Category, cmd.Description); err != nil {
		return nil, fmt.Errorf("failed to update product: %w", err)
	}

	// Save (버전 관리 자동화)
	if err := h.repository.Save(ctx, productAggregate, 0); err != nil {
		return nil, fmt.Errorf("failed to save product: %w", err)
	}

	return productAggregate, nil
}
