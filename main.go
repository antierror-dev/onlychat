package main

import (
"encoding/base64"
"log"
"net/http"

socketio "github.com/googollee/go-socket.io"
"github.com/googollee/go-socket.io/engineio"
"github.com/googollee/go-socket.io/engineio/transport"
"github.com/googollee/go-socket.io/engineio/transport/polling"
"github.com/googollee/go-socket.io/engineio/transport/websocket"

)

func decodeBase64(encodedString string) (string, error) {
decodedBytes, err := base64.StdEncoding.DecodeString(encodedString)
if err != nil {
return "", err
}
return string(decodedBytes), nil
}

type User struct {
Username string
Room     string
}

var users = make(map[socketio.Conn]*User)

var allowOriginFunc = func(r *http.Request) bool {
	  return true
  }

func main() {

server := socketio.NewServer(&engineio.Options{
		  Transports: []transport.Transport{
		    &polling.Transport{
				  CheckOrigin: allowOriginFunc,
			  },
			  &websocket.Transport{
				  CheckOrigin: allowOriginFunc,
			  },
		  },
	 })

server.OnConnect("/", func(s socketio.Conn) error {
users[s] = &User{}
log.Println("Connected:", s.ID())
return nil
})

server.OnEvent("/", "join", func(s socketio.Conn, room string, username string) {
s.Join(room)
users[s].Room = room
users[s].Username = username
server.BroadcastToRoom("", room, "alert", username+" join in the chat!")
log.Println(username + " Joined room: " + room)
})

server.OnEvent("/", "leave", func(s socketio.Conn) {
user, ok := users[s]
if !ok {
log.Println("User not found")
return
}
server.BroadcastToRoom("", user.Room, "alert", user.Username+" has left the chat")
s.Leave(user.Room)
delete(users, s)
})




  server.OnEvent("/", "chat-message", func(s socketio.Conn, msg string) {
    user, ok := users[s]
    if !ok {return}
    dec_msg,err := decodeBase64(msg)
    if err!=nil{
      log.Println("Errot decoding message")
      return
    }
    server.BroadcastToRoom("", user.Room, "chat-message",user.Username+": "+dec_msg,s.ID())
    
  })



server.OnError("/", func(s socketio.Conn, e error) {
log.Println("Error:", e)
})

server.OnDisconnect("/", func(s socketio.Conn, reason string) {
user, ok := users[s]
if ok {
log.Println("Disconnected:", user.Username,"\nReason:",reason)
delete(users, s)
}
})

go func() {
if err := server.Serve(); err != nil {
log.Fatalf("socketio listen error: %s\n", err)
}
}()
defer server.Close()

http.Handle("/socket.io/", server)
http.Handle("/", http.FileServer(http.Dir("./public")))
http.HandleFunc("/chat-room", BackToLogin)

log.Println("Serving at localhost:8080...")
log.Fatal(http.ListenAndServe(":8080", nil))
}

func BackToLogin(w http.ResponseWriter, r *http.Request) {
    http.Redirect(w, r, "/", http.StatusMovedPermanently)
}
