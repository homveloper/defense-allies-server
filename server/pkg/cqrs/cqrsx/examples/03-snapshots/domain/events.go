package domain

import (
	"time"

	"github.com/shopspring/decimal"
)

// OrderCreated 주문 생성 이벤트
type OrderCreated struct {
	OrderID      string          `json:"order_id" bson:"order_id"`
	CustomerID   string          `json:"customer_id" bson:"customer_id"`
	ShippingCost decimal.Decimal `json:"shipping_cost" bson:"shipping_cost"`
	CreatedAt    time.Time       `json:"created_at" bson:"created_at"`
}

func (e *OrderCreated) EventType() string {
	return "OrderCreated"
}

func CreateOrderCreatedEvent(orderID, customerID string, shippingCost decimal.Decimal) *OrderCreated {
	return &OrderCreated{
		OrderID:      orderID,
		CustomerID:   customerID,
		ShippingCost: shippingCost,
		CreatedAt:    time.Now(),
	}
}

// ItemAdded 상품 추가 이벤트
type ItemAdded struct {
	OrderID     string          `json:"order_id" bson:"order_id"`
	ProductID   string          `json:"product_id" bson:"product_id"`
	ProductName string          `json:"product_name" bson:"product_name"`
	Quantity    int             `json:"quantity" bson:"quantity"`
	UnitPrice   decimal.Decimal `json:"unit_price" bson:"unit_price"`
	TotalPrice  decimal.Decimal `json:"total_price" bson:"total_price"`
	AddedAt     time.Time       `json:"added_at" bson:"added_at"`
}

func (e *ItemAdded) EventType() string {
	return "ItemAdded"
}

func CreateItemAddedEvent(orderID, productID, productName string, quantity int, unitPrice decimal.Decimal) *ItemAdded {
	totalPrice := unitPrice.Mul(decimal.NewFromInt(int64(quantity)))
	return &ItemAdded{
		OrderID:     orderID,
		ProductID:   productID,
		ProductName: productName,
		Quantity:    quantity,
		UnitPrice:   unitPrice,
		TotalPrice:  totalPrice,
		AddedAt:     time.Now(),
	}
}

// ItemRemoved 상품 제거 이벤트
type ItemRemoved struct {
	OrderID   string    `json:"order_id" bson:"order_id"`
	ProductID string    `json:"product_id" bson:"product_id"`
	RemovedAt time.Time `json:"removed_at" bson:"removed_at"`
}

func (e *ItemRemoved) EventType() string {
	return "ItemRemoved"
}

func CreateItemRemovedEvent(orderID, productID string) *ItemRemoved {
	return &ItemRemoved{
		OrderID:   orderID,
		ProductID: productID,
		RemovedAt: time.Now(),
	}
}

// ItemQuantityChanged 상품 수량 변경 이벤트
type ItemQuantityChanged struct {
	OrderID     string          `json:"order_id" bson:"order_id"`
	ProductID   string          `json:"product_id" bson:"product_id"`
	OldQuantity int             `json:"old_quantity" bson:"old_quantity"`
	NewQuantity int             `json:"new_quantity" bson:"new_quantity"`
	UnitPrice   decimal.Decimal `json:"unit_price" bson:"unit_price"`
	NewTotal    decimal.Decimal `json:"new_total" bson:"new_total"`
	ChangedAt   time.Time       `json:"changed_at" bson:"changed_at"`
}

func (e *ItemQuantityChanged) EventType() string {
	return "ItemQuantityChanged"
}

func CreateItemQuantityChangedEvent(orderID, productID string, newQuantity int) *ItemQuantityChanged {
	return &ItemQuantityChanged{
		OrderID:     orderID,
		ProductID:   productID,
		NewQuantity: newQuantity,
		ChangedAt:   time.Now(),
	}
}

// DiscountApplied 할인 적용 이벤트
type DiscountApplied struct {
	OrderID      string          `json:"order_id" bson:"order_id"`
	DiscountRate decimal.Decimal `json:"discount_rate" bson:"discount_rate"`
	Reason       string          `json:"reason" bson:"reason"`
	AppliedAt    time.Time       `json:"applied_at" bson:"applied_at"`
}

func (e *DiscountApplied) EventType() string {
	return "DiscountApplied"
}

func CreateDiscountAppliedEvent(orderID string, discountRate decimal.Decimal, reason string) *DiscountApplied {
	return &DiscountApplied{
		OrderID:      orderID,
		DiscountRate: discountRate,
		Reason:       reason,
		AppliedAt:    time.Now(),
	}
}

// OrderConfirmed 주문 확정 이벤트
type OrderConfirmed struct {
	OrderID     string          `json:"order_id" bson:"order_id"`
	TotalAmount decimal.Decimal `json:"total_amount" bson:"total_amount"`
	FinalAmount decimal.Decimal `json:"final_amount" bson:"final_amount"`
	ConfirmedAt time.Time       `json:"confirmed_at" bson:"confirmed_at"`
}

