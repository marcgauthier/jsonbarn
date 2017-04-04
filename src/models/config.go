/*Package models Config.GO
______________________________________________________________________________

This file contain functions to control the CONFIGURATION bucket,
this Bucket contain options and settings for the program.
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

	"github.com/antigloss/go/logger"
	"github.com/asaskevich/govalidator"
	uuid "github.com/satori/go.uuid"
)

/*tsmtp use for CONFIGURATION

This bucket should not sync with other server

*/

const configIdValue = "a62cfcd3-a7d2-4483-aaa8-2931be5927fb" // use to read and save config in SQL

type tsmtp struct {
	IP string `json:"smtpip"` // smtp server ip to send alert email

	User string `json:"smtpuser"` // smtp server username to send alert email

	Password string `json:"smtppassword"` // smtp server password to send alert email

	Port int `json:"smtpport"` // smtp server port to send alert email

	Emailfrom string `json:"smtpemailfrom"` // smtp email from to send alert email

	Ssl int `json:"smtpssl"` // smtp server use SSL to send alert email

	Timeout int `json:"smtptimeout"` // smtp server timeout

	Auth int `json:"smtpauth"` // smtp type of authentication

	Function int `json:"smtpfunction"` // smtp type of function

	Enabled int `json:"smtpenabled"` // is smtp email alert enabled? default is false

}

/*TConfig structure type to hold all the configuration parameters

This Bucket will contain "current" key configuration and any other backup configutation

This bucket should not sync with other serverStat

*/
type TConfig struct {
	ID string `storm:"id" json:"serverid"` // unique server id UUIL

	NetworkID string `json:"networkid"` // what network this server is running on? from user networks list

	POSTGRESQLUser string `json:"postgresqluser"`

	POSTGRESQLPass string `json:"postgresqlpass"`

	POSTGRESQLHost string `json:"postgresqlhost"`

	POSTGRESQLPort int `json:"postgresqlport"`

	KeepLogForDays int `json:"keeplogfordays"` // indicate number of days to keep log default 365 days

	SMTP tsmtp `storm:"inline" json:"smtp"` // smtp details to send email!

	Port int `json:"port"` // listen for connection on this port

	Addr string `json:"addr"` // list for connection on this ip blank is localhost.

	MaxReadItemsFromDB int `json:"maxreaditemsfromdb"` // maximum number of items to return per query.

	LoginPerMin int `json:"loginpermin"` // amount of login attemp allowed per minute.

	EmailAlertSubject string `json:"emailalertsubject"`

	EmailAlertBody string `json:"emailalertbody"`

	MaxIdleSQLConns int `json:"maxidlesqlconns"`

	MaxOpenSQLConns int `json:"maxopensqlconns"`

	MaxLifetimeSQLConns int `json:"maxlifetimesqlconns"` // in seconds default to 0 unlimited
}

/*ConfigBUCKET name of the command send by front-end to access the configuration.
  This value is use by database.go
*/
var ConfigBUCKET = []byte("CONFIGURATION")

/*Configuration hold the current Configuration.
 */
var Configuration TConfig

/*GetConfiguration rights have already been checked
 */
func GetConfiguration(packet *MsgClientCmd) ([]byte, error) {

	logger.Trace("request read system configuration ")

	// if access if not granted by default then check if the user has rights
	access, err := UserHasRight([]byte(packet.Username), []byte(packet.Password), "CONFIGURATION-read")
	if err != nil {
		logger.Warn(packet.Username + " read configuration error: " + err.Error())
		return PrepMessageForUser("Error while reading."), err
	}

	if access == false {
		logger.Warn(packet.Username + " read configuration access denied.")
		return PrepMessageForUser("Access denined."), err
	}

	logger.Trace(packet.Username + " reed system configuration")

	j, err := json.Marshal(Configuration)

	if err != nil {
		return nil, err
	}

	buffer := new(bytes.Buffer)

	// what type of information user want to extract?

	buffer.WriteString("{\"action\":\"read\", \"bucketname\": \"CONFIG\", \"items\" : [")

	buffer.Write(j)

	buffer.WriteString("]}")

	return buffer.Bytes(), nil

}

