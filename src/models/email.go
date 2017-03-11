/*Package models - misc.go

This file contain miscillaneous functions helper.

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
	"bytes"
	"encoding/json"
	"errors"
	"html/template"
	"log"
	"net/smtp"
	"strconv"

	"github.com/antigloss/go/logger"
	uuid "github.com/satori/go.uuid"
	"github.com/tidwall/gjson"
)

/*EmailAlertRequest contain the struct to change the email alert
 */
type EmailAlertRequest struct {
	ID          string   `storm:"id" json:"id"`
	Email       string   `json:"email"`
	Buckets     []string `json:"buckets"`
	DateRequest float64  `json:"daterequest"`
}

/*EmailAlert contain the struct to save all the email alert
 */
type EmailAlert struct {
	Email   string   `storm:"id" json:"email"`
	Buckets []string `json:"buckets"`
}

/*ReceiveEmailAlertChangeReq Receive a request to change email alert from the websocket

self.serversocket.send("{\"action\":\"EMAILALERT\", \"data\": {\"email\":\"" + email + "\", \"buckets\":" + JSON.stringify(buckets) + " }}");

*/
func ReceiveEmailAlertChangeReq(packet *MsgClientCmd) ([]byte, error) {

	logger.Trace("Processing email alert change request")
	// generate uuid code
	// save change request in STORM
	// send email

	Info := EmailAlertRequest{}
	err := json.Unmarshal(packet.Data, &Info)
	if err != nil {
		logger.Error(err.Error())
		return PrepMessageForUser(err.Error()), err
	}
	Info.ID = uuid.NewV4().String() // generate an ID that need to be confirm by user using get url from browser
	Info.DateRequest = UnixUTCSecs()

	DB.Save(&Info)

	SendEmail([]string{Info.Email},
		Configuration.SMTP.Emailfrom,
		Configuration.EmailAlertSubject,
		Configuration.EmailAlertBody+"\n\n https://"+Configuration.Addr+"/confirm/?ID="+Info.ID)

	return PrepMessageForUser("A confirmation request has been sent to your email address."), nil

}

/*ReceiveConfirmationEmailAlert This function is called when user click on the http link to confirm a req change email alert
 */
func ReceiveConfirmationEmailAlert(ID string) error {

	logger.Trace("Confirming request id " + ID)

	Info := EmailAlertRequest{}
	err := DB.One("ID", ID, &Info)
	if err != nil {
		return err
	}

	// we are good save the list of buckets (alert) for the specify email address
	e := EmailAlert{}
	e.Buckets = Info.Buckets
	e.Email = Info.Email
	err = DB.Save(&e)
	if err != nil {

		err = DB.Update(&e)
		if err != nil {
			return err
		}
	}

	// success delete the Change request
	DB.DeleteStruct(&Info)

	return nil
}

/*SendEmail This function send a built email.

models.SendEmail([]string{"marc.gauthier3@gmail.com"}, "marc.gauthier3@gmail.com", "allo", "this is my info")
	panic("test email")


*/
func SendEmail(to []string, from, subject, body string) {

	logger.Trace("sending email")

	/* for debuging! */
	Configuration.SMTP.Enabled = 1
	Configuration.SMTP.User = "marc.gauthier3@gmail.com"
	Configuration.SMTP.Password = "azy4azy4"
	Configuration.SMTP.IP = "smtp.gmail.com"
	Configuration.SMTP.Port = 587 //465 //587
	Configuration.SMTP.Emailfrom = "marc.gauthier3@gmail.com"

	// check if email alert are enabled.
	if Configuration.SMTP.Enabled == 0 {
		return
	}

	// Set up authentication information.
	auth := smtp.PlainAuth("", Configuration.SMTP.User, Configuration.SMTP.Password, Configuration.SMTP.IP)

	tolist := ""
	for i := 0; i < len(to); i++ {
		if i > 0 {
			tolist += ";" + to[i]
		} else {
			tolist += to[i]
		}
	}

	msg := []byte("To: " + tolist + "\r\n" +
		"From: " + from + "\r\n" +
		"Subject: " + subject + "\r\n" +
		body)

	logger.Trace("connecting to smtp server")
	err := smtp.SendMail(Configuration.SMTP.IP+":"+strconv.Itoa(Configuration.SMTP.Port), auth, from, to, msg)

	if err != nil {
		log.Fatal("fataerror :" + err.Error())
		return
	}
	logger.Trace("No error on smtp func")

}

/* return list of email address of user that want to receive email alert from
a specific bucketname
*/
func generateToList(bucketname string) []string {

	var list []string

	var users []EmailAlert

	err := DB.All(&users)

	if err != nil {
		return list
	}

	for i := range users {
		if IsStrInArray(bucketname, users[i].Buckets) {
			list = append(list, users[i].Email)
		}
	}

	return list
}

/*GenerateEmailTemplate Generate html code for sending an email after a change

Templates are store as JSON inside the jsonobjects database they need to have
the bucketname set to TEMPLATES with a JSON object that contain:
status int and body text document containing the html template and
subject text document containing the html template for the subject.

*/
func GenerateEmailTemplate(bucketname string, jsonobject string) error {

	// Extract status from the jsonobject we need it to select the correct template!
	value := gjson.Get(jsonobject, "status")
	status := value.Int()

	// get data from this json that contain data + bucket + ...
	data := gjson.Get(jsonobject, "")

	items, ok := gjson.Parse(data.String()).Value().(map[string]interface{})
	if !ok {
		// not a map
		logger.Error("Unable to use GJSON to get data from item")
		return nil
	}

	if !(status >= 0 && status <= 255) {
		msg := "Generate EmailTemplate status inside jsonobject is not of type int:"
		logger.Error(msg + jsonobject)
		return errors.New(msg)
	}

	// valid status are 0 pending, 1 active, 2 completed
	// search and select template
	query := "SELECT DATA->>'body' AS body, DATA->>'subject' AS subject FROM ecureuil.JSONOBJECTS WHERE BucketName = 'TEMPLATES' AND DATA->>'bucket' = $2 AND CAST(DATA->>'status' AS INT) = $1"

	rows, err := sqldb.Query(query, status, bucketname)

	if err != nil {
		logger.Error(err.Error())
		return err
	}

	if rows == nil {
		m := "No template available for bucketname " + bucketname + " with status " + strconv.Itoa(int(status))
		logger.Trace(m)
		return errors.New(m)
	}

	for rows.Next() {

		var body string
		var subject string

		err = rows.Scan(&body, &subject)
		if err != nil {
			logger.Error(err.Error())
			return err
		}

		logger.Trace("Generating body template")
		t := template.New(bucketname + "body")
		t, _ = t.Parse(body)
		var BodyOutput bytes.Buffer
		t.Execute(&BodyOutput, items)

		logger.Trace("generating subject template")
		s := template.New(bucketname + "subject")
		s, _ = s.Parse(subject)
		var SubjectOutput bytes.Buffer
		s.Execute(&SubjectOutput, items)

		logger.Trace("generating email to list")
		//to := generateToList(bucketname)

		logger.Trace("Subject: " + SubjectOutput.String())
		logger.Trace("Body: " + BodyOutput.String())

		//SendEmail(to, Configuration.SMTP.Emailfrom, SubjectOutput.String(), BodyOutput.String())

		// should be only one template with status and bucketname but it's possible to have more than one.

	}

	return nil

}
