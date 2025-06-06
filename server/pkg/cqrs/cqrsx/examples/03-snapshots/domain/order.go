package domain

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
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
	*cqrs.BaseAggregate

	// 비즈니스 필드들 (private)
	customerID   string
	items        []OrderItem
	status       OrderStatus
	totalAmount  decimal.Decimal
	discountRate decimal.Decimal
	taxRate      decimal.Decimal
	shippingCost decimal.Decimal
	finalAmount  decimal.Decimal
	createdAt    time.Time
	updatedAt    time.Time
	confirmedAt  *time.Time
	shippedAt    *time.Time
	deliveredAt  *time.Time
	cancelledAt  *time.Time
	metadata     map[string]interface{}
}

// NewOrder 새로운 주문 생성
func NewOrder() *Order {
	id := uuid.New().String()

	order := &Order{
		BaseAggregate: cqrs.NewBaseAggregate(id, "Order"),
		items:         make([]OrderItem, 0),
		status:        OrderStatusPending,
		totalAmount:   decimal.Zero,
		discountRate:  decimal.Zero,
		taxRate:       decimal.NewFromFloat(0.1), // 기본 10% 세율
		shippingCost:  decimal.Zero,
		finalAmount:   decimal.Zero,
		metadata:      make(map[string]interface{}),
	}
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

	// 이벤트 즉시 적용
	if err := o.ApplyEvent(eventMessage); err != nil {
		return fmt.Errorf("failed to apply OrderCreated event: %w", err)
	}

	return nil
}

