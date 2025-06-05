package domain

import (
	"errors"
	"fmt"
	"time"

	"github.com/shopspring/decimal"
	"go.mongodb.org/mongo-driver/bson"

	"defense-allies-server/pkg/cqrs"
)

// OrderStatus 주문 상태
type OrderStatus string

const (
	OrderStatusPending   OrderStatus = "pending"
	OrderStatusConfirmed OrderStatus = "confirmed"
	OrderStatusShipped   OrderStatus = "shipped"
	OrderStatusDelivered OrderStatus = "delivered"
	OrderStatusCancelled OrderStatus = "cancelled"
)

// OrderItem 주문 상품
type OrderItem struct {
	ProductID   string          `json:"product_id" bson:"product_id"`
	ProductName string          `json:"product_name" bson:"product_name"`
	Quantity    int             `json:"quantity" bson:"quantity"`
	UnitPrice   decimal.Decimal `json:"unit_price" bson:"unit_price"`
	TotalPrice  decimal.Decimal `json:"total_price" bson:"total_price"`
}

// Order 주문 Aggregate
type Order struct {
	// BaseAggregate 필드들
	AggregateID_     string              `json:"aggregate_id" bson:"aggregate_id"`
	AggregateType_   string              `json:"aggregate_type" bson:"aggregate_type"`
	Version_         int                 `json:"version" bson:"version"`
	OriginalVersion_ int                 `json:"original_version" bson:"original_version"`
	Changes_         []cqrs.EventMessage `json:"-" bson:"-"` // 직렬화하지 않음

	// 비즈니스 필드들
	CustomerID   string                 `json:"customer_id" bson:"customer_id"`
	Items        []OrderItem            `json:"items" bson:"items"`
	Status       OrderStatus            `json:"status" bson:"status"`
	TotalAmount  decimal.Decimal        `json:"total_amount" bson:"total_amount"`
	DiscountRate decimal.Decimal        `json:"discount_rate" bson:"discount_rate"`
	TaxRate      decimal.Decimal        `json:"tax_rate" bson:"tax_rate"`
	ShippingCost decimal.Decimal        `json:"shipping_cost" bson:"shipping_cost"`
	FinalAmount  decimal.Decimal        `json:"final_amount" bson:"final_amount"`
	CreatedAt    time.Time              `json:"created_at" bson:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at" bson:"updated_at"`
	ConfirmedAt  *time.Time             `json:"confirmed_at,omitempty" bson:"confirmed_at,omitempty"`
	ShippedAt    *time.Time             `json:"shipped_at,omitempty" bson:"shipped_at,omitempty"`
	DeliveredAt  *time.Time             `json:"delivered_at,omitempty" bson:"delivered_at,omitempty"`
	CancelledAt  *time.Time             `json:"cancelled_at,omitempty" bson:"cancelled_at,omitempty"`
	Metadata     map[string]interface{} `json:"metadata" bson:"metadata"`
}

// NewOrder 새로운 주문 생성
func NewOrder() *Order {
	order := &Order{
		BaseAggregate: cqrs.NewBaseAggregate("", "Order"),
		Items:         make([]OrderItem, 0),
		Status:        OrderStatusPending,
		TotalAmount:   decimal.Zero,
		DiscountRate:  decimal.Zero,
		TaxRate:       decimal.NewFromFloat(0.1), // 기본 10% 세율
		ShippingCost:  decimal.Zero,
		FinalAmount:   decimal.Zero,
		Metadata:      make(map[string]interface{}),
	}
	order.syncFromBaseAggregate()
	return order
}

// NewOrderWithID ID를 가진 새로운 주문 생성
func NewOrderWithID(id string) *Order {
	order := NewOrder()
	order.BaseAggregate = cqrs.NewBaseAggregate(id, "Order")
	return order
}

// CreateOrder 주문 생성 - 비즈니스 로직
func (o *Order) CreateOrder(orderID, customerID string, shippingCost decimal.Decimal) error {
	// 비즈니스 규칙 검증
	if err := o.validateOrderCreation(orderID, customerID); err != nil {
		return err
	}

	// 이미 생성된 주문인지 확인
	if o.Version() > 0 {
		return errors.New("order already exists")
	}

	// Aggregate ID 설정
	o.BaseAggregate = cqrs.NewBaseAggregate(orderID, "Order")

	// 이벤트 생성 및 추적
	event := CreateOrderCreatedEvent(orderID, customerID, shippingCost)
	eventMessage := cqrs.NewBaseEventMessage(
		event.EventType(),
		orderID,
		"Order",
		1,
		event,
	)

	o.TrackChange(eventMessage)
	return nil
}

