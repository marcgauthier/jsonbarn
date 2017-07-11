/*Package models - hub.go

This file contain the functions to handle the requests that come from the
Websocket connections, it also check for the broadcast queue and send
message that need to be broadcasted to all connected users.

______________________________________________________________________________

 Ecureuil - Web framework for real-time javascript app.
_____________________________________________________________________________

MIT License

Copyright (c) 2014-2016 Marc Gauthier

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.

______________________________________________________________________________________


Revision:
    01 Nov 2016 - Clean code, audit. we don't have to monitor thinks all the time


______________________________________________________________________________

*/
package models

import (
	"encoding/json"
	"strconv"
	"time"

	"github.com/antigloss/go/logger"
	"github.com/gorilla/websocket"
)

const websocketBufferSize = 8192

/*MsgClientCmd Structure use by the client to send command to the backend.
 */
type MsgClientCmd struct {
	Action      string          `json:"action"`      // action LOGIN, READ, DELETE, UPDATE
	Username    string          `json:"username"`    // username to check if user has access to the command
	Password    string          `json:"password"`    // password to check if user has access to the command
	Bucketname  string          `json:"bucketname"`  // bucket name is the table that need to be update.
	SearchField string          `json:"searchfield"` // Key to insert,edit or query data.
	Key         string          `json:"key"`         // Key to insert,edit or query data.
	MaxKey      string          `json:"maxkey"`      //
	Field       string          `json:"field"`       // use with Key parameter for  FindOne and FindMany to get the data.
	Defered     uint64          `json:"defered"`     // execute command at a later date
	Data        json.RawMessage `json:"data"`        // contain the JSON serialized object to be saved, it will be HTML Sanitized
}

/*Hub Structure to manage Hub ressources.
 */
type Hub struct {
	clients      map[*Client]bool // list of client in the hub
	broadcast    chan []byte      // broadcast channel
	addClient    chan *Client     // func to add client in the hub
	removeClient chan *Client     // func to remove client from the hub
}

/*hub initialize a new hub
 */
var hub = Hub{
	broadcast:    make(chan []byte),
	addClient:    make(chan *Client),
	removeClient: make(chan *Client),
	clients:      make(map[*Client]bool),
}

/*getBucket extract the name of the bucket from the message and return the
message
*/
func getBucket(s []byte) (string, []byte) {
	bucket := ""
	for i := 0; i < len(s); i++ {
		if s[i] == byte(':') {
			return bucket, s[i+1:]
		}
		bucket += string(s[i])
	}
	return "", s
}

/*start hub and Runs forever as a goroutine
 */
func (hub *Hub) start() {
	for {
		// one of these fires when a channel
		// receives data
		select {
		case conn := <-hub.addClient:
			// add a new client
			hub.clients[conn] = true
		case conn := <-hub.removeClient:
			// remove a client
			if _, ok := hub.clients[conn]; ok {
				delete(hub.clients, conn)
				close(conn.send)
			}
		case message := <-hub.broadcast:
			// broadcast a message to all clients that have register to the bucket "EVENTNAME"
			for conn := range hub.clients {

				b, msg := getBucket(message)

				if b == "" || IsStrInArray(b, conn.registerEvents) {

					select {
					case conn.send <- msg:
					default:
						// oups! no message sent
						// pause 1/100 of second and try again
						time.Sleep(10 * time.Millisecond)

						// send blocking!
						conn.send <- msg
					}
				}

			}
		}
	}
}

/*HubStart Start the functions that monitor for activities.
 */
func HubStart() {

	go hub.start()
	go checkforBroadcast()

}

/*checkforBroadcast Check if we need to send a broadcast read object from the broadcast
queue and if item are present push the object in the hub channel.
*/
func checkforBroadcast() {

	for {

		// send all messages until the queue is empty
		var msg []byte

		for {
			if msg = BroadcastGet(); msg == nil {
				break
			}
			hub.broadcast <- msg
		}

		// queue is empty take a 1/4 sec pause.
		time.Sleep(250 * time.Millisecond)

		// infinite loop until program end.
	}

}