/*PutConfiguration this function is called by func DBUpdateBucketValue inside the database.go file.
To verify new config and save it to the database.
*/
func PutConfiguration(packet *MsgClientCmd) ([]byte, error) {

	var err error

	logger.Trace("request update system configuration ")

	// if access if not granted by default then check if the user has rights
	access, err := UserHasRight([]byte(packet.Username), []byte(packet.Password), "CONFIGURATION-write")
	if err != nil {
		logger.Warn(packet.Username + " update " + packet.Bucketname + " error: " + err.Error())
		return PrepMessageForUser("Error while updating or access denied."), err
	}

	if access == false {
		logger.Warn(packet.Username + " update " + packet.Bucketname + " access denied.")
		return PrepMessageForUser("Access denined."), err
	}

	item := TConfig{}

	// deserialize object to confirm it is actually valid
	//***************************************************
	errcode := json.Unmarshal(packet.Data, &item)
	if errcode != nil {
		logger.Error("Unable to unmarshal configuration provided by User: " + packet.Username + " error: " + errcode.Error())
		return PrepMessageForUser("Configuration provided is unreadable"), errors.New("Configuration provided is unreadable")
	}

	// Make sure value in the config object are valid.
	//************************************************
	err = ValidateConfig(&item)
	if err != nil {
		logger.Warn("Can't validate configuration provided by User: " + packet.Username + " error: " + errcode.Error())
		return nil, err
	}

	// overwrite header information.
	//************************************************
	oldconfig, err := json.Marshal(Configuration)

	Configuration.NetworkID = item.NetworkID
	Configuration.SMTP.IP = item.SMTP.IP
	Configuration.SMTP.User = item.SMTP.User
	Configuration.SMTP.Password = item.SMTP.Password
	Configuration.SMTP.Port = item.SMTP.Port
	Configuration.SMTP.Emailfrom = item.SMTP.Emailfrom
	Configuration.SMTP.Ssl = item.SMTP.Ssl
	Configuration.SMTP.Timeout = item.SMTP.Timeout
	Configuration.SMTP.Auth = item.SMTP.Auth
	Configuration.SMTP.Function = item.SMTP.Function
	Configuration.SMTP.Enabled = item.SMTP.Enabled
	Configuration.Addr = item.Addr
	Configuration.Port = item.Port
	Configuration.MaxReadItemsFromDB = item.MaxReadItemsFromDB
	Configuration.POSTGRESQLHost = item.POSTGRESQLHost
	Configuration.POSTGRESQLPass = item.POSTGRESQLPass
	Configuration.POSTGRESQLPort = item.POSTGRESQLPort
	Configuration.POSTGRESQLUser = item.POSTGRESQLUser
	Configuration.KeepLogForDays = item.KeepLogForDays
	Configuration.LoginPerMin = item.LoginPerMin
	Configuration.EmailAlertSubject = item.EmailAlertSubject
	Configuration.EmailAlertBody = item.EmailAlertBody
	Configuration.MaxIdleSQLConns = item.MaxIdleSQLConns
	Configuration.MaxOpenSQLConns = item.MaxOpenSQLConns
	Configuration.MaxLifetimeSQLConns = item.MaxLifetimeSQLConns

	// ReSerialize packet to save and do not broadast.
	// user can set any key they want but "currentconfig" need to be use
	// to change the current settings, other key can be use for
	// storing backup settings in the database.
	//************************************************

	err = saveConfig(&Configuration)
	if err != nil {
		logger.Error(err.Error())
		return PrepMessageForUser("Error while saving configuration:" + err.Error()), nil
	}

	go DBLog("CONFIGURATION", packet.Username, "UPDATE", oldconfig, packet.Data)

	return PrepMessageForUser("Configuration saved"), nil

}

/*ConfigurationINIT this function is called to create the bucket and default configuration.
 */