// AddItem 상품 추가
func (o *Order) AddItem(productID, productName string, quantity int, unitPrice decimal.Decimal) error {
	if err := o.validateItemAddition(productID, quantity, unitPrice); err != nil {
		return err
	}

	if o.Status() != OrderStatusPending {
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

	// 이벤트 즉시 적용
	if err := o.ApplyEvent(eventMessage); err != nil {
		return fmt.Errorf("failed to apply ItemAdded event: %w", err)
	}

	return nil
}

// RemoveItem 상품 제거
func (o *Order) RemoveItem(productID string) error {
	if o.Status() != OrderStatusPending {
		return errors.New("cannot remove items from non-pending order")
	}

	// 상품이 존재하는지 확인
	found := false
	for _, item := range o.Items() {
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

	if o.Status() != OrderStatusPending {
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
	if o.Status() != OrderStatusPending {
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

	// 이벤트 즉시 적용
	if err := o.ApplyEvent(eventMessage); err != nil {
		return fmt.Errorf("failed to apply DiscountApplied event: %w", err)
	}

	return nil
}

// ConfirmOrder 주문 확정
func (o *Order) ConfirmOrder() error {
	if o.Status() != OrderStatusPending {
		return fmt.Errorf("cannot confirm order with status %s", o.Status())
	}

	if len(o.Items()) == 0 {
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

	// 이벤트 즉시 적용
	if err := o.ApplyEvent(eventMessage); err != nil {
		return fmt.Errorf("failed to apply OrderConfirmed event: %w", err)
	}

	return nil
}

// ShipOrder 주문 배송 시작
func (o *Order) ShipOrder(trackingNumber string) error {
	if o.Status() != OrderStatusConfirmed {
		return fmt.Errorf("cannot ship order with status %s", o.Status())
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
	if o.Status() != OrderStatusShipped {
		return fmt.Errorf("cannot deliver order with status %s", o.Status())
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
	if o.Status() == OrderStatusDelivered {
		return errors.New("cannot cancel delivered order")
	}

	if o.Status() == OrderStatusCancelled {
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

// Getter 메서드들 (private 필드 접근용)
func (o *Order) CustomerID() string {
	return o.customerID
}

func (o *Order) Items() []OrderItem {
	// 방어적 복사를 통해 불변성 보장
	items := make([]OrderItem, len(o.items))
	copy(items, o.items)
	return items
}

func (o *Order) Status() OrderStatus {
	return o.status
}

func (o *Order) TotalAmount() decimal.Decimal {
	return o.totalAmount
}

func (o *Order) DiscountRate() decimal.Decimal {
	return o.discountRate
}

func (o *Order) TaxRate() decimal.Decimal {
	return o.taxRate
}

func (o *Order) ShippingCost() decimal.Decimal {
	return o.shippingCost
}

func (o *Order) FinalAmount() decimal.Decimal {
	return o.finalAmount
}

func (o *Order) CreatedAt() time.Time {
	return o.createdAt
}

func (o *Order) UpdatedAt() time.Time {
	return o.updatedAt
}

func (o *Order) ConfirmedAt() *time.Time {
	return o.confirmedAt
}

func (o *Order) ShippedAt() *time.Time {
	return o.shippedAt
}

func (o *Order) DeliveredAt() *time.Time {
	return o.deliveredAt
}

func (o *Order) CancelledAt() *time.Time {
	return o.cancelledAt
}

func (o *Order) Metadata() map[string]interface{} {
	// 방어적 복사를 통해 불변성 보장
	metadata := make(map[string]interface{})
	for k, v := range o.metadata {
		metadata[k] = v
	}
	return metadata
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
	return len(o.items)
}

func (o *Order) IsEmpty() bool {
	return len(o.items) == 0
}

func (o *Order) HasItem(productID string) bool {
	for _, item := range o.items {
		if item.ProductID == productID {
			return true
		}
	}
	return false
}

func (o *Order) GetItem(productID string) (OrderItem, bool) {
	for _, item := range o.items {
		if item.ProductID == productID {
			return item, true
		}
	}
	return OrderItem{}, false
}

// String 주문 정보를 문자열로 반환
func (o *Order) String() string {
	return fmt.Sprintf("Order{ID: %s, Customer: %s, Status: %s, Items: %d, Total: %s, Version: %d}",
		o.ID(), o.customerID, o.Status(), len(o.Items()), o.FinalAmount().String(), o.Version())
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
	for _, item := range o.items {
		total = total.Add(item.TotalPrice)
	}
	return total
}

func (o *Order) calculateFinalAmount() decimal.Decimal {
	// 총액 계산
	subtotal := o.calculateTotalAmount()

	// 할인 적용
	discountAmount := subtotal.Mul(o.discountRate)
	afterDiscount := subtotal.Sub(discountAmount)

	// 세금 적용
	taxAmount := afterDiscount.Mul(o.taxRate)
	afterTax := afterDiscount.Add(taxAmount)

	// 배송비 추가
	finalAmount := afterTax.Add(o.shippingCost)

	return finalAmount
}

// ApplyEvent 새로운 이벤트를 적용하여 상태 변경 (추적함)
func (o *Order) ApplyEvent(event cqrs.EventMessage) error {
	// BaseAggregate의 ApplyEvent 메서드 호출 (버전 관리 및 추적)
	if err := o.BaseAggregate.ApplyEvent(event); err != nil {
		return err
	}

	// 도메인별 이벤트 적용
	return o.applyDomainEvent(event)
}

// ReplayEvent 기존 이벤트를 재생하여 상태 변경 (추적하지 않음)
func (o *Order) ReplayEvent(event cqrs.EventMessage) error {
	// BaseAggregate의 ReplayEvent 메서드 호출 (버전 관리만)
	if err := o.BaseAggregate.ReplayEvent(event); err != nil {
		return err
	}

	// 도메인별 이벤트 적용
	return o.applyDomainEvent(event)
}

// applyDomainEvent 공통 도메인 이벤트 적용 로직
func (o *Order) applyDomainEvent(event cqrs.EventMessage) error {

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
	o.customerID = event.CustomerID
	o.shippingCost = event.ShippingCost
	o.status = OrderStatusPending
	o.createdAt = event.CreatedAt
	o.updatedAt = event.CreatedAt
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
	o.items = append(o.items, item)
	o.totalAmount = o.calculateTotalAmount()
	o.finalAmount = o.calculateFinalAmount()
	o.updatedAt = event.AddedAt
	return nil
}

func (o *Order) applyItemRemoved(event *ItemRemoved) error {
	for i, item := range o.items {
		if item.ProductID == event.ProductID {
			o.items = append(o.items[:i], o.items[i+1:]...)
			break
		}
	}
	o.totalAmount = o.calculateTotalAmount()
	o.finalAmount = o.calculateFinalAmount()
	o.updatedAt = event.RemovedAt
	return nil
}

func (o *Order) applyItemQuantityChanged(event *ItemQuantityChanged) error {
	for i, item := range o.items {
		if item.ProductID == event.ProductID {
			o.items[i].Quantity = event.NewQuantity
			o.items[i].TotalPrice = item.UnitPrice.Mul(decimal.NewFromInt(int64(event.NewQuantity)))
			break
		}
	}
	o.totalAmount = o.calculateTotalAmount()
	o.finalAmount = o.calculateFinalAmount()
	o.updatedAt = event.ChangedAt
	return nil
}

func (o *Order) applyDiscountApplied(event *DiscountApplied) error {
	o.discountRate = event.DiscountRate
	o.finalAmount = o.calculateFinalAmount()
	o.updatedAt = event.AppliedAt
	return nil
}

func (o *Order) applyOrderConfirmed(event *OrderConfirmed) error {
	o.status = OrderStatusConfirmed
	o.confirmedAt = &event.ConfirmedAt
	o.updatedAt = event.ConfirmedAt
	return nil
}

func (o *Order) applyOrderShipped(event *OrderShipped) error {
	o.status = OrderStatusShipped
	o.shippedAt = &event.ShippedAt
	o.updatedAt = event.ShippedAt
	// 메타데이터에 추적 번호 저장
	o.metadata["tracking_number"] = event.TrackingNumber
	return nil
}

func (o *Order) applyOrderDelivered(event *OrderDelivered) error {
	o.status = OrderStatusDelivered
	o.deliveredAt = &event.DeliveredAt
	o.updatedAt = event.DeliveredAt
	return nil
}

func (o *Order) applyOrderCancelled(event *OrderCancelled) error {
	o.status = OrderStatusCancelled
	o.cancelledAt = &event.CancelledAt
	o.updatedAt = event.CancelledAt
	// 메타데이터에 취소 사유 저장
	o.metadata["cancellation_reason"] = event.Reason
	return nil
}

func (o *Order) applyTaxRateChanged(event *TaxRateChanged) error {
	o.taxRate = event.NewTaxRate
	o.finalAmount = o.calculateFinalAmount()
	o.updatedAt = event.ChangedAt
	return nil
}

func (o *Order) applyShippingCostChanged(event *ShippingCostChanged) error {
	o.shippingCost = event.NewShippingCost
	o.finalAmount = o.calculateFinalAmount()
	o.updatedAt = event.ChangedAt
	// 메타데이터에 변경 사유 저장
	o.metadata["shipping_cost_change_reason"] = event.Reason
	return nil
}

func (o *Order) applyMetadataUpdated(event *MetadataUpdated) error {
	if o.metadata == nil {
		o.metadata = make(map[string]interface{})
	}
	o.metadata[event.Key] = event.Value
	o.updatedAt = event.UpdatedAt
	return nil
}