func (e *OrderConfirmed) EventType() string {
	return "OrderConfirmed"
}

func CreateOrderConfirmedEvent(orderID string) *OrderConfirmed {
	return &OrderConfirmed{
		OrderID:     orderID,
		ConfirmedAt: time.Now(),
	}
}

// OrderShipped 주문 배송 시작 이벤트
type OrderShipped struct {
	OrderID        string    `json:"order_id" bson:"order_id"`
	TrackingNumber string    `json:"tracking_number" bson:"tracking_number"`
	ShippedAt      time.Time `json:"shipped_at" bson:"shipped_at"`
}

func (e *OrderShipped) EventType() string {
	return "OrderShipped"
}

func CreateOrderShippedEvent(orderID, trackingNumber string) *OrderShipped {
	return &OrderShipped{
		OrderID:        orderID,
		TrackingNumber: trackingNumber,
		ShippedAt:      time.Now(),
	}
}

// OrderDelivered 주문 배송 완료 이벤트
type OrderDelivered struct {
	OrderID     string    `json:"order_id" bson:"order_id"`
	DeliveredAt time.Time `json:"delivered_at" bson:"delivered_at"`
}

func (e *OrderDelivered) EventType() string {
	return "OrderDelivered"
}

func CreateOrderDeliveredEvent(orderID string) *OrderDelivered {
	return &OrderDelivered{
		OrderID:     orderID,
		DeliveredAt: time.Now(),
	}
}

// OrderCancelled 주문 취소 이벤트
type OrderCancelled struct {
	OrderID     string    `json:"order_id" bson:"order_id"`
	Reason      string    `json:"reason" bson:"reason"`
	CancelledAt time.Time `json:"cancelled_at" bson:"cancelled_at"`
}

func (e *OrderCancelled) EventType() string {
	return "OrderCancelled"
}

func CreateOrderCancelledEvent(orderID, reason string) *OrderCancelled {
	return &OrderCancelled{
		OrderID:     orderID,
		Reason:      reason,
		CancelledAt: time.Now(),
	}
}

// TaxRateChanged 세율 변경 이벤트 (추가 이벤트)
type TaxRateChanged struct {
	OrderID    string          `json:"order_id" bson:"order_id"`
	OldTaxRate decimal.Decimal `json:"old_tax_rate" bson:"old_tax_rate"`
	NewTaxRate decimal.Decimal `json:"new_tax_rate" bson:"new_tax_rate"`
	ChangedAt  time.Time       `json:"changed_at" bson:"changed_at"`
}

func (e *TaxRateChanged) EventType() string {
	return "TaxRateChanged"
}

func CreateTaxRateChangedEvent(orderID string, oldTaxRate, newTaxRate decimal.Decimal) *TaxRateChanged {
	return &TaxRateChanged{
		OrderID:    orderID,
		OldTaxRate: oldTaxRate,
		NewTaxRate: newTaxRate,
		ChangedAt:  time.Now(),
	}
}

// ShippingCostChanged 배송비 변경 이벤트 (추가 이벤트)
type ShippingCostChanged struct {
	OrderID         string          `json:"order_id" bson:"order_id"`
	OldShippingCost decimal.Decimal `json:"old_shipping_cost" bson:"old_shipping_cost"`
	NewShippingCost decimal.Decimal `json:"new_shipping_cost" bson:"new_shipping_cost"`
	Reason          string          `json:"reason" bson:"reason"`
	ChangedAt       time.Time       `json:"changed_at" bson:"changed_at"`
}

func (e *ShippingCostChanged) EventType() string {
	return "ShippingCostChanged"
}

func CreateShippingCostChangedEvent(orderID string, oldCost, newCost decimal.Decimal, reason string) *ShippingCostChanged {
	return &ShippingCostChanged{
		OrderID:         orderID,
		OldShippingCost: oldCost,
		NewShippingCost: newCost,
		Reason:          reason,
		ChangedAt:       time.Now(),
	}
}

// MetadataUpdated 메타데이터 업데이트 이벤트 (추가 이벤트)
type MetadataUpdated struct {
	OrderID   string                 `json:"order_id" bson:"order_id"`
	Key       string                 `json:"key" bson:"key"`
	Value     interface{}            `json:"value" bson:"value"`
	OldValue  interface{}            `json:"old_value,omitempty" bson:"old_value,omitempty"`
	UpdatedAt time.Time              `json:"updated_at" bson:"updated_at"`
}

func (e *MetadataUpdated) EventType() string {
	return "MetadataUpdated"
}

func CreateMetadataUpdatedEvent(orderID, key string, value, oldValue interface{}) *MetadataUpdated {
	return &MetadataUpdated{
		OrderID:   orderID,
		Key:       key,
		Value:     value,
		OldValue:  oldValue,
		UpdatedAt: time.Now(),
	}
}
