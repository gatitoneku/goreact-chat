package main

import (
"log"
"net/http"

"github.com/gorilla/websocket"
)

//List of clients, mapped to the connection status (true/false)
var clients = make(map[*websocket.Conn]bool)
//a channel that takes messages
var broadcast = make(chan Message)

//Upgrader to WebSocket requests
var upgrader = websocket.Upgrader{}

type Message struct {
        Email    string `json:"email"`
        Username string `json:"username"`
        Message  string `json:"message"`
}

func handleConnections(w http.ResponseWriter, r *http.Request){
    //upgrade to a Websocket request (ws)
    ws, err := upgrader.Upgrade(w, r, nil)
    if err != nil {
        log.Fatal(err)
    }
    
    defer ws.Close()
    clients[ws] = true

    for {
        var msg Message
        
        //Read JSON message in ws and store it in a Message interface
        err := ws.ReadJSON(&msg)
        if err != nil {
            //Delete clients [ws]
            log.Printf("error: %v", err)
            delete(clients, ws)
            break
        }
        
        //Send the received message to broadcast channel
        broadcast <- msg
    }
}

func handleMessages() {
    for {
        msg := <- broadcast //lets change this to other variable name later
            
        //send it to all clients connected
        for client := range clients {
            err := client.WriteJSON(msg)
            if err != nil {
                log.Printf("error: %v", err)
                client.Close()
                delete(clients, client)
            }  
        }
    }
}

func main() {
    fs := http.FileServer(http.Dir("../public"))
    //Routing
    http.Handle("/", fs)
    http.HandleFunc("/ws", handleConnections)
        
    go handleMessages()
        
    //Listen for connections
    log.Println("http server started on :8000")
    err := http.ListenAndServe(":8000", nil)
    if err != nil {
        log.Fatal("ListenAndServe:", err)
    }
}