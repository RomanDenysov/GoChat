package main

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)
type Client struct {
    conn     *websocket.Conn // Соединение WebSocket.
    username string          // Имя пользователя.
}

var clients = make(map[*Client]bool)
var broadcast = make(chan []byte)

var upgrader = websocket.Upgrader{
	CheckOrigin: func (r *http.Request) bool  {
		return true
	},
}

func main() {
	 // Установка обработчика WebSocket
	 http.HandleFunc("/ws", handleConnections)
	 go handleMessages()
 
	 // Настройка сервера для обслуживания статических файлов из папки сборки React
	//  fs := http.FileServer(http.Dir("dist"))
	//  http.Handle("/", fs)
 
	 // Запуск HTTP сервера
	 log.Println("HTTP server started on :8080")
	 err := http.ListenAndServe(":8080", nil)
	 if err != nil {
		 log.Fatal("ListenAndServe: ", err)
	 }
 }

 func handleConnections(w http.ResponseWriter, r *http.Request) {
    ws, err := upgrader.Upgrade(w, r, nil)
    if err != nil {
        log.Fatal(err)
    }
    defer ws.Close()

    ws.WriteMessage(websocket.TextMessage, []byte("Введите ваше имя пользователя:"))
    _, usernameMsg, err := ws.ReadMessage()
    if err != nil {
        log.Println("Error reading username:", err)
        return
    }
    username := string(usernameMsg)
	log.Println("Username:", username)
	
    client := &Client{conn: ws, username: username}
    clients[client] = true

    for {
        _, msg, err := ws.ReadMessage()
        if err != nil {
            log.Printf("error: %v", err)
            delete(clients, client)
            break
        }
        broadcast <- msg
    }
}

func handleMessages() {
    for {
        msg := <-broadcast

        for client := range clients {
            // Здесь мы обращаемся к `conn`, которое является частью структуры `Client`
            err := client.conn.WriteMessage(websocket.TextMessage, msg)
            if err != nil {
                log.Printf("error: %v", err)
                client.conn.Close() // Закрываем соединение
                delete(clients, client)
            }
        }
    }
}