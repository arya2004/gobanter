package handlers

import (
	"fmt"
	"log"
	"net/http"
	"sort"
	"time"

	"github.com/CloudyKit/jet/v6"
	"github.com/gorilla/websocket"
)

// Hub encapsulates all WebSocket related global state
type Hub struct {
	Clients map[*websocket.Conn]string
	Channel chan WsPayload
}

// Create a single hub instance

var hub = Hub{
	Clients: make(map[*websocket.Conn]string),
	Channel: make(chan WsPayload),
}

// template engine and WebSocket upgrader

var (
	views = jet.NewSet(
		jet.NewOSFileSystemLoader("./templates"),
		jet.InDevelopmentMode(),
	)
	upgradeConnection = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin:     func(r *http.Request) bool { return true },
	}
)

// Home renders the home page of the application

func Home(w http.ResponseWriter, r *http.Request) {
	log.Println("Rendering home page")
	if err := renderPage(w, "home.html", nil); err != nil {
		log.Println("Error rendering home page:", err)
	}
}

// WsJsonResponse represents the JSON structure for WebSocket responses
type WsJsonResponse struct {
	Action         string   `json:"action"`
	Message        string   `json:"message"`
	MessageType    string   `json:"message_type"`
	ConnectedUsers []string `json:"connected_users"`
	From           string   `json:"from"`
	To             string   `json:"to"`
	TimeStamp      string   `json:"timestamp"`
}

// WsPayload represents the payload received from WebSocket clients
type WsPayload struct {
	Action   string          `json:"action"`
	Username string          `json:"username"`
	Message  string          `json:"message"`
	To       string          `json:"to"`
	Conn     *websocket.Conn `json:"-"`
}

// WsEndpoint upgrades HTTP connections to WebSocket and initializes communication

func WsEndpoint(w http.ResponseWriter, r *http.Request) {
	log.Println("Client attempting to connect to WebSocket endpoint")

	ws, err := upgradeConnection.Upgrade(w, r, nil)
	if err != nil {
		log.Println("WebSocket upgrade failed:", err)
		return
	}

	log.Println("Client successfully connected to WebSocket endpoint")

	response := WsJsonResponse{
		Message:   "<em><small>Connected to server</small></em>",
		TimeStamp: time.Now().Format("15:07"),
	}

	hub.Clients[ws] = "" // Add connection to Hub

	if err := ws.WriteJSON(response); err != nil {
		log.Println("Error writing JSON response:", err)
		return
	}

	go ListenForWs(ws)
}

// ListenToWsChannel listens for messages on the Hub channel

func ListenToWsChannel() {
	for {
		e := <-hub.Channel
		var response WsJsonResponse

		switch e.Action {
		case "username":
			hub.Clients[e.Conn] = e.Username
			response.Action = "list_users"
			response.ConnectedUsers = getUserList()
			broadcastToAll(response)

		case "left":
			response.Action = "list_users"
			delete(hub.Clients, e.Conn)
			response.ConnectedUsers = getUserList()
			broadcastToAll(response)

		case "broadcast":
			response.Action = "broadcast"
			response.Message = fmt.Sprintf("<strong>%s</strong>: %s", e.Username, e.Message)
			response.TimeStamp = time.Now().Format("15:07")
			broadcastToAll(response)

		case "private":
			handlePrivateMessage(e)
		}
	}
}

// handlePrivateMessage sends a message to a specific user
func handlePrivateMessage(payload WsPayload) {
	var recipientConn *websocket.Conn
	for conn, username := range hub.Clients {
		if username == payload.To {
			recipientConn = conn
			break
		}
	}

	if recipientConn == nil {
		errorResponse := WsJsonResponse{
			Action:      "error",
			Message:     fmt.Sprintf("User '%s' not found or offline", payload.To),
			MessageType: "error",
			TimeStamp:   time.Now().Format("15:07"),
		}

		if err := payload.Conn.WriteJSON(errorResponse); err != nil {
			log.Printf("Error sending error message to sender: %v", err)
		}
		return
	}

	response := WsJsonResponse{
		Action:      "private",
		Message:     payload.Message,
		MessageType: "private",
		From:        payload.Username,
		To:          payload.To,
		TimeStamp:   time.Now().Format("15:07"),
	}

	if err := recipientConn.WriteJSON(response); err != nil {
		log.Printf("Error sending private message to recipient: %v", err)
	} else {
		log.Printf("Private message sent from %s to %s", payload.Username, payload.To)
	}

	if err := payload.Conn.WriteJSON(response); err != nil {
		log.Printf("Error sending confirmation to sender: %v", err)
	}
}

// getUserList returns a sorted list of connected usernames

func getUserList() []string {
	var userList []string
	for _, username := range hub.Clients {
		if username != "" {
			userList = append(userList, username)
		}
	}
	sort.Strings(userList)
	return userList
}

// broadcastToAll sends a WebSocket response to all connected clients

func broadcastToAll(response WsJsonResponse) {
	for client := range hub.Clients {
		if err := client.WriteJSON(response); err != nil {
			log.Printf("Error sending message to client: %v", err)
			_ = client.Close()
			delete(hub.Clients, client)
		}
	}
}

// ListenForWs listens for messages from a specific WebSocket client

func ListenForWs(conn *websocket.Conn) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Recovered from error: %v", r)
		}
	}()

	var payload WsPayload
	for {
		if err := conn.ReadJSON(&payload); err != nil {
			log.Println("Error reading WebSocket message:", err)
			return
		}
		payload.Conn = conn
		hub.Channel <- payload
	}
}

// renderPage renders a template with the given data

func renderPage(w http.ResponseWriter, tmpl string, data jet.VarMap) error {
	view, err := views.GetTemplate(tmpl)
	if err != nil {
		log.Println("Error loading template:", err)
		return err
	}

	if err := view.Execute(w, data, nil); err != nil {
		log.Println("Error executing template:", err)
		return err
	}
	return nil
}
