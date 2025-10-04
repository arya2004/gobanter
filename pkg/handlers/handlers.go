package handlers

import (
	"fmt"
	"log"
	"net/http"
	"sort"

	"github.com/CloudyKit/jet/v6"
	"github.com/gorilla/websocket"
)

// Global variables for managing WebSocket connections
var (
	wsChannel = make(chan WsPayload)             // Channel for handling WebSocket messages
	clients   = make(map[*websocket.Conn]string) // Map of connected clients and their usernames
	views     = jet.NewSet(                      // Template engine configuration
		jet.NewOSFileSystemLoader("./templates"),
		jet.InDevelopmentMode(),
	)
	upgradeConnection = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin:     func(r *http.Request) bool { return true }, // Allow all origins
	}
)

// Home renders the home page of the application
func Home(w http.ResponseWriter, r *http.Request) {
	log.Println("Rendering home page")
	err := renderPage(w, "home.html", nil)
	if err != nil {
		log.Println("Error rendering home page:", err)
	}
}

// WsJsonResponse represents the JSON structure for WebSocket responses
// added "from" and "to" for private messages
type WsJsonResponse struct {
	Action         string   `json:"action"`
	Message        string   `json:"message"`
	MessageType    string   `json:"message_type"`
	ConnectedUsers []string `json:"connected_users"`
	From           string   `json:"from"`
	To             string   `json:"to"`
}

// WsPayload represents the payload received from WebSocket clients
// added "to" for private messages
type WsPayload struct {
	Action   string          `json:"action"`
	Username string          `json:"username"`
	Message  string          `json:"message"`
	To       string          `json:"to"`
	Conn     *websocket.Conn `json:"-"` // Exclude from JSON serialization
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

	// Initial connection response
	response := WsJsonResponse{
		Message: "<em><small>Connected to server</small></em>",
	}

	clients[ws] = "" // Add the new connection to clients map

	err = ws.WriteJSON(response)
	if err != nil {
		log.Println("Error writing JSON response:", err)
		return
	}

	// Start listening for messages from this client
	go ListenForWs(ws)
}

// ListenToWsChannel listens for messages on the WebSocket channel and handles them
func ListenToWsChannel() {
	for {
		e := <-wsChannel
		var response WsJsonResponse

		switch e.Action {
		case "username":
			// Handle username assignment
			clients[e.Conn] = e.Username
			response.Action = "list_users"
			response.ConnectedUsers = getUserList()
			broadcastToAll(response)

		case "left":
			// Handle client disconnection
			response.Action = "list_users"
			delete(clients, e.Conn)
			response.ConnectedUsers = getUserList()
			broadcastToAll(response)

		case "broadcast":
			// Handle broadcast messages
			response.Action = "broadcast"
			response.Message = fmt.Sprintf("<strong>%s</strong>: %s", e.Username, e.Message)
			broadcastToAll(response)

		case "private":
			// handle private messages
			handlePrivateMessage(e)
		}
	}
}

//function handlePrivateMessage sends a message to a specific user

func handlePrivateMessage(payload WsPayload) {

	// finding the recipient connection
	var recipientConn *websocket.Conn
	for conn, username := range clients {
		if username == payload.To {
			recipientConn = conn
			break
		}
	}

	// if recipient not found , then send an error message to sender

	if recipientConn == nil {
		errorResponse := WsJsonResponse{
			Action:      "error",
			Message:     fmt.Sprintf("User '%s' not found or offline", payload.To),
			MessageType: "error",
		}

		if err := payload.Conn.WriteJSON(errorResponse); err != nil {
			log.Printf("Error sending error message to sender: %v", err)
		}
		return
	}

	// creating a private message response

	response := WsJsonResponse{
		Action:      "private",
		Message:     payload.Message,
		MessageType: "private",
		From:        payload.Username,
		To:          payload.To,
	}

	// this is send to recipient

	if err := recipientConn.WriteJSON(response); err != nil {
		log.Printf("Error sending private message to recipient: %v", err)
	} else {
		log.Printf("Private message sent from %s to %s", payload.Username, payload.To)
	}

	// echo back to the sender (for confirmation on the Ui side)

	if err := payload.Conn.WriteJSON(response); err != nil {
		log.Printf("Error sending confirmation to sender: %v", err)
	}
}

// getUserList returns a sorted list of connected usernames
func getUserList() []string {
	var userList []string
	for _, username := range clients {
		if username != "" {
			userList = append(userList, username)
		}
	}
	sort.Strings(userList)
	return userList
}

// broadcastToAll sends a WebSocket response to all connected clients
func broadcastToAll(response WsJsonResponse) {
	for client := range clients {
		if err := client.WriteJSON(response); err != nil {
			log.Printf("Error sending message to client: %v", err)
			_ = client.Close()
			delete(clients, client)
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

		// Assign connection to payload and send it to the channel
		payload.Conn = conn
		wsChannel <- payload
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