/*ClientAdd A client just connected to the website and opened a websocket connection
add the client in the list so we can provide broadcast to this user.
*/
func ClientAdd(conn *websocket.Conn) {

	// create client struct

	client := &Client{
		ws:             conn,
		send:           make(chan []byte, websocketBufferSize),
		registerEvents: []string{},
		password:       "",
		username:       "",
	}

	// add client in the hub

	hub.addClient <- client

	// each client have a write and read concurrent function.

	go client.write()
	go client.read()

	return
}

/*Client Start of CLIENT -----------------------------------------------------
 */
type Client struct {
	ws *websocket.Conn
	// Hub passes broadcast messages to this channel
	send           chan []byte
	registerEvents []string
	username       string
	password       string
	LoginAttempts  []uint64 // contain the time when login attempt was made.
}

/*ClearLoginAttempt remove the login attempt that are older than 1 minutes and return
how many attempt have been made in the last minute
*/
func (c *Client) ClearLoginAttempt() int {

	t := uint64(time.Now().Add(-60 * time.Second).UTC().Unix())
	i := 0
	count := 0
	for {
		if i >= len(c.LoginAttempts) {
			break
		}
		if c.LoginAttempts[i] < t {
			// delete
			c.LoginAttempts = removeIndex(c.LoginAttempts, i)
		} else {
			count++
			i++
		}
	}
	return count
}

/*write Hub broadcasts a new message and this fires
 */
func (c *Client) write() {
	// make sure to close the connection incase the loop exits
	defer func() {
		c.ws.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:

			if !ok {

				logger.Trace("channel send error closed: " + string(message))

				//c.ws.WriteMessage(websocket.CloseMessage, []byte{})
				//Data        json.RawMessage `json:"data"`       // contain the JSON serialized object to be saved, it will be HTML Sanitized

				return

			} else {

				logger.Trace(string(message))

				if message != nil && len(message) > 0 {
					c.ws.WriteMessage(websocket.TextMessage, message)
				}
			}

		}
	}
}

/*New message received from user over websocket process the message
 */
