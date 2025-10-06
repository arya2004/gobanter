package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

//
// ---- Mock Setup ----
//

// Mock WebSocket upgrader
var upgrader = websocket.Upgrader{}

// Define a simplified version of your structures
type WsPayload struct {
	Action   string          `json:"action"`
	Username string          `json:"username"`
	Message  string          `json:"message"`
	Conn     *websocket.Conn `json:"-"`
}

type WsJsonResponse struct {
	Action         string   `json:"action"`
	Message        string   `json:"message,omitempty"`
	ConnectedUsers []string `json:"connected_users,omitempty"`
	TimeStamp      string   `json:"timestamp,omitempty"`
}

// Mock global variables (replace these with actual ones if already defined)
var (
	wsChannel = make(chan WsPayload, 5)
	clients   = make(map[*websocket.Conn]string)
	mu        sync.Mutex
)

// ---- Mock helper functions ----
func getUserList() []string {
	mu.Lock()
	defer mu.Unlock()
	users := []string{}
	for _, u := range clients {
		users = append(users, u)
	}
	return users
}

func broadcastToAll(response WsJsonResponse) {
	mu.Lock()
	defer mu.Unlock()
	for c := range clients {
		_ = c.WriteJSON(response)
	}
}

func handlePrivateMessage(e WsPayload) {
	// mock function for private message
	fmt.Printf("Private message to %s: %s\n", e.Username, e.Message)
}

//
// ---- tests ----
//

// test 1: Ensure ListenForWs reads JSON & sends it to wsChannel

func TestListenForWs(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			t.Fatalf("Failed to upgrade: %v", err)
		}
		defer conn.Close()
		go ListenForWs(conn)
	}))
	defer server.Close()

	wsURL := "ws" + server.URL[4:]
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("Dial failed: %v", err)
	}
	defer conn.Close()

	msg := WsPayload{Message: "hello test"}
	data, _ := json.Marshal(msg)

	if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
		t.Fatalf("Failed to send message: %v", err)
	}

	select {
	case payload := <-wsChannel:
		if payload.Message != "hello test" {
			t.Errorf("Expected 'hello test', got %v", payload.Message)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Timeout: wsChannel did not receive message")
	}
}

// test 2: Ensure ListenToWsChannel handles broadcast and username actions

func TestListenToWsChannel(t *testing.T) {
	// Setup mock WebSocket client pair
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			t.Fatalf("Failed to upgrade: %v", err)
		}
		defer conn.Close()
		mu.Lock()
		clients[conn] = "tester"
		mu.Unlock()
		go ListenToWsChannel()
		select {} // keep server running
	}))
	defer server.Close()

	wsURL := "ws" + server.URL[4:]
	clientConn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("Failed to connect client: %v", err)
	}
	defer clientConn.Close()

	// Send fake payload to channel (broadcast action)

	wsChannel <- WsPayload{
		Action:   "broadcast",
		Username: "tester",
		Message:  "this is a test broadcast",
	}

	// Expect a message from broadcast

	clientConn.SetReadDeadline(time.Now().Add(2 * time.Second))
	_, msg, err := clientConn.ReadMessage()
	if err != nil {
		t.Fatalf("Client did not receive broadcast: %v", err)
	}

	if !contains(string(msg), "this is a test broadcast") {
		t.Errorf("Expected broadcast message, got: %s", string(msg))
	}
}

//
// ---- Utility function ----
//

func contains(str, substr string) bool {
	return len(str) >= len(substr) && (str == substr || (len(str) > len(substr) && (fmt.Sprintf("%v", str) != "" && (string([]rune(str)[0:len(substr)]) == substr || contains(string([]rune(str)[1:]), substr)))))
}
