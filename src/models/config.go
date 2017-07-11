/*Package models Config.GO
______________________________________________________________________________

This file contain functions to control the CONFIGURATION bucket,
this Bucket contain options and settings for the program.
______________________________________________________________________________

______________________________________________________________________________

 Ecureuil - Web framework for real-time javascript app.
_____________________________________________________________________________


MIT License

Copyright (c) 2014-2017 Marc Gauthier

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

  07 Jul 2017 - Clean code, audit.
  01 Nov 2016 - Clean code, audit.

______________________________________________________________________________
*/
package models

import (
	"bytes"
	"encoding/json"
	"errors"

	"github.com/Jeffail/gabs"
	"github.com/antigloss/go/logger"
	"github.com/asaskevich/govalidator"
	uuid "github.com/satori/go.uuid"
)

/*tsmtp use for CONFIGURATION

This bucket should not sync with other server

*/

const configIdValue = "a62cfcd3-a7d2-4483-aaa8-2931be5927fb" // use to read and save config in SQL

/*TConfig structure type to hold all the configuration parameters

This Bucket will contain "current" key configuration and any other backup configutation

This bucket should not sync with other serverStat

*/
type TConfig struct {
	ID string `json:"serverid"` // unique server id UUIL

	NetworkID string `json:"networkid"` // what network this server is running on? from user networks list

	POSTGRESQLUser string `json:"postgresqluser"`

	POSTGRESQLPass string `json:"postgresqlpass"`

	POSTGRESQLHost string `json:"postgresqlhost"`

	POSTGRESQLPort int `json:"postgresqlport"`

	KeepLogForDays int `json:"keeplogfordays"` // indicate number of days to keep log default 365 days

	SMTPIP string `json:"smtpip"` // smtp server ip to send alert email

	SMTPUser string `json:"smtpuser"` // smtp server username to send alert email

	SMTPPassword string `json:"smtppassword"` // smtp server password to send alert email

	SMTPPort int `json:"smtpport"` // smtp server port to send alert email

	SMTPEmailfrom string `json:"smtpemailfrom"` // smtp email from to send alert email

	SMTPSsl int `json:"smtpssl"` // smtp server use SSL to send alert email

	SMTPTimeout int `json:"smtptimeout"` // smtp server timeout

	SMTPAuth int `json:"smtpauth"` // smtp type of authentication

	SMTPFunction int `json:"smtpfunction"` // smtp type of function

	SMTPEnabled int `json:"smtpenabled"` // is smtp email alert enabled? default is false

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

/*GetConfiguration rights have already been checked configuration is from memory not what saved on SQL
 */
func GetConfiguration(packet *MsgClientCmd) ([]byte, error) {

	logger.Trace("request read system configuration ")

	// check if user as configuration-read right or admin
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

	Configuration.NetworkID = item.NetworkID
	Configuration.SMTPIP = item.SMTPIP
	Configuration.SMTPUser = item.SMTPUser
	Configuration.SMTPPassword = item.SMTPPassword
	Configuration.SMTPPort = item.SMTPPort
	Configuration.SMTPEmailfrom = item.SMTPEmailfrom
	Configuration.SMTPSsl = item.SMTPSsl
	Configuration.SMTPTimeout = item.SMTPTimeout
	Configuration.SMTPAuth = item.SMTPAuth
	Configuration.SMTPFunction = item.SMTPFunction
	Configuration.SMTPEnabled = item.SMTPEnabled
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

	err = saveConfig(&Configuration, packet.Username)

	if err != nil {
		logger.Error(err.Error())
		return PrepMessageForUser("Error while saving configuration:" + err.Error()), nil
	}

	return PrepMessageForUser("Configuration saved"), nil

}

/*ConfigurationINIT this function is called to create the bucket and default configuration.
 */
func ConfigurationINIT() {

	sqlquery := "select DATA FROM ecureuil.jsonobjects WHERE data->>'$id' = $1"

	logger.Trace(sqlquery)

	rows, err := sqldb.Query(sqlquery, configIdValue)

	if err != nil {

		// unable to read configuration
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
	err = saveConfig(&Configuration, "system")
	if err != nil {
		logger.Error(err.Error())
		panic("Unable to save configuration!")
	}

}

/*ValidateConfig this function validate configuration provided by the FRONTEND
 */
func ValidateConfig(config *TConfig) error {

	if config.SMTPIP != "" {
		if !govalidator.IsIP(config.SMTPIP) {
			return errors.New("SMTP IP is not a valid IP address")
		}
	}

	if config.SMTPEmailfrom != "" {
		if !govalidator.IsEmail(config.SMTPEmailfrom) {
			return errors.New("SMTP Email from is not a valid email address")
		}
	}

	if config.SMTPPort < 0 || config.SMTPPort > 65535 {
		return errors.New("SMTP Port is not valid (0..65535)")
	}

	// configuration is valid
	return nil
}

func saveConfig(configuration *TConfig, Username string) error {

	newconfig, err := json.Marshal(Configuration)
	if err != nil {
		return err
	}

	jsonParsed, err := gabs.ParseJSON([]byte(newconfig))

	if err != nil {
		logger.Error(err.Error())
		return err
	}

	jsonParsed.SetP(configIdValue, "$id")
	jsonParsed.SetP(Configuration.NetworkID, "$createdonnetwork")
	jsonParsed.SetP(Configuration.ID, "$createdonserver")
	jsonParsed.SetP(uint64(UnixUTCSecs()), "$createdtime")
	jsonParsed.SetP(uint64(UnixUTCSecs()), "$updatedtime")
	jsonParsed.SetP(Username, "$updatedby")

	sqlquery := "INSERT INTO ecureuil.JSONOBJECTS (DATA) " +
		"VALUES ($2) ON CONFLICT (data->>'$id') DO UPDATE SET DATA = $2 WHERE ecureuil.JSONOBJECTS.data->>'$id' = $1;"

	sqlquery = "INSERT INTO ecureuil.JSONOBJECTS (DATA) VALUES ($1);"

	logger.Trace(sqlquery)
	logger.Trace(jsonParsed.String())

	_, err = sqldb.Exec(sqlquery, jsonParsed.String())

	return err
}

func setDefaultConfig() {

	Configuration.ID = uuid.NewV4().String() // ServerID
	Configuration.MaxReadItemsFromDB = 1000000

	Configuration.SMTPIP = ""
	Configuration.SMTPUser = ""
	Configuration.SMTPPassword = ""
	Configuration.SMTPPort = 25
	Configuration.SMTPEmailfrom = ""
	Configuration.SMTPSsl = 0
	Configuration.SMTPTimeout = 60
	Configuration.SMTPAuth = 0
	Configuration.SMTPFunction = 0
	Configuration.SMTPEnabled = 0
	Configuration.Port = 443
	Configuration.Addr = ""
	Configuration.LoginPerMin = 3
	Configuration.KeepLogForDays = 365
	Configuration.MaxReadItemsFromDB = 1000000
	Configuration.NetworkID = ""
	Configuration.POSTGRESQLHost = "127.0.0.1"
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
