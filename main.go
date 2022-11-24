package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/matoous/go-nanoid/v2"
)

type Client struct {
	Name string `json:"name"`
	Conn *websocket.Conn `json:"web_socket"`
	RoomCode string `json:"room_code"`
	Host bool `json:"host"` // host (owner or not)
}

type ChatServer struct {
	RoomCode string `json:"room_code"`
	Clients []*Client `json:"clients"`
}

type Response struct {
	Status int `json:"status"`
	Msg string `json:"message"`
}


// array of chat servers
var servers []*ChatServer

var upgrader = websocket.Upgrader {}

// creates unique roomID
func randomConnectionID() string {
	id, err := gonanoid.New()

	if err!= nil {
		log.Panic(err)
	}

	return id
}

// createRoom for name!
func CreatRoomHandler(w http.ResponseWriter,r *http.Request ) {
	name := r.URL.Query()["name"]
	con, err := upgrader.Upgrade(w,r,nil)
	if len(name) != 0 {

  name := name[0]
	if err != nil {
		log.Println(err)
		return;
	}

	roomCode := randomConnectionID();

	createServer := &ChatServer{RoomCode: roomCode}
	createClient := &Client{Name: name,Conn: con,RoomCode:roomCode, Host: true}
	createServer.Clients = append(createServer.Clients, createClient )
	servers = append(servers,createServer);

	writeMessage(con,&Response{Status: 200,Msg: roomCode})

	createClient.readLoop()


  } else {
		writeMessage(con,&Response{Status: 404,Msg: "NAME MISSING"})
	}
}

// checks if RoomCode exists in list of chat servers
func isRoomCodeAvailable(roomCode string) bool {
	for i:=0;i<len(servers);i++ {
		if roomCode == servers[i].RoomCode {
			return true;
		}
	}
	return false;
}

// checks if name is already available in server or not
func isNameAvailableInServer(name string,clients []*Client,len int) bool {
	for i:=0;i<len;i++ {
		if name == clients[i].Name {
			return true
		}
	}
	return false;
}

// gets server by room code
func getServerByRoomCode(roomCode string) *ChatServer {

	var currentServer *ChatServer

	for i:=0;i<len(servers);i++ {
		if(servers[i].RoomCode == roomCode) {
			currentServer = servers[i];
		}
	}
	return currentServer;
}


// handles Join ROOM
func JoinRoomHandler(w http.ResponseWriter,r *http.Request) {
	roomCode := r.URL.Query()["room"][0]
	name := r.URL.Query()["name"][0]

	con, _ := upgrader.Upgrade(w,r,nil)

	// CHECKS IF ROOM CODE AVAIABLE
	if isRoomCodeAvailable(roomCode) {

		// gets current server to check if name already exists in a server or not..
		var currentServer *ChatServer = getServerByRoomCode(roomCode);

		if !isNameAvailableInServer(name,currentServer.Clients,len(currentServer.Clients)) {

			// append new client to the server
		createClient := &Client{Name: name,Conn: con,RoomCode:roomCode, Host: false}

		for i:=0;i<len(servers);i++ {
			if roomCode == servers[i].RoomCode {
				servers[i].Clients = append(servers[i].Clients,createClient);
				break;
			}
		}

		// broadcast msgs to all clients
		go createClient.readLoop()

	  } else {
			writeMessage(con,&Response{Status:404,Msg:"DUPLICATE NAME"})
			con.Close()
		}
	} else {
		writeMessage(con,&Response{Status:404,Msg:"NOT FOUND ROOM CODE"})
		con.Close()
	}
}

// helper to send msg to client
func writeMessage(c *websocket.Conn,resp *Response) {
	w, _ := c.NextWriter(1);

	json.NewEncoder(w).Encode(&resp)

	w.Close()
}

// broadcasts msg to all other clients in a server!
func(c *Client) readLoop() {
	for {
		   _, str, err := c.Conn.NextReader();

			 if err != nil {
				c.Conn.Close()
				break
			 }

			 buf:= new(bytes.Buffer)

			 buf.ReadFrom(str)
			 msgToBroadCast := buf.String()

			 for i:=0;i<len(servers);i++ {
					clients := servers[i].Clients;
					for j:=0;j<len(clients);j++ {
						  if(c.Name != clients[j].Name && c.RoomCode == clients[j].RoomCode) {
								writeMessage(clients[j].Conn,&Response{200,msgToBroadCast})
							}
					}
				}
		}
}

// delete the room (ONLY HOST CAN DELETE)
func DeleteRoomHandler(w http.ResponseWriter,r *http.Request) {
	roomCode := r.URL.Query()["room"][0]
	name := r.URL.Query()["name"][0]

	con, _ := upgrader.Upgrade(w,r,nil)

	if isRoomCodeAvailable(roomCode) {

		if isNameHostFortheServer(name,roomCode) {

			var currentServer *ChatServer = getServerByRoomCode(roomCode)
			var temp []*ChatServer

			temp = make([]*ChatServer, 0)

			for i:=0;i<len(servers);i++ {
				if(currentServer.RoomCode != servers[i].RoomCode) {
					temp = append(temp,servers[i]);
				}
			}


			writeMessage(con,&Response{Status:200,Msg:"DELETED ROOM SUCCESSFUL"})
			closeConnectionForClients(currentServer)

			servers = make([]*ChatServer, 0)
			servers = temp;

		} else {
			writeMessage(con,&Response{Status:404,Msg:"HOST CAN ONLY DELETE ROOM"})
			con.Close()
		}

	} else {
		writeMessage(con,&Response{Status:404,Msg:"NOT FOUND ROOM CODE"})
		con.Close()
	}
}

func closeConnectionForClients(server *ChatServer) {
	for i:=0;i<len(server.Clients);i++ {
		server.Clients[i].Conn.Close()
	}
}

func isNameHostFortheServer(name string,roomCode string) bool {
	var currentServer *ChatServer = getServerByRoomCode(roomCode)

	for i:=0;i<len(currentServer.Clients);i++ {
		if currentServer.Clients[i].Name == name && currentServer.Clients[i].Host{
			return true;
		}
	}

	return false
}



func main() {

	http.HandleFunc("/",CreatRoomHandler)
	http.HandleFunc("/join",JoinRoomHandler)
	http.HandleFunc("/delete",DeleteRoomHandler)

	fmt.Println("@port : 8080")
	http.ListenAndServe(":8080",nil)
}