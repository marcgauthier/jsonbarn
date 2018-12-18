/*

This package role is to connect to a JSONBARN server, a secure websocket connection
is eastablish and the client can listen to specific buckets 

*/

package jsonbarn

import (
	"crypto/tls"
	"github.com/gorilla/websocket"
	"github.com/tidwall/gjson"
	"net/url"
	"fmt"
	"errors"
	"time"
)


/* This Public variable can be changed to show details of operation on the console for debuging
*/
var ShowTrace bool = true 

/* Size of the buffer for the receive channel, the minimum size is 128, default size is 2048 
*/
var Bufsize int = 2048


/* JsonBarn object 
*/
type JsonBarn struct {
	c         *websocket.Conn
	Ch		  chan []byte		//receive channel
	exit      bool
	connected bool
	loggedIn  bool
	showtrace bool
	NewDialer *websocket.Dialer
}


/* function to show debug info 
*/
func trace(msg string) {
	if ShowTrace && len(msg) > 0 {
		fmt.Println(msg)
	}
}

/* create new item
*/
func New() JsonBarn {
	if Bufsize <= 128 {
		Bufsize = 128
	}
	return JsonBarn{c: nil, exit: false, Ch: make(chan []byte, Bufsize), connected: false, loggedIn: false, NewDialer: &websocket.Dialer{}}
}



func (j *JsonBarn) Send(msg string) error {
	if j==nil {
		return errors.New("JsonBarnIsNil")
	}
	if !j.loggedIn {
		return errors.New("NotLoggedIn")
	}
	if !j.connected {
		return errors.New("NotConnected")
	}
	if j.c == nil {
		return errors.New("ConnectionObjectNil")
	}
	return j.c.WriteMessage(websocket.TextMessage, []byte(msg))
	
}

func (j *JsonBarn) Close() error {
	if j==nil {
		return errors.New("JsonBarnIsNil")
	}
	j.c.Close()
	j.loggedIn = false
	j.connected = false
	j.exit = true
	return nil
}

// jsonBarn.Create
func (j *JsonBarn) Connect(Host, Port, Path, username, password string, tlsConfig *tls.Config) error {
	if j==nil {
		return errors.New("JsonBarnIsNil")
	}
	if tlsConfig == nil {
		j.NewDialer.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	} else {
		j.NewDialer.TLSClientConfig = tlsConfig
	}

	go func() {

		var err error
		u := url.URL{Scheme: "wss", Host: Host + ":" + Port, Path: Path}

		for {
			if j.exit {
				break
			}

			if !j.connected {
				j.loggedIn = false
				for {
					trace("connecting " + u.String())
					j.c, _, err = j.NewDialer.Dial(u.String(), nil)
					if err != nil {
						time.Sleep(time.Second)
					} else {
						j.connected = true
						trace("sending login " + username + " " + password)
						m := `{"$jsonbarn_action": "LOGIN", "$jsonbarn_username": "` + username + `", "$jsonbarn_password": "` + password + `"}`
						err = j.c.WriteMessage(websocket.TextMessage, []byte(m))
						break
					}
				}
			}


			_, message, err := j.c.ReadMessage()
			if err != nil {
				j.connected = false
				j.loggedIn = false
			} else {
				
				msg := string(message)
				trace("rx " + string(message))
						
				// validate json
				if !gjson.Valid(msg) {
					continue
				}
				value := gjson.Get(msg, "$login")
				if value.Exists() {
					trace("logged in!")
					j.loggedIn = true
				}
				trace("sending message into rx channel ")
		
				for {
					select {
						case  j.Ch <- message:
							break
						default:
							// oups buffer is full we have to wait until it's not 
							time.Sleep(100 * time.Nanosecond)
					}
				}
    
			}
		}
	}()

	return nil

}