func ConfigurationINIT() {

	sqlquery := "select DATA FROM ecureuil.jsonobjects WHERE ID = $1"

	logger.Trace(sqlquery)

	rows, err := sqldb.Query(sqlquery, configIdValue)

	if err != nil {

		logger.Error(err.Error())
		panic(err.Error())

	}

	if rows.Next() {

		var data string
		err = rows.Scan(&data)
		if err != nil {
			logger.Error(err.Error())
			panic("bad configuraton!")
		}

		err = json.Unmarshal([]byte(data), &Configuration)
		if err != nil {
			logger.Error(err.Error())
			panic("bad configuraton!")
		}
		logger.Trace("Configuration reed from SQL!")
		return // good to go!
	}

	logger.Trace("no configuration found creating a new one.")
	setDefaultConfig()
	err = saveConfig(&Configuration)
	if err != nil {
		logger.Error(err.Error())
		panic("Unable to save configuration!")
	}

}

/*ValidateConfig this function validate configuration provided by the FRONTEND
 */
func ValidateConfig(config *TConfig) error {

	if config.SMTP.IP != "" {
		if !govalidator.IsIP(config.SMTP.IP) {
			return errors.New("SMTP IP is not a valid IP address")
		}
	}

	if config.SMTP.Emailfrom != "" {
		if !govalidator.IsEmail(config.SMTP.Emailfrom) {
			return errors.New("SMTP Email from is not a valid email address")
		}
	}

	if config.SMTP.Port < 0 || config.SMTP.Port > 65535 {
		return errors.New("SMTP Port is not valid (0..65535)")
	}

	sqlquery := "select DATA FROM ecureuil.jsonobjects WHERE ID = $1"

	logger.Trace(sqlquery)

	rows, err := sqldb.Query(sqlquery, configIdValue)

	if err != nil {

		logger.Error(err.Error())
		panic(err.Error())

	}

	if rows.Next() {

		var data string
		err = rows.Scan(&data)
		if err != nil {
			logger.Error(err.Error())
			panic("bad configuraton!")
		}

		err = json.Unmarshal([]byte(data), &Configuration)
		if err != nil {
			logger.Error(err.Error())
			panic("bad configuraton!")
		}
		logger.Trace("Configuration reed from SQL!")
		return nil // good to go!
	}
	return nil
}

func saveConfig(configuration *TConfig) error {

	newconfig, err := json.Marshal(Configuration)
	if err != nil {
		return err
	}

	sqlquery := "INSERT INTO ecureuil.JSONOBJECTS (ID, DATA, bucketname, CREATEDBY, UPDATEDBY, CREATEDTIME, UPDATEDTIME, CREATEDONSERVER) " +
		"VALUES ($1, $2, 'CONFIG', 'SYSTEM', 'SYSTEM',  $3, $3, $4) ON CONFLICT (ID) DO UPDATE SET DATA = $2 WHERE ecureuil.JSONOBJECTS.ID = $1;"

	_, err = sqldb.Exec(sqlquery, configIdValue, string(newconfig), uint64(UnixUTCSecs()), Configuration.ID)

	return err
}

func setDefaultConfig() {

	Configuration.ID = uuid.NewV4().String() // ServerID
	Configuration.MaxReadItemsFromDB = 1000000

	Configuration.SMTP.IP = ""
	Configuration.SMTP.User = ""
	Configuration.SMTP.Password = ""
	Configuration.SMTP.Port = 25
	Configuration.SMTP.Emailfrom = ""
	Configuration.SMTP.Ssl = 0
	Configuration.SMTP.Timeout = 60
	Configuration.SMTP.Auth = 0
	Configuration.SMTP.Function = 0
	Configuration.SMTP.Enabled = 0
	Configuration.Port = 443
	Configuration.Addr = ""
	Configuration.LoginPerMin = 3
	Configuration.KeepLogForDays = 365
	Configuration.MaxReadItemsFromDB = 1000000
	Configuration.NetworkID = ""
	Configuration.POSTGRESQLHost = "192.168.56.101"
	Configuration.POSTGRESQLPass = "bitnami"
	Configuration.POSTGRESQLPort = 5432
	Configuration.POSTGRESQLUser = "ecureuiladmin"
	Configuration.EmailAlertSubject = "Email Alert Change request confirmation!"
	Configuration.EmailAlertBody = "Hello, you have recently made a request to receive or stop receiving email alert from our system" +
		" please click the link bellow to confirm you want to activate the changes."
	Configuration.MaxOpenSQLConns = 0
	Configuration.MaxIdleSQLConns = 0
	Configuration.MaxLifetimeSQLConns = 0
	Configuration.LoginPerMin = 3

}
