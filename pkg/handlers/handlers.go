package handlers

import (
	"log"
	"net/http"

	"github.com/CloudyKit/jet/v6"
	"github.com/gorilla/websocket"
)

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
	err = ws.WriteJSON(response)
	
	if err != nil {
		log.Panicln(err)
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
