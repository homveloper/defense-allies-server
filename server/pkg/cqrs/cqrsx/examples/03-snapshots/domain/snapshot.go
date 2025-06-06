package domain

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/shopspring/decimal"

	"defense-allies-server/pkg/cqrs"
)

// OrderSnapshot Order Aggregate의 스냅샷 구조체
type OrderSnapshot struct {
	ID           string                 `json:"aggregate_id" bson:"aggregate_id"`
	Ver          int                    `json:"version" bson:"version"`
	CustomerID   string                 `json:"customer_id" bson:"customer_id"`
	Items        []OrderItem            `json:"items" bson:"items"`
	Status       string                 `json:"status" bson:"status"`
	TotalAmount  string                 `json:"total_amount" bson:"total_amount"`   // decimal as string
	DiscountRate string                 `json:"discount_rate" bson:"discount_rate"` // decimal as string
	TaxRate      string                 `json:"tax_rate" bson:"tax_rate"`           // decimal as string
	ShippingCost string                 `json:"shipping_cost" bson:"shipping_cost"` // decimal as string
	FinalAmount  string                 `json:"final_amount" bson:"final_amount"`   // decimal as string
	CreatedAt    time.Time              `json:"created_at" bson:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at" bson:"updated_at"`
	ConfirmedAt  *time.Time             `json:"confirmed_at,omitempty" bson:"confirmed_at,omitempty"`
	ShippedAt    *time.Time             `json:"shipped_at,omitempty" bson:"shipped_at,omitempty"`
	DeliveredAt  *time.Time             `json:"delivered_at,omitempty" bson:"delivered_at,omitempty"`
	CancelledAt  *time.Time             `json:"cancelled_at,omitempty" bson:"cancelled_at,omitempty"`
	Metadata     map[string]interface{} `json:"metadata" bson:"metadata"`
	CreatedTime  time.Time              `json:"timestamp" bson:"timestamp"`
}

// CreateSnapshot Order에서 스냅샷을 생성하는 메서드
func (o *Order) CreateSnapshot() (cqrs.SnapshotData, error) {
	return &OrderSnapshot{
		ID:           o.ID(),
		Ver:          o.Version(),
		CustomerID:   o.CustomerID(),
		Items:        o.Items(),
		Status:       string(o.Status()),
		TotalAmount:  o.TotalAmount().String(),
		DiscountRate: o.DiscountRate().String(),
		TaxRate:      o.TaxRate().String(),
		ShippingCost: o.ShippingCost().String(),
		FinalAmount:  o.FinalAmount().String(),
		CreatedAt:    o.CreatedAt(),
		UpdatedAt:    o.UpdatedAt(),
		ConfirmedAt:  o.ConfirmedAt(),
		ShippedAt:    o.ShippedAt(),
		DeliveredAt:  o.DeliveredAt(),
		CancelledAt:  o.CancelledAt(),
		Metadata:     o.Metadata(),
		CreatedTime:  time.Now(),
	}, nil
}

// RestoreFromSnapshot 스냅샷에서 Order를 복원하는 함수
func RestoreFromSnapshot(snapshot *OrderSnapshot) (*Order, error) {
	// decimal 필드들 파싱
	totalAmount, err := decimal.NewFromString(snapshot.TotalAmount)
	if err != nil {
		return nil, err
	}

	discountRate, err := decimal.NewFromString(snapshot.DiscountRate)
	if err != nil {
		return nil, err
	}

	taxRate, err := decimal.NewFromString(snapshot.TaxRate)
	if err != nil {
		return nil, err
	}

	shippingCost, err := decimal.NewFromString(snapshot.ShippingCost)
	if err != nil {
		return nil, err
	}

	finalAmount, err := decimal.NewFromString(snapshot.FinalAmount)
	if err != nil {
		return nil, err
	}

	// Order 인스턴스 생성
	baseAggregate := cqrs.NewBaseAggregate(snapshot.ID, "Order")
	baseAggregate.SetOriginalVersion(snapshot.Ver)

	order := &Order{
		BaseAggregate: baseAggregate,
		customerID:    snapshot.CustomerID,
		items:         snapshot.Items,
		status:        OrderStatus(snapshot.Status),
		totalAmount:   totalAmount,
		discountRate:  discountRate,
		taxRate:       taxRate,
		shippingCost:  shippingCost,
		finalAmount:   finalAmount,
		createdAt:     snapshot.CreatedAt,
		updatedAt:     snapshot.UpdatedAt,
		confirmedAt:   snapshot.ConfirmedAt,
		shippedAt:     snapshot.ShippedAt,
		deliveredAt:   snapshot.DeliveredAt,
		cancelledAt:   snapshot.CancelledAt,
		metadata:      snapshot.Metadata,
	}

	return order, nil
}

