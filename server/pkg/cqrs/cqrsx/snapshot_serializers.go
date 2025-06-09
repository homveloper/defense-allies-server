package cqrsx

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"

	"go.mongodb.org/mongo-driver/bson"

	"cqrs"
)

// AdvancedSnapshotSerializer extends the basic SnapshotSerializer with compression and content type support
type AdvancedSnapshotSerializer interface {
	SnapshotSerializer
	GetContentType() string
	GetCompressionType() string
}

// JSONSnapshotSerializer JSON-based snapshot serialization
type JSONSnapshotSerializer struct {
	prettyPrint bool
}

// NewJSONSnapshotSerializer creates a JSON serializer
func NewJSONSnapshotSerializer(prettyPrint bool) *JSONSnapshotSerializer {
	return &JSONSnapshotSerializer{
		prettyPrint: prettyPrint,
	}
}

func (s *JSONSnapshotSerializer) SerializeSnapshot(aggregate cqrs.AggregateRoot) ([]byte, error) {
	if s.prettyPrint {
		return json.MarshalIndent(aggregate, "", "  ")
	}
	return json.Marshal(aggregate)
}

func (s *JSONSnapshotSerializer) DeserializeSnapshot(data []byte, aggregateType string) (cqrs.AggregateRoot, error) {
	// This would need to be implemented based on your aggregate factory
	// For now, return an error indicating the need for implementation
	return nil, fmt.Errorf("aggregate factory not implemented for type: %s", aggregateType)
}

func (s *JSONSnapshotSerializer) GetContentType() string {
	return "application/json"
}

func (s *JSONSnapshotSerializer) GetCompressionType() string {
	return "none"
}

// BSONSnapshotSerializer BSON-based snapshot serialization
type BSONSnapshotSerializer struct{}

// NewBSONSnapshotSerializer creates a BSON serializer
func NewBSONSnapshotSerializer() *BSONSnapshotSerializer {
	return &BSONSnapshotSerializer{}
}

func (s *BSONSnapshotSerializer) SerializeSnapshot(aggregate cqrs.AggregateRoot) ([]byte, error) {
	return bson.Marshal(aggregate)
}

func (s *BSONSnapshotSerializer) DeserializeSnapshot(data []byte, aggregateType string) (cqrs.AggregateRoot, error) {
	// This would need to be implemented based on your aggregate factory
	// For now, return an error indicating the need for implementation
	return nil, fmt.Errorf("aggregate factory not implemented for type: %s", aggregateType)
}

func (s *BSONSnapshotSerializer) GetContentType() string {
	return "application/bson"
}

func (s *BSONSnapshotSerializer) GetCompressionType() string {
	return "none"
}

// CompressedJSONSnapshotSerializer compressed JSON serialization
type CompressedJSONSnapshotSerializer struct {
	baseSerializer  *JSONSnapshotSerializer
	compressionType string
}

// NewCompressedJSONSnapshotSerializer creates a compressed JSON serializer
func NewCompressedJSONSnapshotSerializer(compressionType string, prettyPrint bool) *CompressedJSONSnapshotSerializer {
	if compressionType != "gzip" {
		compressionType = "gzip" // default
	}

	return &CompressedJSONSnapshotSerializer{
		baseSerializer:  NewJSONSnapshotSerializer(prettyPrint),
		compressionType: compressionType,
	}
}

func (s *CompressedJSONSnapshotSerializer) SerializeSnapshot(aggregate cqrs.AggregateRoot) ([]byte, error) {
	// First serialize to JSON
	jsonData, err := s.baseSerializer.SerializeSnapshot(aggregate)
	if err != nil {
		return nil, err
	}

	// Then compress
	return s.compress(jsonData)
}

func (s *CompressedJSONSnapshotSerializer) DeserializeSnapshot(data []byte, aggregateType string) (cqrs.AggregateRoot, error) {
	// Decompress first
	decompressed, err := s.decompress(data)
	if err != nil {
		return nil, fmt.Errorf("failed to decompress data: %w", err)
	}

	// Then deserialize JSON
	return s.baseSerializer.DeserializeSnapshot(decompressed, aggregateType)
}

func (s *CompressedJSONSnapshotSerializer) GetContentType() string {
	return "application/json"
}

func (s *CompressedJSONSnapshotSerializer) GetCompressionType() string {
	return s.compressionType
}

func (s *CompressedJSONSnapshotSerializer) compress(data []byte) ([]byte, error) {
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

func (s *CompressedJSONSnapshotSerializer) decompress(data []byte) ([]byte, error) {
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

// CompressedBSONSnapshotSerializer compressed BSON serialization
type CompressedBSONSnapshotSerializer struct {
	baseSerializer  *BSONSnapshotSerializer
	compressionType string
}

// NewCompressedBSONSnapshotSerializer creates a compressed BSON serializer
func NewCompressedBSONSnapshotSerializer(compressionType string) *CompressedBSONSnapshotSerializer {
	if compressionType != "gzip" {
		compressionType = "gzip" // default
	}

	return &CompressedBSONSnapshotSerializer{
		baseSerializer:  NewBSONSnapshotSerializer(),
		compressionType: compressionType,
	}
}

func (s *CompressedBSONSnapshotSerializer) SerializeSnapshot(aggregate cqrs.AggregateRoot) ([]byte, error) {
	// First serialize to BSON
	bsonData, err := s.baseSerializer.SerializeSnapshot(aggregate)
	if err != nil {
		return nil, err
	}

	// Then compress
	return s.compress(bsonData)
}

func (s *CompressedBSONSnapshotSerializer) DeserializeSnapshot(data []byte, aggregateType string) (cqrs.AggregateRoot, error) {
	// Decompress first
	decompressed, err := s.decompress(data)
	if err != nil {
		return nil, fmt.Errorf("failed to decompress data: %w", err)
	}

	// Then deserialize BSON
	return s.baseSerializer.DeserializeSnapshot(decompressed, aggregateType)
}

func (s *CompressedBSONSnapshotSerializer) GetContentType() string {
	return "application/bson"
}

func (s *CompressedBSONSnapshotSerializer) GetCompressionType() string {
	return s.compressionType
}

func (s *CompressedBSONSnapshotSerializer) compress(data []byte) ([]byte, error) {
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

func (s *CompressedBSONSnapshotSerializer) decompress(data []byte) ([]byte, error) {
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

// SnapshotSerializerFactory creates serializers based on configuration
type SnapshotSerializerFactory struct{}

// NewSnapshotSerializerFactory creates a factory
func NewSnapshotSerializerFactory() *SnapshotSerializerFactory {
	return &SnapshotSerializerFactory{}
}

// CreateSerializer creates a serializer based on type and options
func (f *SnapshotSerializerFactory) CreateSerializer(serializerType, compressionType string, options map[string]interface{}) (AdvancedSnapshotSerializer, error) {
	switch serializerType {
	case "json":
		prettyPrint := false
		if val, ok := options["pretty_print"]; ok {
			if pp, ok := val.(bool); ok {
				prettyPrint = pp
			}
		}

		if compressionType == "none" || compressionType == "" {
			return NewJSONSnapshotSerializer(prettyPrint), nil
		} else {
			return NewCompressedJSONSnapshotSerializer(compressionType, prettyPrint), nil
		}

	case "bson":
		if compressionType == "none" || compressionType == "" {
			return NewBSONSnapshotSerializer(), nil
		} else {
			return NewCompressedBSONSnapshotSerializer(compressionType), nil
		}

	default:
		return nil, fmt.Errorf("unsupported serializer type: %s", serializerType)
	}
}
