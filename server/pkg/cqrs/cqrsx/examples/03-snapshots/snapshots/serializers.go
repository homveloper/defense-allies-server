package snapshots

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"

	"go.mongodb.org/mongo-driver/bson"

	"defense-allies-server/pkg/cqrs/cqrsx/examples/03-snapshots/domain"
)

// JSONSerializer JSON 기반 스냅샷 직렬화
type JSONSerializer struct {
	prettyPrint bool
}

// NewJSONSerializer JSON 직렬화기 생성
func NewJSONSerializer(prettyPrint bool) *JSONSerializer {
	return &JSONSerializer{
		prettyPrint: prettyPrint,
	}
}

func (s *JSONSerializer) Serialize(aggregate Aggregate) ([]byte, error) {
	if s.prettyPrint {
		return json.MarshalIndent(aggregate, "", "  ")
	}
	return json.Marshal(aggregate)
}

func (s *JSONSerializer) Deserialize(data []byte, aggregateType string) (Aggregate, error) {
	switch aggregateType {
	case "Order":
		order := domain.NewOrder()
		if err := json.Unmarshal(data, order); err != nil {
			return nil, fmt.Errorf("failed to deserialize Order: %w", err)
		}
		return order, nil
	default:
		return nil, fmt.Errorf("unsupported aggregate type: %s", aggregateType)
	}
}

func (s *JSONSerializer) GetContentType() string {
	return "application/json"
}

func (s *JSONSerializer) GetCompressionType() string {
	return "none"
}

// BSONSerializer BSON 기반 스냅샷 직렬화
type BSONSerializer struct{}

// NewBSONSerializer BSON 직렬화기 생성
func NewBSONSerializer() *BSONSerializer {
	return &BSONSerializer{}
}

func (s *BSONSerializer) Serialize(aggregate Aggregate) ([]byte, error) {
	return bson.Marshal(aggregate)
}

func (s *BSONSerializer) Deserialize(data []byte, aggregateType string) (Aggregate, error) {
	switch aggregateType {
	case "Order":
		order := domain.NewOrder()
		if err := bson.Unmarshal(data, order); err != nil {
			return nil, fmt.Errorf("failed to deserialize Order: %w", err)
		}
		return order, nil
	default:
		return nil, fmt.Errorf("unsupported aggregate type: %s", aggregateType)
	}
}

func (s *BSONSerializer) GetContentType() string {
	return "application/bson"
}

func (s *BSONSerializer) GetCompressionType() string {
	return "none"
}

// CompressedJSONSerializer 압축된 JSON 직렬화
type CompressedJSONSerializer struct {
	baseSerializer  *JSONSerializer
	compressionType string
}

// NewCompressedJSONSerializer 압축된 JSON 직렬화기 생성
func NewCompressedJSONSerializer(compressionType string, prettyPrint bool) *CompressedJSONSerializer {
	if compressionType != "gzip" {
		compressionType = "gzip" // 기본값
	}

	return &CompressedJSONSerializer{
		baseSerializer:  NewJSONSerializer(prettyPrint),
		compressionType: compressionType,
	}
}

func (s *CompressedJSONSerializer) Serialize(aggregate Aggregate) ([]byte, error) {
	// 먼저 JSON으로 직렬화
	jsonData, err := s.baseSerializer.Serialize(aggregate)
	if err != nil {
		return nil, err
	}

	// 압축
	return s.compress(jsonData)
}

func (s *CompressedJSONSerializer) Deserialize(data []byte, aggregateType string) (Aggregate, error) {
	// 압축 해제
	decompressed, err := s.decompress(data)
	if err != nil {
		return nil, fmt.Errorf("failed to decompress data: %w", err)
	}

	// JSON 역직렬화
	return s.baseSerializer.Deserialize(decompressed, aggregateType)
}

func (s *CompressedJSONSerializer) GetContentType() string {
	return "application/json"
}

func (s *CompressedJSONSerializer) GetCompressionType() string {
	return s.compressionType
}

