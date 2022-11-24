# websocket-chat-server

## Chat Server Supports

   - Create Room
   - Join Room 
   - Delete Room
   
   
### Create Room 
   - &name (query param) to create room (Owner)
   ```localhost:8080/?name=host``` 
   returns roomCode
   
### Join Room 

   - Join Room by Room Code and Name 
   ```localhost:8080/join?room=sYmQDdqZZuM1DtjPf_51U&name=mystica```
   
### Delete Room
   - Delete Room by Room Code and Name
   - Only Host Can delete the room
   ```localhost:8080/delete?room=sYmQDdqZZuM1DtjPf_51U&name=host```