func (c *Client) read() {
	defer func() {
		hub.removeClient <- c
		c.ws.Close()
	}()

	for {

		Msgtype, message, err := c.ws.ReadMessage()

		if err != nil {

			logger.Warn("client.read websocket error: " + err.Error())
			// user as probably disconnected!!! from from the Hub List and terminate Go function
			break
			// break will exit the loop and execute defer removeclient+close

		}

		/*
		   Packet of information are sent using msgClientCmd serializaed JSON
		*/

		packet := MsgClientCmd{}

		// show received message
		logger.Trace("rx from frontend: " + string(message))

		if Msgtype == websocket.BinaryMessage {
			err = json.Unmarshal(message, &packet)
		} else {
			err = json.Unmarshal(message, &packet)
		}

		if err != nil {

			// command sent is not a valid JSON send a warning to the user.
			// do not log returned error, because the error is cause by the frontend not the backend
			// c.ws.WriteMessage(websocket.TextMessage, PrepMessageForUser("JSON OBJECT provided was invalid: "+SanitizeStrHTML(err.Error())))
			c.send <- PrepMessageForUser("JSON OBJECT provided was invalid: " + SanitizeStrHTML(err.Error()))

		} else {

			// Here we have a valid JSON object check if we can do something with it!

			var err error
			var user []byte

			// first check if user defer execution of this command

			if packet.Action == "LOGIN" {

				c.LoginAttempts = append(c.LoginAttempts, uint64(time.Now().UTC().Unix()))

				count := c.ClearLoginAttempt()
				if count > Configuration.LoginPerMin {
					user = PrepMessageForUser("You have exceeded the maximum number of login attempt, try again in 1 min!")
				} else {

					user, err = DBLogin(&packet)

					if err == nil {
						// store crendential for the duration of the websocket connection
						c.password = packet.Password
						c.username = packet.Username
						logger.Info("User " + c.username + " as logged in on this websocket!")
					}
					// if credential are already loaded they are not lost by doing a bad request!

				}

			} else if packet.Action == "LOGOUT" {

				c.password = ""
				c.username = ""
				err = nil
				user = []byte("{ \"action\":\"logout\"}")

			} else if packet.Action == "QUERY" || packet.Action == "READALL" || packet.Action == "READONE" || packet.Action == "READFIND" || packet.Action == "READRANGE" {

				/*
				   Request range of information from a bucket should contain:
				   bucketname, startdate, enddate, username and password

				   Username and password is optional depending on the type of information requested!

				   if startdate and enddate is null default range will be selected for some item such as open Bulletin the default range is
				   all items in the Bucket!
				*/

				// overwrite any provided credential with the proper credential
				packet.Username = c.username
				packet.Password = c.password

				user, err = DBRead(&packet)

			} else if packet.Action == "LOGS" {

				/*
				   Request range of information from a bucket should contain:
				   bucketname, startdate, enddate, username and password

				   Username and password is optional depending on the type of information requested!

				   if startdate and enddate is null default range will be selected for some item such as open Bulletin the default range is
				   all items in the Bucket!
				*/

				// overwrite any provided credential with the proper credential
				packet.Username = c.username
				packet.Password = c.password

				user, err = DBGetLogs(&packet)

			} else if packet.Action == "UPDATE" {

				/*
				   Request Update information in the database  should contain:
				   bucketname, username and password
				   if key is not present or empty this is an insert else it is an update.

				   if the update is valid a broadcast will be sent

				*/

				// overwrite any provided credential with the proper credential
				packet.Username = c.username
				packet.Password = c.password
				user, err = DBUpdate(&packet, false)

			} else if packet.Action == "SETUSERSETTING" {

				/*
				   Request to Update a single property of a specific item.
				   itemID is in key
				   json object in data

				*/

				// overwrite any provided credential with the proper credential
				packet.Username = c.username
				packet.Password = c.password
				user, err = DBUserSettings(&packet)

			} else if packet.Action == "INSERT" {

				/*
				   Request Update information in the database  should contain:
				   bucketname, username and password
				   if key is not present or empty this is an insert else it is an update.

				   if the update is valid a broadcast will be sent

				*/

				// overwrite any provided credential with the proper credential
				packet.Username = c.username
				packet.Password = c.password
				user, err = DBInsert(&packet, false)

			} else if packet.Action == "REGISTEREVENT" {

				// overwrite any provided credential with the proper credential
				packet.Username = c.username
				packet.Password = c.password
				user, err = registerEvent(c, &packet)

			} else if packet.Action == "UNREGISTEREVENT" {

				// overwrite any provided credential with the proper credential
				packet.Username = c.username
				packet.Password = c.password
				user, err = unregisterEvent(c, &packet)

			} else if packet.Action == "GETTIME" {

				user = GetTime()

			} else if packet.Action == "GETCONFIG" {

				packet.Username = c.username
				packet.Password = c.password
				user, err = GetConfiguration(&packet)

			} else if packet.Action == "GETUSERS" {

				packet.Username = c.username
				packet.Password = c.password
				user, err = GetUsers(&packet)

			} else if packet.Action == "PUTCONFIG" {

				packet.Username = c.username
				packet.Password = c.password
				user, err = PutConfiguration(&packet)

			} else if packet.Action == "INDEXDROP" {

				// overwrite any provided credential with the proper credential
				packet.Username = c.username
				packet.Password = c.password
				user, err = DBDropIndex(&packet)

			} else if packet.Action == "INDEXCREATE" {

				// overwrite any provided credential with the proper credential
				packet.Username = c.username
				packet.Password = c.password
				user, err = DBCreateIndex(&packet)

			} else if packet.Action == "INDEXLIST" {

				// overwrite any provided credential with the proper credential
				packet.Username = c.username
				packet.Password = c.password
				user, err = DBListIndex(&packet)

			} else if packet.Action == "EMAILALERT" {

				// overwrite any provided credential with the proper credential
				packet.Username = c.username
				packet.Password = c.password
				user, err = ReceiveEmailAlertChangeReq(&packet)

			} else if packet.Action == "DELETE" {

				/*
				   Request to delete an item from the database should contain:
				   bucketname, key, username and password

				   if the delete is valid a broadcast will be sent

				*/

				// overwrite any provided credential with the proper credential
				packet.Username = c.username
				packet.Password = c.password
				user, err = DBDelete(&packet, false)

			} else {

				/*
				   Received Invalid command just ignore it....
				*/

				user = nil
				err = nil

			}

			/*
			   User contain message or command to be sent to user.
			   info contain information that should be logged.
			   err contain any error that occur during the process of the command.

			*/

			if user != nil {
				logger.Trace("Sending: " + string(user))
				c.send <- user
			}

			if err != nil {
				logger.Error(err.Error())
			}

		}

	} // infinite for loop

}