// AddItem 상품 추가
func (o *Order) AddItem(productID, productName string, quantity int, unitPrice decimal.Decimal) error {
	if err := o.validateItemAddition(productID, quantity, unitPrice); err != nil {
		return err
	}

	if o.Status != OrderStatusPending {
		return errors.New("cannot add items to non-pending order")
	}

	// 이벤트 생성
	event := CreateItemAddedEvent(o.ID(), productID, productName, quantity, unitPrice)
	eventMessage := cqrs.NewBaseEventMessage(
		event.EventType(),
		o.ID(),
		"Order",
		o.Version()+1,
		event,
	)

	o.TrackChange(eventMessage)
	return nil
}

// RemoveItem 상품 제거
func (o *Order) RemoveItem(productID string) error {
	if o.Status != OrderStatusPending {
		return errors.New("cannot remove items from non-pending order")
	}

	// 상품이 존재하는지 확인
	found := false
	for _, item := range o.Items {
		if item.ProductID == productID {
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("product %s not found in order", productID)
	}

	// 이벤트 생성
	event := CreateItemRemovedEvent(o.ID(), productID)
	eventMessage := cqrs.NewBaseEventMessage(
		event.EventType(),
		o.ID(),
		"Order",
		o.Version()+1,
		event,
	)

	o.TrackChange(eventMessage)
	return nil
}

// ChangeItemQuantity 상품 수량 변경
func (o *Order) ChangeItemQuantity(productID string, newQuantity int) error {
	if err := o.validateQuantityChange(productID, newQuantity); err != nil {
		return err
	}

	if o.Status != OrderStatusPending {
		return errors.New("cannot change item quantity in non-pending order")
	}

	// 이벤트 생성
	event := CreateItemQuantityChangedEvent(o.ID(), productID, newQuantity)
	eventMessage := cqrs.NewBaseEventMessage(
		event.EventType(),
		o.ID(),
		"Order",
		o.Version()+1,
		event,
	)

	o.TrackChange(eventMessage)
	return nil
}

// ApplyDiscount 할인 적용
func (o *Order) ApplyDiscount(discountRate decimal.Decimal, reason string) error {
	if o.Status != OrderStatusPending {
		return errors.New("cannot apply discount to non-pending order")
	}

	if discountRate.LessThan(decimal.Zero) || discountRate.GreaterThan(decimal.NewFromFloat(1.0)) {
		return errors.New("discount rate must be between 0 and 1")
	}

	// 이벤트 생성
	event := CreateDiscountAppliedEvent(o.ID(), discountRate, reason)
	eventMessage := cqrs.NewBaseEventMessage(
		event.EventType(),
		o.ID(),
		"Order",
		o.Version()+1,
		event,
	)

	o.TrackChange(eventMessage)
	return nil
}

// ConfirmOrder 주문 확정
func (o *Order) ConfirmOrder() error {
	if o.Status != OrderStatusPending {
		return fmt.Errorf("cannot confirm order with status %s", o.Status)
	}

	if len(o.Items) == 0 {
		return errors.New("cannot confirm order with no items")
	}

	// 이벤트 생성
	event := CreateOrderConfirmedEvent(o.ID())
	eventMessage := cqrs.NewBaseEventMessage(
		event.EventType(),
		o.ID(),
		"Order",
		o.Version()+1,
		event,
	)

	o.TrackChange(eventMessage)
	return nil
}

// ShipOrder 주문 배송 시작
func (o *Order) ShipOrder(trackingNumber string) error {
	if o.Status != OrderStatusConfirmed {
		return fmt.Errorf("cannot ship order with status %s", o.Status)
	}

	// 이벤트 생성
	event := CreateOrderShippedEvent(o.ID(), trackingNumber)
	eventMessage := cqrs.NewBaseEventMessage(
		event.EventType(),
		o.ID(),
		"Order",
		o.Version()+1,
		event,
	)

	o.TrackChange(eventMessage)
	return nil
}

// DeliverOrder 주문 배송 완료
func (o *Order) DeliverOrder() error {
	if o.Status != OrderStatusShipped {
		return fmt.Errorf("cannot deliver order with status %s", o.Status)
	}

	// 이벤트 생성
	event := CreateOrderDeliveredEvent(o.ID())
	eventMessage := cqrs.NewBaseEventMessage(
		event.EventType(),
		o.ID(),
		"Order",
		o.Version()+1,
		event,
	)

	o.TrackChange(eventMessage)
	return nil
}

// CancelOrder 주문 취소
func (o *Order) CancelOrder(reason string) error {
	if o.Status == OrderStatusDelivered {
		return errors.New("cannot cancel delivered order")
	}

	if o.Status == OrderStatusCancelled {
		return errors.New("order is already cancelled")
	}

	// 이벤트 생성
	event := CreateOrderCancelledEvent(o.ID(), reason)
	eventMessage := cqrs.NewBaseEventMessage(
		event.EventType(),
		o.ID(),
		"Order",
		o.Version()+1,
		event,
	)

	o.TrackChange(eventMessage)
	return nil
}

// 편의 메서드들
func (o *Order) Version() int {
	return o.CurrentVersion()
}

func (o *Order) ID() string {
	return o.AggregateID()
}

func (o *Order) Type() string {
	return o.AggregateType()
}

func (o *Order) GetUncommittedChanges() []cqrs.EventMessage {
	return o.GetChanges()
}

func (o *Order) ItemCount() int {
	return len(o.Items)
}

func (o *Order) IsEmpty() bool {
	return len(o.Items) == 0
}

func (o *Order) HasItem(productID string) bool {
	for _, item := range o.Items {
		if item.ProductID == productID {
			return true
		}
	}
	return false
}

func (o *Order) GetItem(productID string) (OrderItem, bool) {
	for _, item := range o.Items {
		if item.ProductID == productID {
			return item, true
		}
	}
	return OrderItem{}, false
}

// String 주문 정보를 문자열로 반환
func (o *Order) String() string {
	return fmt.Sprintf("Order{ID: %s, Customer: %s, Status: %s, Items: %d, Total: %s, Version: %d}",
		o.ID(), o.CustomerID, o.Status, len(o.Items), o.FinalAmount.String(), o.Version())
}

// 비즈니스 규칙 검증 메서드들
func (o *Order) validateOrderCreation(orderID, customerID string) error {
	if orderID == "" {
		return errors.New("order ID cannot be empty")
	}
	if customerID == "" {
		return errors.New("customer ID cannot be empty")
	}
	return nil
}

func (o *Order) validateItemAddition(productID string, quantity int, unitPrice decimal.Decimal) error {
	if productID == "" {
		return errors.New("product ID cannot be empty")
	}
	if quantity <= 0 {
		return errors.New("quantity must be positive")
	}
	if unitPrice.LessThan(decimal.Zero) {
		return errors.New("unit price cannot be negative")
	}
	return nil
}

func (o *Order) validateQuantityChange(productID string, newQuantity int) error {
	if productID == "" {
		return errors.New("product ID cannot be empty")
	}
	if newQuantity < 0 {
		return errors.New("quantity cannot be negative")
	}

	// 상품이 존재하는지 확인
	if !o.HasItem(productID) {
		return fmt.Errorf("product %s not found in order", productID)
	}

	return nil
}

// 금액 계산 메서드들
func (o *Order) calculateTotalAmount() decimal.Decimal {
	total := decimal.Zero
	for _, item := range o.Items {
		total = total.Add(item.TotalPrice)
	}
	return total
}

func (o *Order) calculateFinalAmount() decimal.Decimal {
	// 총액 계산
	subtotal := o.calculateTotalAmount()

	// 할인 적용
	discountAmount := subtotal.Mul(o.DiscountRate)
	afterDiscount := subtotal.Sub(discountAmount)

	// 세금 적용
	taxAmount := afterDiscount.Mul(o.TaxRate)
	afterTax := afterDiscount.Add(taxAmount)

	// 배송비 추가
	finalAmount := afterTax.Add(o.ShippingCost)

	return finalAmount
}

// Apply 이벤트를 적용하여 상태 변경
func (o *Order) Apply(event cqrs.EventMessage) error {
	// BaseAggregate의 Apply 메서드 호출 (버전 관리)
	o.BaseAggregate.Apply(event, false)

	// BSON에서 역직렬화된 데이터를 구체적인 이벤트 타입으로 변환
	eventData, err := o.convertEventData(event.EventType(), event.EventData())
	if err != nil {
		return fmt.Errorf("failed to convert event data for %s: %w", event.EventType(), err)
	}

	switch e := eventData.(type) {
	case *OrderCreated:
		return o.applyOrderCreated(e)
	case *ItemAdded:
		return o.applyItemAdded(e)
	case *ItemRemoved:
		return o.applyItemRemoved(e)
	case *ItemQuantityChanged:
		return o.applyItemQuantityChanged(e)
	case *DiscountApplied:
		return o.applyDiscountApplied(e)
	case *OrderConfirmed:
		return o.applyOrderConfirmed(e)
	case *OrderShipped:
		return o.applyOrderShipped(e)
	case *OrderDelivered:
		return o.applyOrderDelivered(e)
	case *OrderCancelled:
		return o.applyOrderCancelled(e)
	case *TaxRateChanged:
		return o.applyTaxRateChanged(e)
	case *ShippingCostChanged:
		return o.applyShippingCostChanged(e)
	case *MetadataUpdated:
		return o.applyMetadataUpdated(e)
	default:
		return fmt.Errorf("unknown event type: %T", e)
	}
}

// convertEventData BSON 데이터를 구체적인 이벤트 타입으로 변환
func (o *Order) convertEventData(eventType string, data interface{}) (interface{}, error) {
	// BSON 바이트로 변환
	bsonData, err := bson.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal event data: %w", err)
	}

	// 이벤트 타입에 따라 구체적인 구조체로 역직렬화
	switch eventType {
	case "OrderCreated":
		var event OrderCreated
		if err := bson.Unmarshal(bsonData, &event); err != nil {
			return nil, fmt.Errorf("failed to unmarshal OrderCreated: %w", err)
		}
		return &event, nil

	case "ItemAdded":
		var event ItemAdded
		if err := bson.Unmarshal(bsonData, &event); err != nil {
			return nil, fmt.Errorf("failed to unmarshal ItemAdded: %w", err)
		}
		return &event, nil

	case "ItemRemoved":
		var event ItemRemoved
		if err := bson.Unmarshal(bsonData, &event); err != nil {
			return nil, fmt.Errorf("failed to unmarshal ItemRemoved: %w", err)
		}
		return &event, nil

	case "ItemQuantityChanged":
		var event ItemQuantityChanged
		if err := bson.Unmarshal(bsonData, &event); err != nil {
			return nil, fmt.Errorf("failed to unmarshal ItemQuantityChanged: %w", err)
		}
		return &event, nil

	case "DiscountApplied":
		var event DiscountApplied
		if err := bson.Unmarshal(bsonData, &event); err != nil {
			return nil, fmt.Errorf("failed to unmarshal DiscountApplied: %w", err)
		}
		return &event, nil

	case "OrderConfirmed":
		var event OrderConfirmed
		if err := bson.Unmarshal(bsonData, &event); err != nil {
			return nil, fmt.Errorf("failed to unmarshal OrderConfirmed: %w", err)
		}
		return &event, nil

	case "OrderShipped":
		var event OrderShipped
		if err := bson.Unmarshal(bsonData, &event); err != nil {
			return nil, fmt.Errorf("failed to unmarshal OrderShipped: %w", err)
		}
		return &event, nil

	case "OrderDelivered":
		var event OrderDelivered
		if err := bson.Unmarshal(bsonData, &event); err != nil {
			return nil, fmt.Errorf("failed to unmarshal OrderDelivered: %w", err)
		}
		return &event, nil

	case "OrderCancelled":
		var event OrderCancelled
		if err := bson.Unmarshal(bsonData, &event); err != nil {
			return nil, fmt.Errorf("failed to unmarshal OrderCancelled: %w", err)
		}
		return &event, nil

	case "TaxRateChanged":
		var event TaxRateChanged
		if err := bson.Unmarshal(bsonData, &event); err != nil {
			return nil, fmt.Errorf("failed to unmarshal TaxRateChanged: %w", err)
		}
		return &event, nil

	case "ShippingCostChanged":
		var event ShippingCostChanged
		if err := bson.Unmarshal(bsonData, &event); err != nil {
			return nil, fmt.Errorf("failed to unmarshal ShippingCostChanged: %w", err)
		}
		return &event, nil

	case "MetadataUpdated":
		var event MetadataUpdated
		if err := bson.Unmarshal(bsonData, &event); err != nil {
			return nil, fmt.Errorf("failed to unmarshal MetadataUpdated: %w", err)
		}
		return &event, nil

	default:
		return nil, fmt.Errorf("unknown event type: %s", eventType)
	}
}

