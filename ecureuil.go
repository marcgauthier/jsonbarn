/*
______________________________________________________________________________

 OWLSO - Overwatch Link and Service Observer.
______________________________________________________________________________

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

______________________________________________________________________________


This file contain the startup code for the backend server of OWLSO.



______________________________________________________________________________________


Revision:


	01 Nov 2016 - Clean code, audit.



______________________________________________________________________________

*/

package main

import (
	"errors"
	"flag"
	"fmt"
	"models"
	"net/http"
	"os"

	"github.com/antigloss/go/logger"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

// Object to upgrade an https connection to a WSS connection

var upgrader = websocket.Upgrader{} // use default options

/*

Handle the request to open a Websocket service connection, a Client object
will be created to maintain informtion about the connectin and
to push broadcast when database changes occur.

*/

func wsPage(w http.ResponseWriter, r *http.Request) {

	conn, err := upgrader.Upgrade(w, r, nil)

	if err != nil {
		http.NotFound(w, r)
		return
	}

	models.ClientAdd(conn)

}

/*
	Handle all request for a static files.  The OWLSO backend does answer to
	request for Data in the database but also server static files.
	The files are not server from a folder but from the database itself.
	Any files stored in the database is public user does not require
	special rights to read them.

*/

type appHandler func(http.ResponseWriter, *http.Request) (int, error)

// Our appHandler type will now satisify http.Handler
func (fn appHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if status, err := fn(w, r); err != nil {

		logger.Error(err.Error())

		switch status {
		// We can have cases as granular as we like, if we wanted to
		// return custom errors for specific status codes.
		//  case http.StatusNotFound:
		//      notFound(w, r)

		case http.StatusInternalServerError:
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		default:
			// Catch any other errors we haven't explicitly handled
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
	}
}

/*
	Handle all request for a static files see previous comment.

*/

func myHandler(w http.ResponseWriter, r *http.Request) (int, error) {

	filepath := ""

	// only answer GET request.

	if r.Method == "GET" {

		// if home page redirect to index.html

		if r.URL.Path == "/" {
			filepath = "/index.html"

		} else if r.URL.Path == "/confirm/" {

			// https://"+Configuration.Addr+"/confirm/?id="+Info.ID)"
			logger.Trace("Receive Email Alert confirmation, verifying")
			queryValues := r.URL.Query()
			ID := queryValues.Get("ID")
			if ID == "" {
				logger.Trace("ID was not provided for email alert confirmation")
				w.Write([]byte("No ID request was provided!"))
				return http.StatusOK, nil
			}

			err := models.ReceiveConfirmationEmailAlert(ID)
			if err == nil {
				logger.Trace("Email Alert processed!")
				w.Write([]byte("Your request to change email alerts has been approved!"))
				return http.StatusOK, nil
			}
			logger.Trace("Email Alert ID was not found!")
			w.Write([]byte("The ID request you have provided cannot be found!"))
			return http.StatusOK, nil

		} else {

			filepath = r.URL.Path
		}

		// Extract the file from the database if present.

		buffer, ext, err := models.GetStaticFile(filepath)

		if err != nil || buffer == nil {
			logger.Warn("HTTP REQ not found (GET): " + filepath)
			return http.StatusNotFound, errors.New(http.StatusText(http.StatusNotFound)) // ... and again.
		}

		// Verify what type of file we just requested, the function GetStaticFile always return the ext in lowercase.
		// We must set the content-type variable that we will be returning to the client.

		if ext == ".css" {
			w.Header().Set("Content-Type", "text/css")
		} else if ext == ".jpg" || ext == ".jpeg" {
			w.Header().Set("Content-Type", "image/jpeg")
		} else if ext == ".png" {
			w.Header().Set("Content-Type", "image/png")
		} else if ext == ".js" {
			w.Header().Set("Content-Type", "application/javascript")
		}

		// Content-Length will be set automatically

		w.Write(buffer)

	}

	return http.StatusOK, nil
}

/*

	Main function of the program

*/

func redirectToHTTPS(w http.ResponseWriter, r *http.Request) {
	// Redirect the incoming HTTP request to HTTPS
	logger.Trace("redirecting http REQ to https://" + r.Host + r.URL.String())
	http.Redirect(w, r,
		"https://"+r.Host+r.URL.String(),
		http.StatusMovedPermanently)
}

//sudo ./ecureuil -createdb -host=192.168.56.101 -user=postgres -password=bitnami
//sudo ./ecureuil -dropdb -host=192.168.56.101 -user=postgres -password=bitnami
/// -trace to show trace!

func main() {

	user := flag.String("user", "", "postgresql user name with admin rights")
	password := flag.String("password", "", "postgresql user password")
	host := flag.String("host", "", "postgresql server ip or hostname")
	boolPtr := flag.Bool("createdb", false, "a bool")
	boolPtr2 := flag.Bool("dropdb", false, "a bool")
	showTraceFlag := flag.Bool("trace", false, "a bool")

	flag.Parse()

	if *boolPtr {
		result := models.CreateDB(host, user, password)
		fmt.Println(result)
		return
	}

	if *boolPtr2 {
		result := models.DropDB(host, user, password)
		fmt.Println(result)
		return
	}

	showTrace := false
	if *showTraceFlag {
		showTrace = true
	}

	if *host == "" {
		for {
			fmt.Println("Enter Postgre host/ip: ")
			input := ""
			fmt.Scanln(&input)
			if input != "" {
				break
			}
			*host = input
		}
	}

	if *user == "" {
		for {
			fmt.Println("Enter Postgre username: ")
			input := ""
			fmt.Scanln(&input)
			if input != "" {
				break
			}
			*user = input
		}
	}

	if *password == "" {
		for {
			fmt.Println("Enter Postgre password: ")
			input := ""
			fmt.Scanln(&input)
			if input != "" {
				break
			}
			*password = input
		}
	}

	//fmt.Println("Host=" + *host + " username=" + *user)

	// create folder for storing log files if it does not exists.
	path := "./log"
	if _, err := os.Stat(path); os.IsNotExist(err) {
		os.Mkdir(path, os.ModeDir)
	}

	// logger options are Trace, Info, Warn, Error, Panic and Abort
	// logs are saved into the log/ folder

	if *boolPtr2 {
		result := models.DropDB(host, user, password)
		fmt.Println(result)
		return
	}

	logger.Init("./log", // specify the directory to save the logfiles
		400,       // maximum logfiles allowed under the specified log directory
		20,        // number of logfiles to delete when number of logfiles exceeds the configured limit
		100,       // maximum size of a logfile in MB
		showTrace) // whether logs with Trace level are written down

	logger.SetLogToConsole(true) // show all log on the monitor

	models.InitFileCache()

	/* OPEN database, initialize and create default values if required */
	models.Open(*host, *user, *password)
	defer models.Close()

	logger.Trace("Openning Database Completed!")

	/* start the hub that broadcast messages on Websocket connections */
	logger.Trace("Starting the HUB go routine")
	models.HubStart()

	/* Create HTTPS mux router */
	r := mux.NewRouter()
	r.HandleFunc("/wss/", wsPage)                    // websocket request
	r.PathPrefix("/").Handler(appHandler(myHandler)) // any other request check for static files
	http.Handle("/", r)

	/* if server.crt or server.key does not exists create the files!! */

	CertificateExists, _ := models.FileExists("server.crt")

	KeyExists, _ := models.FileExists("server.key")

	if !CertificateExists || !KeyExists {

		logger.Info("Creating https certificates...")

		err := models.CreateCertificates()

		if err != nil {

			logger.Error(err.Error())
			logger.Panic("Unable to create certificates, please provide server.crt and server.key.")

			panic("Unable to create certificates, please provide server.crt and server.key.")

		} else {

			logger.Info("Certificates successfuly created.")
		}

	}

	/*
		Start the webserver and respond to HTTPS request only
	*/

	s := "localhost:443"

	logger.Trace("Starting the HTTPS Listenner on " + s)

	go http.ListenAndServeTLS(s, "server.crt", "server.key", nil)

	// starting a redirection service from HTTP to HTTPS
	req := "HTTP localhost:80 to HTTPS localhost:443"

	logger.Trace("Starting the HTTP to HTTPS redirect Listenner " + req)

	http.ListenAndServe("localhost:80", http.HandlerFunc(redirectToHTTPS))

}