/*GetTime return the time on the server
 */
func GetTime() []byte {

	t := UnixUTCSecs()

	return []byte("{\"action\": \"gettime\", \"time\":" + strconv.FormatFloat(t, 'f', 6, 64) + "}")

}

/*registerEvent request to be sent all event that occur in a specific bucket,
i.e. update, insert, delete
This function does not care if event is already register it simply add one
item in the list.
*/
func registerEvent(c *Client, packet *MsgClientCmd) ([]byte, error) {

	logger.Trace("Req register event for " + packet.Bucketname + " from " + packet.Username)

	// Check if the user has rights
	access, err := UserHasRight([]byte(packet.Username), []byte(packet.Password), packet.Bucketname+"-read")
	if err != nil {
		logger.Error("Register event " + packet.Username + "  for " + packet.Bucketname + " error: " + err.Error())
		return []byte("{\"action\": \"registerevent\", \"bucketname\":\"" + EscDoubleQuote(packet.Bucketname) + "\", \"status\":false, \"error\":\"" + err.Error() + "\" }"), nil
	}

	if access == false {
		logger.Warn("Access denied: User " + packet.Username + " register event for " + packet.Bucketname + " error: " + err.Error())
		return []byte("{\"action\": \"registerevent\", \"bucketname\":\"" + EscDoubleQuote(packet.Bucketname) + "\", \"status\":false, \"error\":\"access denied\" }"), nil
	}

	logger.Trace("Registering Event for " + packet.Bucketname + " from " + packet.Username)

	c.registerEvents = append(c.registerEvents, packet.Bucketname)

	return []byte("{\"action\": \"registerevent\", \"bucketname\":\"" + EscDoubleQuote(packet.Bucketname) + "\", \"status\":true}"), nil
}

/*unregisterEvent request to no longer receive event for a specific bucket.
 */
func unregisterEvent(c *Client, packet *MsgClientCmd) ([]byte, error) {

	logger.Trace("Req to unregister event for " + packet.Bucketname + " from " + packet.Username)

	// Check if the user has rights
	access, err := UserHasRight([]byte(packet.Username), []byte(packet.Password), packet.Bucketname+"-read")
	if err != nil {
		logger.Error("Unregister event " + packet.Username + "  for " + packet.Bucketname + " error: " + err.Error())
		return []byte("{\"action\": \"unregisterevent\", \"bucketname\":\"" + EscDoubleQuote(packet.Bucketname) + "\", \"status\":false, \"error\":\"" + err.Error() + "\" }"), nil
	}

	if access == false {
		logger.Warn("Access denied: User " + packet.Username + " register event for " + packet.Bucketname + " error: " + err.Error())
		return []byte("{\"action\": \"unregisterevent\", \"bucketname\":\"" + EscDoubleQuote(packet.Bucketname) + "\", \"status\":false, \"error\":\"access denied\" }"), nil
	}

	logger.Trace("Unregistering Event for " + packet.Bucketname + " for " + packet.Username)

	for i := 0; i < len(c.registerEvents); i++ {
		if c.registerEvents[i] == packet.Bucketname {
			c.registerEvents = append(c.registerEvents[:i], c.registerEvents[i+1:]...)
			break
		}
	}
	return []byte("{\"action\": \"unregisterevent\", \"bucketname\":\"" + EscDoubleQuote(packet.Bucketname) + "\", \"status\":true}"), nil
}
