package services

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"defense-allies-server/pkg/redis"
)

// Event 이벤트 구조체
type Event struct {
	ID        string      `json:"id,omitempty"`
	Type      string      `json:"type"`
	Data      interface{} `json:"data"`
	Timestamp time.Time   `json:"timestamp"`
	PlayerID  string      `json:"player_id,omitempty"`
	GameID    string      `json:"game_id,omitempty"`
}

// SSEClient SSE 클라이언트 정보
type SSEClient struct {
	ID       string
	PlayerID string
	GameID   string
	Channel  chan Event
	Writer   http.ResponseWriter
	Request  *http.Request
	LastSeen time.Time
}

// EventService 이벤트 서비스
type EventService struct {
	redisClient *redis.Client
	clients     map[string]*SSEClient
	gameClients map[string]map[string]*SSEClient // gameID -> clientID -> client
	register    chan *SSEClient
	unregister  chan *SSEClient
	broadcast   chan Event
	mutex       sync.RWMutex
}

// NewEventService 새로운 이벤트 서비스를 생성합니다
func NewEventService(redisClient *redis.Client) *EventService {
	return &EventService{
		redisClient: redisClient,
		clients:     make(map[string]*SSEClient),
		gameClients: make(map[string]map[string]*SSEClient),
		register:    make(chan *SSEClient),
		unregister:  make(chan *SSEClient),
		broadcast:   make(chan Event),
	}
}

// Start 이벤트 서비스를 시작합니다
func (s *EventService) Start() {
	go s.run()
	go s.subscribeToRedis()
}

// run 이벤트 서비스의 메인 루프
func (s *EventService) run() {
	ticker := time.NewTicker(30 * time.Second) // 30초마다 핑
	defer ticker.Stop()

	for {
		select {
		case client := <-s.register:
			s.registerClient(client)

		case client := <-s.unregister:
			s.unregisterClient(client)

		case event := <-s.broadcast:
			s.broadcastEvent(event)

		case <-ticker.C:
			// 주기적으로 핑 이벤트 전송
			s.SendToAll(Event{
				Type:      "ping",
				Data:      map[string]interface{}{"timestamp": time.Now().Unix()},
				Timestamp: time.Now(),
			})
		}
	}
}

// registerClient 클라이언트를 등록합니다
func (s *EventService) registerClient(client *SSEClient) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.clients[client.ID] = client

	// 게임별 클라이언트 그룹에 추가
	if client.GameID != "" {
		if s.gameClients[client.GameID] == nil {
			s.gameClients[client.GameID] = make(map[string]*SSEClient)
		}
		s.gameClients[client.GameID][client.ID] = client
	}

	log.Printf("SSE client registered: %s (Player: %s, Game: %s)", 
		client.ID, client.PlayerID, client.GameID)
}

// unregisterClient 클라이언트를 등록 해제합니다
func (s *EventService) unregisterClient(client *SSEClient) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if _, ok := s.clients[client.ID]; ok {
		delete(s.clients, client.ID)
		close(client.Channel)

		// 게임별 클라이언트 그룹에서 제거
		if client.GameID != "" && s.gameClients[client.GameID] != nil {
			delete(s.gameClients[client.GameID], client.ID)
			if len(s.gameClients[client.GameID]) == 0 {
				delete(s.gameClients, client.GameID)
			}
		}
	}

	log.Printf("SSE client unregistered: %s", client.ID)
}

// broadcastEvent 이벤트를 브로드캐스트합니다
func (s *EventService) broadcastEvent(event Event) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	// 특정 게임의 이벤트인 경우
	if event.GameID != "" {
		if gameClients, ok := s.gameClients[event.GameID]; ok {
			for _, client := range gameClients {
				s.sendToClient(client, event)
			}
		}
		return
	}

	// 전체 브로드캐스트
	for _, client := range s.clients {
		s.sendToClient(client, event)
	}
}

// sendToClient 특정 클라이언트에게 이벤트를 전송합니다
func (s *EventService) sendToClient(client *SSEClient, event Event) {
	select {
	case client.Channel <- event:
	default:
		log.Printf("Failed to send event to client %s: channel full", client.ID)
	}
}

