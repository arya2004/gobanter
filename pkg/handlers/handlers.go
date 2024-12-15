package handlers

import (
	"fmt"
	"log"
	"net/http"
	"sort"

	"github.com/CloudyKit/jet/v6"
	"github.com/gorilla/websocket"
)

var wsChannel = make(chan WsPayload)

var clients = make(map[*websocket.Conn]string)


var views =	jet.NewSet(
	jet.NewOSFileSystemLoader("./templates"),
	jet.InDevelopmentMode(),

)	

var upgradeConnection = websocket.Upgrader{
	ReadBufferSize: 1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r* http.Request) bool {return true},
}


func Home(w http.ResponseWriter, r *http.Request) {
	log.Println("called home")
	err := renderPage(w, "home.html", nil)
	if err != nil{
		log.Println(err)
	
	}
}



type WsJsonResponse struct {
	Action string `json:"action"`
	Message string `json:"message"`
	MessageType string `json:"message_type"`
	ConnectedUsers []string `json:"connected_users"`
}

type WsPayload struct {
	Action string `json:"action"`
	Username string `json:"username"`
	Message string `json:"message"`
	Conn *websocket.Conn `json:"-"`
}


func WsEndpoint(w http.ResponseWriter, r* http.Request){

	log.Println("client connected to endpoint")

	ws,err := upgradeConnection.Upgrade(w,r, nil)
	if err != nil {
		log.Panicln(err)
	}

	log.Println("client connected to endpoint")
	var response WsJsonResponse
	
	response.Message = `<em><small>connected to server</small></em>`

	clients[ws] = ""


	err = ws.WriteJSON(response)
	
	if err != nil {
		log.Panicln(err)
	}

	go ListenForWs(ws)
}

func ListenToWsChannel(){
	var response WsJsonResponse
	
	for {
		e := <- wsChannel
		switch e.Action {
		case "username":
			clients[e.Conn] = e.Username
			users := getUserList()
			response.Action = "list_users"
			response.ConnectedUsers = users
			broadcastToAll(response)

		case "left":
			response.Action = "list_users"
			delete(clients, e.Conn)
			users := getUserList()
			response.ConnectedUsers = users
			broadcastToAll(response)

		case "broadcast":
			response.Action = "broadcast"
			response.Message = fmt.Sprintf("<strong>%s</strong>: %s", e.Username, e.Message)
			broadcastToAll(response)

		}
	}
}


func getUserList() []string{
	var userList []string
	
	for _,x := range clients{
		if(x != ""){
		userList = append(userList, x)
		}
	}
	sort.Strings(userList)
	return userList
}

func broadcastToAll(response WsJsonResponse){
	for client := range clients {
		err := client.WriteJSON(response)
		if err != nil {
			log.Printf("error: %v", err)
			_ = client.Close()
			delete(clients, client)
		}
	}
}


func ListenForWs(conn *websocket.Conn){
	defer func(){
		if r := recover(); r != nil {
			log.Printf("error %v", r)
		}
	}()

	var payload WsPayload
	for {
		
		err := conn.ReadJSON(&payload)
		if err != nil {
			log.Println(err)
			return
		}else{
			payload.Conn = conn
			wsChannel <- payload
		}

		
	}
}

func renderPage(w http.ResponseWriter, tmpl string, data jet.VarMap) error {
	view, err := views.GetTemplate(tmpl)
	if err != nil{
		log.Println(err)
		return err
	}

	err = view.Execute(w, data, nil)
	if err != nil{
		log.Println(err)
		return err
	}
	return nil
}