// 이벤트 적용 메서드들
func (o *Order) applyOrderCreated(event *OrderCreated) error {
	o.CustomerID = event.CustomerID
	o.ShippingCost = event.ShippingCost
	o.Status = OrderStatusPending
	o.CreatedAt = event.CreatedAt
	o.UpdatedAt = event.CreatedAt
	return nil
}

func (o *Order) applyItemAdded(event *ItemAdded) error {
	item := OrderItem{
		ProductID:   event.ProductID,
		ProductName: event.ProductName,
		Quantity:    event.Quantity,
		UnitPrice:   event.UnitPrice,
		TotalPrice:  event.TotalPrice,
	}
	o.Items = append(o.Items, item)
	o.TotalAmount = o.calculateTotalAmount()
	o.FinalAmount = o.calculateFinalAmount()
	o.UpdatedAt = event.AddedAt
	return nil
}

func (o *Order) applyItemRemoved(event *ItemRemoved) error {
	for i, item := range o.Items {
		if item.ProductID == event.ProductID {
			o.Items = append(o.Items[:i], o.Items[i+1:]...)
			break
		}
	}
	o.TotalAmount = o.calculateTotalAmount()
	o.FinalAmount = o.calculateFinalAmount()
	o.UpdatedAt = event.RemovedAt
	return nil
}