// HandleSSE SSE HTTP 핸들러
func (s *EventService) HandleSSE(w http.ResponseWriter, r *http.Request) {
	// SSE 헤더 설정
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Cache-Control")

	// 클라이언트 정보 추출
	clientID := r.URL.Query().Get("client_id")
	if clientID == "" {
		clientID = fmt.Sprintf("client_%d", time.Now().UnixNano())
	}

	playerID := r.URL.Query().Get("player_id")
	gameID := r.URL.Query().Get("game_id")

	// 새 클라이언트 생성
	client := &SSEClient{
		ID:       clientID,
		PlayerID: playerID,
		GameID:   gameID,
		Channel:  make(chan Event, 10),
		Writer:   w,
		Request:  r,
		LastSeen: time.Now(),
	}

	// 클라이언트 등록
	s.register <- client

	// 연결 해제 시 정리
	defer func() {
		s.unregister <- client
	}()

	// 이벤트 전송 루프
	for {
		select {
		case event := <-client.Channel:
			if err := s.writeEvent(w, event); err != nil {
				log.Printf("Error writing SSE event: %v", err)
				return
			}

		case <-r.Context().Done():
			log.Printf("SSE client disconnected: %s", clientID)
			return
		}
	}
}

// writeEvent 이벤트를 클라이언트에게 전송합니다
func (s *EventService) writeEvent(w http.ResponseWriter, event Event) error {
	if event.ID != "" {
		fmt.Fprintf(w, "id: %s\n", event.ID)
	}

	if event.Type != "" {
		fmt.Fprintf(w, "event: %s\n", event.Type)
	}

	// 데이터를 JSON으로 직렬화
	data, err := json.Marshal(event.Data)
	if err != nil {
		return err
	}

	fmt.Fprintf(w, "data: %s\n\n", string(data))

	// 즉시 전송
	if flusher, ok := w.(http.Flusher); ok {
		flusher.Flush()
	}

	return nil
}

// SendToAll 모든 클라이언트에게 이벤트를 전송합니다
func (s *EventService) SendToAll(event Event) {
	event.Timestamp = time.Now()
	s.broadcast <- event
}

// SendToGame 특정 게임의 모든 클라이언트에게 이벤트를 전송합니다
func (s *EventService) SendToGame(gameID string, event Event) {
	event.GameID = gameID
	event.Timestamp = time.Now()
	s.broadcast <- event
}

// SendToPlayer 특정 플레이어에게 이벤트를 전송합니다
func (s *EventService) SendToPlayer(playerID string, event Event) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	event.PlayerID = playerID
	event.Timestamp = time.Now()

	for _, client := range s.clients {
		if client.PlayerID == playerID {
			s.sendToClient(client, event)
		}
	}
}

// PublishToRedis Redis Pub/Sub로 이벤트를 발행합니다
func (s *EventService) PublishToRedis(channel string, event Event) error {
	event.Timestamp = time.Now()
	data, err := json.Marshal(event)
	if err != nil {
		return err
	}

	return s.redisClient.Publish(channel, string(data))
}

// subscribeToRedis Redis Pub/Sub에서 이벤트를 구독합니다
func (s *EventService) subscribeToRedis() {
	pubsub := s.redisClient.Subscribe("events:*")
	defer pubsub.Close()

	ch := pubsub.Channel()
	for msg := range ch {
		var event Event
		if err := json.Unmarshal([]byte(msg.Payload), &event); err != nil {
			log.Printf("Error unmarshaling Redis event: %v", err)
			continue
		}

		s.broadcast <- event
	}
}

// GetClientCount 연결된 클라이언트 수를 반환합니다
func (s *EventService) GetClientCount() int {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return len(s.clients)
}

// GetGameClientCount 특정 게임의 클라이언트 수를 반환합니다
func (s *EventService) GetGameClientCount(gameID string) int {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	
	if gameClients, ok := s.gameClients[gameID]; ok {
		return len(gameClients)
	}
	return 0
}