// SnapshotData 인터페이스 구현
func (s *OrderSnapshot) ID() string {
	return s.ID
}

func (s *OrderSnapshot) Type() string {
	return "Order"
}

func (s *OrderSnapshot) Version() int {
	return s.Ver
}

func (s *OrderSnapshot) Data() interface{} {
	return s
}

func (s *OrderSnapshot) Timestamp() time.Time {
	return s.CreatedTime
}

// GetChecksum 체크섬 반환 (SnapshotData 인터페이스 구현)
func (s *OrderSnapshot) GetChecksum() string {
	// 간단한 체크섬 구현 (실제로는 더 복잡한 해시 알고리즘 사용)
	data, _ := json.Marshal(s)
	return fmt.Sprintf("%x", len(data))
}

// Serialize JSON으로 직렬화
func (s *OrderSnapshot) Serialize() ([]byte, error) {
	return json.Marshal(s)
}

// Deserialize JSON에서 역직렬화
func (s *OrderSnapshot) Deserialize(data []byte) error {
	return json.Unmarshal(data, s)
}

// GetSize 스냅샷 크기 반환 (바이트)
func (s *OrderSnapshot) GetSize() int64 {
	data, err := s.Serialize()
	if err != nil {
		return 0
	}
	return int64(len(data))
}

// Clone 스냅샷 복사본 생성
func (s *OrderSnapshot) Clone() *OrderSnapshot {
	clone := &OrderSnapshot{
		ID:           s.ID,
		Ver:          s.Ver,
		CustomerID:   s.CustomerID,
		Items:        make([]OrderItem, len(s.Items)),
		Status:       s.Status,
		TotalAmount:  s.TotalAmount,
		DiscountRate: s.DiscountRate,
		TaxRate:      s.TaxRate,
		ShippingCost: s.ShippingCost,
		FinalAmount:  s.FinalAmount,
		CreatedAt:    s.CreatedAt,
		UpdatedAt:    s.UpdatedAt,
		ConfirmedAt:  s.ConfirmedAt,
		ShippedAt:    s.ShippedAt,
		DeliveredAt:  s.DeliveredAt,
		CancelledAt:  s.CancelledAt,
		Metadata:     make(map[string]interface{}),
		CreatedTime:  s.CreatedTime,
	}

	// Items 복사
	copy(clone.Items, s.Items)

	// Metadata 복사
	for k, v := range s.Metadata {
		clone.Metadata[k] = v
	}

	return clone
}

// Validate 스냅샷 유효성 검증
func (s *OrderSnapshot) Validate() error {
	if s.ID == "" {
		return cqrs.ErrInvalidSnapshotData
	}
	if s.Ver < 0 {
		return cqrs.ErrInvalidSnapshotData
	}
	if s.CustomerID == "" {
		return cqrs.ErrInvalidSnapshotData
	}
	return nil
}

// IsExpired 스냅샷 만료 여부 확인
func (s *OrderSnapshot) IsExpired(ttl time.Duration) bool {
	return time.Since(s.CreatedTime) > ttl
}

// GetMetadata 메타데이터 반환
func (s *OrderSnapshot) GetMetadata() map[string]interface{} {
	metadata := make(map[string]interface{})
	for k, v := range s.Metadata {
		metadata[k] = v
	}
	return metadata
}

// SetMetadata 메타데이터 설정
func (s *OrderSnapshot) SetMetadata(key string, value interface{}) {
	if s.Metadata == nil {
		s.Metadata = make(map[string]interface{})
	}
	s.Metadata[key] = value
}

// GetItemCount 주문 상품 개수 반환
func (s *OrderSnapshot) GetItemCount() int {
	return len(s.Items)
}

// GetTotalAmountDecimal 총액을 decimal로 반환
func (s *OrderSnapshot) GetTotalAmountDecimal() (decimal.Decimal, error) {
	return decimal.NewFromString(s.TotalAmount)
}

// GetFinalAmountDecimal 최종 금액을 decimal로 반환
func (s *OrderSnapshot) GetFinalAmountDecimal() (decimal.Decimal, error) {
	return decimal.NewFromString(s.FinalAmount)
}

// String 스냅샷 정보를 문자열로 반환
func (s *OrderSnapshot) String() string {
	return fmt.Sprintf("OrderSnapshot{ID: %s, Version: %d, Customer: %s, Items: %d, Status: %s, Timestamp: %s}",
		s.ID, s.Ver, s.CustomerID, len(s.Items), s.Status, s.CreatedTime.Format(time.RFC3339))
}
