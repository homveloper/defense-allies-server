package uuid

import (
	"github.com/google/uuid"
)

type Generator interface {
	NewID() string
}

type UUIDGenerator struct{}

func NewUUIDGenerator() *UUIDGenerator {
	return &UUIDGenerator{}
}

func (g *UUIDGenerator) NewID() string {
	return uuid.New().String()
}