func (o *Order) applyItemQuantityChanged(event *ItemQuantityChanged) error {
	for i, item := range o.Items {
		if item.ProductID == event.ProductID {
			o.Items[i].Quantity = event.NewQuantity
			o.Items[i].TotalPrice = item.UnitPrice.Mul(decimal.NewFromInt(int64(event.NewQuantity)))
			break
		}
	}
	o.TotalAmount = o.calculateTotalAmount()
	o.FinalAmount = o.calculateFinalAmount()
	o.UpdatedAt = event.ChangedAt
	return nil
}

func (o *Order) applyDiscountApplied(event *DiscountApplied) error {
	o.DiscountRate = event.DiscountRate
	o.FinalAmount = o.calculateFinalAmount()
	o.UpdatedAt = event.AppliedAt
	return nil
}

func (o *Order) applyOrderConfirmed(event *OrderConfirmed) error {
	o.Status = OrderStatusConfirmed
	o.ConfirmedAt = &event.ConfirmedAt
	o.UpdatedAt = event.ConfirmedAt
	return nil
}

func (o *Order) applyOrderShipped(event *OrderShipped) error {
	o.Status = OrderStatusShipped
	o.ShippedAt = &event.ShippedAt
	o.UpdatedAt = event.ShippedAt
	// 메타데이터에 추적 번호 저장
	o.Metadata["tracking_number"] = event.TrackingNumber
	return nil
}

