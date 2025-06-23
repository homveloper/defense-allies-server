# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Defense Allies Server is a cooperative multiplayer tower defense game backend built with a sophisticated CQRS/Event Sourcing architecture. The project features an innovative AI balancing system using matrix mathematics and transformer models to ensure dynamic game balance.

## Architecture

### High-Level Structure
- **Server** (`server/`): Go-based backend with CQRS/Event Sourcing architecture
- **Client** (`client/`): Next.js 15 + React 19 frontend with modern tooling
- **Documentation** (`docs/`): Comprehensive design documents and architectural decisions

### Server Architecture
The backend follows Domain-Driven Design (DDD) with CQRS and supports multiple storage strategies:

- **Core CQRS Framework** (`server/pkg/cqrs/`): Complete CQRS implementation with Redis and in-memory support
- **Domain Aggregates**: User, Game Session, Guild, and other business entities
- **Storage Strategy**: Flexible approach supporting Event Sourcing, State-based, and Hybrid patterns
- **Multiple Apps**: Guardian (auth), TimeSquare (game), Command (ops), Health (monitoring)

## Development Commands

### Server (Go)
Navigate to `server/` directory for all server commands:

```bash
# Development setup
make dev-setup          # Install dev tools (golangci-lint)
make tidy               # Clean up dependencies

# Building and running
make build              # Build the server binary
make run                # Run the development server
go run ./cmd/metropolis/main.go  # Alternative run method

# Testing
make test               # Run all tests
make test-coverage      # Run tests with coverage report
cd pkg/cqrs && go test -v ./...  # Run CQRS tests specifically

# Code quality
make lint               # Run linter
make fmt                # Format code

# Cleanup
make clean              # Remove build artifacts
```

### Client (Next.js)
Navigate to `client/` directory for all client commands:

```bash
# Development
npm run dev             # Start dev server with Turbopack
npm start              # Start production server

# Building
npm run build          # Build for production

# Code quality
npm run lint           # Run ESLint
```

### Examples and Testing
The project includes comprehensive examples in `server/examples/`:

```bash
# Run user management example (demonstrates full CQRS flow)
cd server/examples/user && go run main.go

# Run guild management example
cd server/examples/guild && go run main.go

# Run cargo domain example
cd server/examples/cargo && go run main.go
```

## Key Components and Patterns

### CQRS Implementation
- **Commands**: State-changing operations (CreateUser, PlaceTower, etc.)
- **Queries**: Data retrieval operations (GetUser, ListActiveSessions, etc.)
- **Events**: Domain events for state changes (UserCreated, TowerPlaced, etc.)
- **Projections**: Read model builders from event streams
- **Repositories**: Flexible storage with Event Sourcing, State-based, or Hybrid patterns

### Storage Strategy Configuration
The system supports three storage patterns configurable per aggregate:

```yaml
# Example configuration (configs/cqrs.yaml)
aggregates:
  user: event_sourced      # Full audit trail for users
  game_session: hybrid     # Events + fast state access
  player_stats: state_based # Simple CRUD for statistics
```

### Redis Infrastructure
- **Event Store**: Redis Streams for event sourcing
- **State Store**: Redis Hash for state-based storage  
- **Read Store**: Redis Hash/Sets for read models
- **Event Bus**: Redis Pub/Sub for real-time events

## Development Guidelines

### Testing Strategy
- Unit tests for all core components (74+ tests in CQRS package)
- Integration tests with Redis containers
- Performance benchmarks for storage operations
- Example applications demonstrating complete flows

### Code Organization
- Domain logic in `domain/` packages (pure business logic)
- Application services in `application/` packages (orchestration)
- Infrastructure in `infrastructure/` packages (external concerns)
- Shared interfaces in `pkg/cqrs/` (framework code)

### Configuration Management
- Environment-specific configs in `server/configs/`
- Go modules with workspace setup (`go.work`)
- Redis connection pooling and metrics
- Flexible serialization (JSON, BSON, etc.)

## AI Balancing System
The project includes an innovative AI balancing system:
- **Matrix Balancing**: 18 races × 162 towers with mathematical balance guarantees
- **Transformer AI**: Real-time game balance adjustments using modern ML
- **Environmental Factors**: 120 combinations affecting gameplay (time/weather/terrain)

See `docs/discussion/` for detailed system design documents.

## Debugging and Troubleshooting

### Common Issues
1. **Redis Connection**: Ensure Redis is running on localhost:6379 for examples
2. **Go Workspace**: The project uses Go workspaces - run commands from correct directories
3. **Module Dependencies**: Use `make tidy` to resolve dependency issues

### Useful Debug Commands
```bash
# Check Redis connection
redis-cli ping

# View active Go modules
go list -m all

# Check test coverage
make test-coverage && open coverage.html

# Monitor Redis operations
redis-cli monitor
```

## Performance Considerations
- Redis-first architecture for high performance
- Connection pooling with configurable sizes
- Compression and encryption options for storage
- Batch operations for bulk data processing
- Metrics collection throughout the stack

## Security Features
- JWT-based authentication (Guardian server)
- AES-GCM encryption for sensitive data
- Configurable retention policies
- Audit trail through event sourcing

## Next.js Client Architecture
- **App Router**: Modern Next.js 15 routing
- **Turbopack**: Ultra-fast development builds
- **State Management**: Zustand + TanStack Query
- **Real-time**: Server-Sent Events integration
- **3D Graphics**: React Three Fiber for game rendering
- **Styling**: Tailwind CSS + CSS Modules