func (s *CompressedJSONSerializer) compress(data []byte) ([]byte, error) {
	var buf bytes.Buffer

	switch s.compressionType {
	case "gzip":
		writer := gzip.NewWriter(&buf)
		if _, err := writer.Write(data); err != nil {
			return nil, err
		}
		if err := writer.Close(); err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("unsupported compression type: %s", s.compressionType)
	}

	return buf.Bytes(), nil
}

func (s *CompressedJSONSerializer) decompress(data []byte) ([]byte, error) {
	buf := bytes.NewReader(data)

	switch s.compressionType {
	case "gzip":
		reader, err := gzip.NewReader(buf)
		if err != nil {
			return nil, err
		}
		defer reader.Close()
		return io.ReadAll(reader)
	default:
		return nil, fmt.Errorf("unsupported compression type: %s", s.compressionType)
	}
}

// CompressedBSONSerializer 압축된 BSON 직렬화
type CompressedBSONSerializer struct {
	baseSerializer  *BSONSerializer
	compressionType string
}

// NewCompressedBSONSerializer 압축된 BSON 직렬화기 생성
func NewCompressedBSONSerializer(compressionType string) *CompressedBSONSerializer {
	if compressionType != "gzip" {
		compressionType = "gzip" // 기본값
	}

	return &CompressedBSONSerializer{
		baseSerializer:  NewBSONSerializer(),
		compressionType: compressionType,
	}
}

func (s *CompressedBSONSerializer) Serialize(aggregate Aggregate) ([]byte, error) {
	// 먼저 BSON으로 직렬화
	bsonData, err := s.baseSerializer.Serialize(aggregate)
	if err != nil {
		return nil, err
	}

	// 압축
	return s.compress(bsonData)
}

func (s *CompressedBSONSerializer) Deserialize(data []byte, aggregateType string) (Aggregate, error) {
	// 압축 해제
	decompressed, err := s.decompress(data)
	if err != nil {
		return nil, fmt.Errorf("failed to decompress data: %w", err)
	}

	// BSON 역직렬화
	return s.baseSerializer.Deserialize(decompressed, aggregateType)
}

func (s *CompressedBSONSerializer) GetContentType() string {
	return "application/bson"
}

func (s *CompressedBSONSerializer) GetCompressionType() string {
	return s.compressionType
}

func (s *CompressedBSONSerializer) compress(data []byte) ([]byte, error) {
	var buf bytes.Buffer

	switch s.compressionType {
	case "gzip":
		writer := gzip.NewWriter(&buf)
		if _, err := writer.Write(data); err != nil {
			return nil, err
		}
		if err := writer.Close(); err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("unsupported compression type: %s", s.compressionType)
	}

	return buf.Bytes(), nil
}

func (s *CompressedBSONSerializer) decompress(data []byte) ([]byte, error) {
	buf := bytes.NewReader(data)

	switch s.compressionType {
	case "gzip":
		reader, err := gzip.NewReader(buf)
		if err != nil {
			return nil, err
		}
		defer reader.Close()
		return io.ReadAll(reader)
	default:
		return nil, fmt.Errorf("unsupported compression type: %s", s.compressionType)
	}
}

// SerializerFactory 직렬화기 팩토리
type SerializerFactory struct{}

// NewSerializerFactory 팩토리 생성
func NewSerializerFactory() *SerializerFactory {
	return &SerializerFactory{}
}

// CreateSerializer 직렬화기 생성
func (f *SerializerFactory) CreateSerializer(serializerType, compressionType string, options map[string]interface{}) (SnapshotSerializer, error) {
	switch serializerType {
	case "json":
		prettyPrint := false
		if val, ok := options["pretty_print"]; ok {
			if pp, ok := val.(bool); ok {
				prettyPrint = pp
			}
		}

		if compressionType == "none" || compressionType == "" {
			return NewJSONSerializer(prettyPrint), nil
		} else {
			return NewCompressedJSONSerializer(compressionType, prettyPrint), nil
		}

	case "bson":
		if compressionType == "none" || compressionType == "" {
			return NewBSONSerializer(), nil
		} else {
			return NewCompressedBSONSerializer(compressionType), nil
		}

	default:
		return nil, fmt.Errorf("unsupported serializer type: %s", serializerType)
	}
}