func (o *Order) applyOrderDelivered(event *OrderDelivered) error {
	o.Status = OrderStatusDelivered
	o.DeliveredAt = &event.DeliveredAt
	o.UpdatedAt = event.DeliveredAt
	return nil
}

func (o *Order) applyOrderCancelled(event *OrderCancelled) error {
	o.Status = OrderStatusCancelled
	o.CancelledAt = &event.CancelledAt
	o.UpdatedAt = event.CancelledAt
	// 메타데이터에 취소 사유 저장
	o.Metadata["cancellation_reason"] = event.Reason
	return nil
}

func (o *Order) applyTaxRateChanged(event *TaxRateChanged) error {
	o.TaxRate = event.NewTaxRate
	o.FinalAmount = o.calculateFinalAmount()
	o.UpdatedAt = event.ChangedAt
	return nil
}

func (o *Order) applyShippingCostChanged(event *ShippingCostChanged) error {
	o.ShippingCost = event.NewShippingCost
	o.FinalAmount = o.calculateFinalAmount()
	o.UpdatedAt = event.ChangedAt
	// 메타데이터에 변경 사유 저장
	o.Metadata["shipping_cost_change_reason"] = event.Reason
	return nil
}

func (o *Order) applyMetadataUpdated(event *MetadataUpdated) error {
	if o.Metadata == nil {
		o.Metadata = make(map[string]interface{})
	}
	o.Metadata[event.Key] = event.Value
	o.UpdatedAt = event.UpdatedAt
	return nil
}
