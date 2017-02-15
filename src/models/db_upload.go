/*Package models db_upload.go

This file contain functions to help upload the entire database thru an
HTTPS pipe so that a user can create backup copies remotely.
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


Revision:
    01 Nov 2016 - Clean code, audit.

______________________________________________________________________________

*/
package models

import (
	"net/http"
	"strconv"
	"time"

	"github.com/antigloss/go/logger"
	"github.com/boltdb/bolt"
)

/*Upload this function is called by GO http handler with a Request and Response object
the Handler is in owlso.go
*/
func Upload(w http.ResponseWriter, r *http.Request) {

	/*
			Parse parameters so we can read username and password to confirm credentials
			Parameters are sent using POST command of HTTPS
		 	HTTPS is use to protect username and password.
	*/

	err := r.ParseForm()
	if err != nil {
		logger.Error(err.Error())
		http.Error(w, "Unable to process form data", 500)
		return
	}

	/* Confirm that user as db-download rights */

	access, err := UserHasRight([]byte(r.FormValue("username")), []byte(r.FormValue("password")), "DB-DOWNLOAD")
	if err != nil {
		logger.Error("Download database: " + err.Error())
		// return not authorized to download
		http.Error(w, "Download database error!", 500)
		return
	}

	/* If user as db-download rights send database. */

	if access == true {

		/* generate a file with the date and time  */
		filename := "owlsodb_bak_UTC_" + time.Now().UTC().Format("2006-Jan-02_1504") + ".db"

		// open the database and send it to the https connection

		err := DB.Bolt.View(func(tx *bolt.Tx) error {
			w.Header().Set("Content-Type", "application/octet-stream")
			w.Header().Set("Content-Disposition", `attachment; filename="`+filename+`"`)
			w.Header().Set("Content-Length", strconv.Itoa(int(tx.Size())))
			_, err := tx.WriteTo(w)
			return err
		})

		if err != nil {
			logger.Error("Error while downloading the database user: " + r.FormValue("username") + " " + err.Error())
			http.Error(w, "Error while downloading the database!", 500)
			return
		}

	} else {

		logger.Warn("User try to download database but access denied: " + r.FormValue("username"))
		http.Error(w, "Download database access denied!", 500)

	}
}
