/*Package models - database.go
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


/////////////////////////////////////////////////////////////////////////////


/////////////////////////////////////////////////////////////////////////////

POSTGRESQL should only be accessible by localhost because any other app
will first connect to ecureuil to be part of the broadcast network
and share data in real-time.  ecureuil app is the only app that should
be accessing the SQL



______________________________________________________________________________

DATABASE.go
===========

Save data comming from a Javascript SPA.
Use Secure Websocket protocol.

Special Buckets: these bucket are saved in the Storm database not the SQL.

	CONFIGURATION 	<< this is a fixed type bucket.
	USERS 			<< Only admin user can modify admin user, etc.


______________________________________________________________________________________


Revision:
    01 Nov 2016 - Clean code, audit.

______________________________________________________________________________

*/
package models

import (
	"bytes"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/Jeffail/gabs"
	"github.com/antigloss/go/logger"
	"github.com/asaskevich/govalidator"
	"github.com/lib/pq"
	uuid "github.com/satori/go.uuid"
)

var sqldb *sql.DB

var connstring = ""

/*Open Function called at the start of the program to open the database.
 */
func Open(host, username, password string) {

	err := errors.New("")

	connstring = "dbname=postgres user=" + username + " host=" + host + " password=" + password + " sslmode=disable"

	logger.Trace("Opening connection to POSTGRESQL database")

	sqldb, err = sql.Open("postgres", connstring)

	sqldb.SetMaxIdleConns(Configuration.MaxIdleSQLConns)
	sqldb.SetMaxOpenConns(Configuration.MaxOpenSQLConns)
	sqldb.SetConnMaxLifetime(time.Second * time.Duration(Configuration.MaxLifetimeSQLConns))

	if err != nil {
		logger.Error(err.Error())
		panic(err.Error())
	}

	/* Initialize database, create USERS and CONFIGURATION Buckets with  default values */
	ConfigurationINIT()

	/* initialize users database make sure at least one admin user exists */
	UsersINIT()

	reportProblem := func(ev pq.ListenerEventType, err error) {
		if err != nil {
			logger.Trace(err.Error())
		}
	}

	/* monitor the list of defered command and run them when required */
	go runDeferedEvents()

	/* monitor any object that have a starttime, endtime and recurrence and change their status automatically */
	go runMonitorStatusStartEnd()

	logger.Trace("Starting monitoring PostgreSQL...")

	listener := pq.NewListener(connstring, 10*time.Second, time.Minute, reportProblem)
	err = listener.Listen("events_ecureuil")
	if err != nil {
		panic(err)
	}

	go func() {
		for {
			// process all available work before waiting for notifications
			err := waitForNotification(listener)
			if err != nil {
				logger.Error(err.Error())
			}
		}
	}()

}

/*Close Function called to safely close the database.
 */
func Close() {

	logger.Trace("Closing SQL connection.")

	sqldb.Close()

}

/*
DBLog a change in the database
*/
func DBLog(bucketname, username, action string, PreviousData, NewData []byte) {

	jsonParsed, err := gabs.ParseJSON(NewData)

	id := "00000000-0000-0000-0000-000000000000"

	if err == nil {
		id = jsonParsed.Path("$id").String()
	}

	_, err = sqldb.Exec("INSERT into ecureuil.LOGS (BUCKETNAME,USERNAME,ACTION,PREVIOUSDATA,NEWDATA,JSONID) VALUES ($1,$2,$3,$4,$5,$6)",
		bucketname, username, action, PreviousData, NewData, id)

	if err != nil {
		logger.Error(err.Error())
	}

}

/*createJSONSQLFieldName This function is require to convert JSON format identifier to POSTGRESQL
format for JSONB identifier.  For example  marc.work.phone change to data->'marc'->'work'->>'phone'
*/
func createJSONSQLFieldName(field string) string {
	if field == "" {
		return field
	}
	idx := strings.Split(field, ".")
	if len(idx) <= 1 {
		return "DATA->>'" + idx[0] + "'"
	}
	s := "DATA"
	for i := 0; i < len(idx); i++ {
		if i != len(idx)-1 {
			s += "->'" + idx[i] + "'"
		} else {
			s += "->>'" + idx[i] + "'"
		}
	}
	return s
}

/*DBCreateIndex this is to create an index on a json property

input:
	packet.Key = Name of Index to be created
	packet.SearchField = name of column to index	(contact.phone, or phone)

output:
	message that the index is created or error message

*/
func DBCreateIndex(packet *MsgClientCmd) ([]byte, error) {

	logger.Trace(packet.Username + " request to create index " + string(packet.Key) + " for field " + packet.SearchField)

	// Check if the user has rights

	access, err := UserHasRight([]byte(packet.Username), []byte(packet.Password), "createindex")
	if err != nil {
		logger.Warn("Access denied: User " + packet.Username + " create index" + err.Error())
		return PrepMessageForUser("Internal error while accessing your access rights creating"), err
	}

	if access == false {
		logger.Warn("Access denied: User " + packet.Username + " create index" + err.Error())
		return PrepMessageForUser("You do not have access rights to create index"), nil
	}

	// if here the user has access granted
	logger.Trace("Create Index access granted.")

	q := "CREATE INDEX IF NOT EXISTS ecureuil_" + packet.Key + " ON ecureuil.JSONOBJECTS(data->>'$bucketname', (" + createJSONSQLFieldName(packet.SearchField) + "))"

	logger.Trace("Create index: " + q)
	_, err = sqldb.Exec(q)

	if err != nil {
		return PrepMessageForUser("Error while creating index: " + err.Error()), nil
	}

	DBLog("", packet.Username, "CREATEINDEX", []byte(""), []byte(packet.Key))

	return PrepMessageForUser("Index created"), nil
}

/*DBDropIndex this is to drop reate an index on a json property

input:
	packet.Key = Name of Index to be dropped

output:
	message that the index is dropped or error message

*/
func DBDropIndex(packet *MsgClientCmd) ([]byte, error) {

	logger.Trace(packet.Username + " request to drop index " + string(packet.Key))

	// Check if the user has rights

	access, err := UserHasRight([]byte(packet.Username), []byte(packet.Password), "dropindex")
	if err != nil {
		logger.Warn("Access denied: User " + packet.Username + " drop index" + err.Error())
		return PrepMessageForUser("Internal error while accessing your access rights creating"), err
	}

	if access == false {
		logger.Warn("Access denied: User " + packet.Username + " drop index" + err.Error())
		return PrepMessageForUser("You do not have access rights to drop index"), nil
	}

	// if here the user has access granted
	logger.Trace("DROP Index granted to " + packet.Username)

	_, err = sqldb.Exec("DROP INDEX " + packet.Key)

	if err != nil {
		return PrepMessageForUser("Error while dropping index " + packet.Key), nil
	}

	DBLog("", packet.Username, "DROPINDEX", []byte(""), []byte(packet.Key))

	return PrepMessageForUser("Index " + packet.Key + " dropped"), nil

}

/*DBListIndex this is to drop reate an index on a json property

input:
	packet.Key = Name of Index to be dropped

output:
	message that the index is dropped or error message

*/
func DBListIndex(packet *MsgClientCmd) ([]byte, error) {

	logger.Trace(packet.Username + " request to list all indexes " + string(packet.Bucketname))

	// Check if the user has rights

	access, err := UserHasRight([]byte(packet.Username), []byte(packet.Password), "listindex")
	if err != nil {
		logger.Warn("Access denied: User " + packet.Username + " list index" + err.Error())
		return PrepMessageForUser("Internal error while accessing your access rights"), err
	}

	if access == false {
		logger.Warn("Access denied: User " + packet.Username + " list index" + err.Error())
		return PrepMessageForUser("You do not have access rights to list index"), nil
	}

	// if here the user has access granted
	logger.Trace("LIST Indexes for " + packet.Username)

	rows, err := sqldb.Query("select indexname from pg_indexes where indexname like 'ecureuil_%' AND tablename = 'jsonobjects';")

	if err != nil {
		return PrepMessageForUser("Error while listing indexes " + packet.Key), nil
	}

	buffer := new(bytes.Buffer)

	// what type of information user want to extract?

	buffer.WriteString("{\"action\":\"readindexes\", \"indexes\" : [")

	var result string
	var count int

	for rows.Next() {
		var indexname string

		err = rows.Scan(&indexname)
		if err != nil {
			logger.Error(err.Error())
			return nil, err
		}
		if count <= 0 {
			result += "\"" + indexname + "\""
		} else {
			result += "," + "\"" + indexname + "\""
		}
		count++

	}

	buffer.WriteString(result)

	buffer.WriteString("]}")

	DBLog("", packet.Username, "LISTINDEX", []byte(""), []byte(""))

	return buffer.Bytes(), nil

}

type tquery struct {
	Fieldname  string   `json:"property"`
	Type       string   `json:"type"`
	Searchtype string   `json:"st"`
	Values     []string `json:"values"`
	LogicOp    string   `json:"logic"`
}

/*DBQuery advance query for a getBucket

Func Query (bucketname, query { items: [ fieldname, type, searchtype, values [ 1 or 2 ],  logical  ] } )

searchType can be :  “EQ”, “GT”, “GTE”, “LT”, “LTE” “BETWEEN”
type = BIGINT, INT, TEXT, DECIMAL, DOUBLE
Values is text values 1 items except for BETWEEN is two values
Logical is “AND” or “OR” or “” for last item

 packet.bucketname
 packet.key contain the query

CAST(" + createJSONSQLFieldName(packet.SearchField) + " AS " + packet.Field + ")

*/

func buildsubquery(item *tquery) string {

	qt := ""
	if item.Type == "TEXT" {
		qt = "'"
	}

	switch item.Searchtype {

	case "BETWEEN":
		return "CAST(" + createJSONSQLFieldName(item.Fieldname) + " AS " + item.Type + ") BETWEEN " + qt + item.Values[0] + qt + " AND " + qt + item.Values[1] + qt + " " + item.LogicOp
	case "EQ":
		return "CAST(" + createJSONSQLFieldName(item.Fieldname) + " AS " + item.Type + ") = " + qt + item.Values[0] + qt + " " + item.LogicOp
	case "GT":
		return "CAST(" + createJSONSQLFieldName(item.Fieldname) + " AS " + item.Type + ") > " + qt + item.Values[0] + qt + " " + item.LogicOp
	case "GTE":
		return "CAST(" + createJSONSQLFieldName(item.Fieldname) + " AS " + item.Type + ") >= " + qt + item.Values[0] + qt + " " + item.LogicOp
	case "LT":
		return "CAST(" + createJSONSQLFieldName(item.Fieldname) + " AS " + item.Type + ") < " + qt + item.Values[0] + qt + " " + item.LogicOp
	case "LTE":
		return "CAST(" + createJSONSQLFieldName(item.Fieldname) + " AS " + item.Type + ") <= " + qt + item.Values[0] + qt + " " + item.LogicOp
	default:
		return ""
	}
}

func buildQuery(bucketname string, querystring []byte) string {

	validtypes := []string{"BIGINT", "INT", "TEXT", "DECIMAL", "DOUBLE"}
	searchtypes := []string{"EQ", "GT", "GTE", "LT", "LTE", "BETWEEN"}
	logicops := []string{"AND", "OR", ""}

	queryItems := []tquery{}

	err := json.Unmarshal(querystring, &queryItems)
	if err != nil {
		logger.Error("Invalid query items " + string(querystring) + " ERROR = " + err.Error())
		return ""
	}

	subquery := ""

	for i := 0; i < len(queryItems); i++ {

		if queryItems[i].Fieldname == "" {
			return ""
		}
		if !IsStrInArray(queryItems[i].Type, validtypes) {
			return ""
		}
		if !IsStrInArray(queryItems[i].Searchtype, searchtypes) {
			return ""
		}
		if !IsStrInArray(queryItems[i].LogicOp, logicops) {
			return ""
		}
		if queryItems[i].Searchtype == "BETWEEN" && len(queryItems[i].Values) != 2 {
			return ""
		}
		if queryItems[i].Searchtype != "BETWEEN" && len(queryItems[i].Values) != 1 {
			return ""
		}

		subquery += buildsubquery(&queryItems[i]) + " "

		if i+1 == len(queryItems) && queryItems[i].LogicOp != "" {
			// last item can't finish with AND or OR
			return ""
		}

	}

	return subquery
}

/*DBRead extract one item from database
provide table, field and search value
*/
func DBRead(packet *MsgClientCmd) ([]byte, error) {

	logger.Trace(packet.Username + " request read " + packet.Action + " key " + string(packet.Key) + " from " + packet.Bucketname)

	// Check if the user has rights

	access, err := UserHasRight([]byte(packet.Username), []byte(packet.Password), packet.Bucketname+"-read")
	if err != nil || access == false {
		logger.Warn("Access denied: User " + packet.Username + " Find in bucket " + packet.Bucketname + " error: " + err.Error())
		return PrepMessageForUser("Internal error or you do not have access to " + packet.Bucketname), err
	}

	// if here the user has access granted
	logger.Trace("Read " + packet.Action + " granted to " + packet.Username + " " + packet.Bucketname + " " + string(packet.Key))

	buffer := new(bytes.Buffer)

	// what type of information user want to extract?

	buffer.WriteString("{\"action\":\"read\", \"bucketname\": \"" + EscDoubleQuote(string(packet.Bucketname)) + "\", \"items\" : [")

	var sqlquery string

	var rows *sql.Rows

	if packet.Action == "QUERY" {

		logger.Trace("receive query request")

		sqlquery := buildQuery(packet.Bucketname, packet.Data)

		if sqlquery == "" {
			logger.Error("No query could be build")
			return PrepMessageForUser("No query could be build"), nil
		}

		sqlquery = "select DATA FROM ecureuil.jsonobjects WHERE data->>'$bucketname' = $1 AND (" + sqlquery + ") limit $2;"

		logger.Trace("buildquery = " + sqlquery)

		rows, err = sqldb.Query(sqlquery, packet.Bucketname, Configuration.MaxReadItemsFromDB)

	} else if packet.Action == "READALL" {

		sqlquery = "select DATA FROM ecureuil.jsonobjects WHERE data->>'$bucketname' = $1 limit $2;"

		logger.Trace(sqlquery + " " + packet.Bucketname + " " + strconv.Itoa(Configuration.MaxReadItemsFromDB))
		rows, err = sqldb.Query(sqlquery, packet.Bucketname, Configuration.MaxReadItemsFromDB)

	} else if packet.Action == "READONE" {

		sqlquery = "select DATA FROM ecureuil.jsonobjects WHERE data->>'$bucketname' = $1 AND CAST(" + createJSONSQLFieldName(packet.SearchField) + " AS " + packet.Field + ") = $2 limit 1;"

		logger.Trace(sqlquery)
		rows, err = sqldb.Query(sqlquery, packet.Bucketname, packet.Key)

	} else if packet.Action == "READFIND" {

		logger.Trace("Readfind")
		logger.Trace("searchfield=" + packet.SearchField)
		logger.Trace("field=" + packet.Field)

		sqlquery = "select DATA FROM ecureuil.jsonobjects WHERE data->>'$bucketname' = $1 AND CAST(" + createJSONSQLFieldName(packet.SearchField) + " AS " + packet.Field + ") = $2 limit $3;"

		logger.Trace(sqlquery)
		rows, err = sqldb.Query(sqlquery, packet.Bucketname, packet.Key, Configuration.MaxReadItemsFromDB)

	} else if packet.Action == "READRANGE" {

		f := createJSONSQLFieldName(packet.SearchField)

		sqlquery = "select DATA FROM ecureuil.jsonobjects WHERE data->>'$bucketname' = $1 AND CAST(" + f + " AS " + packet.Field + ") BETWEEN $2 AND $3 limit $4;"

		logger.Trace(sqlquery)
		rows, err = sqldb.Query(sqlquery, packet.Bucketname, packet.Key, packet.MaxKey, Configuration.MaxReadItemsFromDB)

	}

	var result string
	var count int

	if err != nil {
		logger.Error(err.Error())
	} else {

		logger.Trace("ready to read rows")

		for rows.Next() {

			var data string

			logger.Trace("read one row")

			err = rows.Scan(&data)
			if err != nil {
				logger.Error(err.Error())
				return nil, err
			}
			if count <= 0 {
				result += data
			} else {
				result += "," + data
			}
			count++
		}
	}

	if rows != nil {
		rows.Close()
	}

	buffer.WriteString(result)

	buffer.WriteString("]}")

	return buffer.Bytes(), err

}

/*DBGetLogs Key = starttime MaxKey = endtime in unix seconds EPOCH
 */
func DBGetLogs(packet *MsgClientCmd) ([]byte, error) {

	logger.Trace(packet.Username + " request read " + packet.Action + " key " + string(packet.Key) + " from " + packet.Bucketname)

	// Check if the user has admin rights

	access, err := UserHasRight([]byte(packet.Username), []byte(packet.Password), "admin")
	if err != nil || access == false {
		logger.Warn("Access denied: User " + packet.Username + " to LOGS error: " + err.Error())
		return PrepMessageForUser("Internal error or you do not have access to LOGS"), err
	}

	// if here the user has access granted
	logger.Trace("Read LOGS granted to " + packet.Username)

	buffer := new(bytes.Buffer)

	buffer.WriteString("{\"action\":\"read\", \"bucketname\":\"LOGS\", \"items\" : [")

	lowdate, err := strconv.ParseInt(packet.Key, 10, 64)
	if err != nil {
		return PrepMessageForUser("Internal error while parsing lowkey"), err
	}

	highdate, err := strconv.ParseInt(packet.MaxKey, 10, 64)
	if err != nil {
		return PrepMessageForUser("Internal error while parsing highkey"), err
	}

	logger.Trace("***************key=" + packet.Key + " maxkey=" + packet.MaxKey)

	var sqlquery string

	var rows *sql.Rows
	/*
		id, timeofaction, jsonid, username, action, previousdata, newdata
	*/
	sqlquery = "select jsonb_build_object('$id', ID," +
		"'timeofaction', timeofaction," +
		"'bucketname', bucketname," +
		"'jsonid', jsonid," +
		"'username', username," +
		"'action', action," +
		"'previousdata', previousdata," +
		"'newdata', newdata) AS DATA FROM ecureuil.logs WHERE ecureuil.logs.timeofaction between $1 AND $2;"

	rows, err = sqldb.Query(sqlquery, time.Unix(lowdate, 0), time.Unix(highdate, 0))

	logger.Trace(sqlquery + " " + time.Unix(lowdate, 0).Format("2006-01-02T15:04:05") + " " + time.Unix(highdate, 0).Format("2006-01-02T15:04:05"))

	var result string
	var count int

	if err != nil {
		logger.Error(err.Error())
	} else {

		logger.Trace("ready to read rows")

		for rows.Next() {

			var data string

			err = rows.Scan(&data)
			if err != nil {
				logger.Error(err.Error())
				return nil, err
			}
			if count <= 0 {
				result += data
			} else {
				result += "," + data
			}
			count++
		}
	}

	if rows != nil {
		rows.Close()
	}

	buffer.WriteString(result)

	buffer.WriteString("]}")

	return buffer.Bytes(), err

}

/*DBDelete user is asking to delete information from the database  make sure he has access to perform delete and reply to user sucess or failure!
delete header and data and broadcast change.
return usermessage and error
*/
func DBDelete(packet *MsgClientCmd, defered bool) ([]byte, error) {

	var err error

	if !defered {

		logger.Trace("Request delete in " + packet.Bucketname + " from " + string(packet.Username))

		// if access is not granted by default then check if the user has rights
		access, err := UserHasRight([]byte(packet.Username), []byte(packet.Password), packet.Bucketname+"-delete")

		if err != nil {
			logger.Warn(packet.Username + " try to delete item from " + packet.Bucketname + " error: " + err.Error())
			return PrepMessageForUser("Internal error while deleting record."), err
		}

		if access == false {
			logger.Warn(packet.Username + " try to delete item from " + packet.Bucketname + " access denied!")
			return PrepMessageForUser("Delete: Access denied."), err
		}

		if packet.Bucketname == string(UserBUCKET) {
			//USERS BUCKET ==========================================================

			err := UserDelete(packet)
			if err != nil {
				// the UserDelete function use the logger no need to pass a message to log.
				logger.Error(packet.Username + " try to delete " + string(packet.Key) + " - " + err.Error())
				return PrepMessageForUser(err.Error()), nil
			}

			logger.Info(packet.Username + " delete user: " + string(packet.Key))
			return PrepMessageForUser("User as been removed."), nil

		}

		if float64(packet.Defered) >= UnixUTCSecs() {
			return DBDeferAction(packet)
		}

	}

	logger.Trace("access granted to delete.")

	_, err = sqldb.Exec("DELETE from ecureuil.JSONOBJECTS WHERE data->>'$id' = $1", packet.Key)

	if err != nil {
		logger.Trace(err.Error())
		return PrepMessageForUser("Error while deleting object for " + packet.Bucketname + " " + err.Error()), nil
	}

	// delete associated defered commands
	_, err = sqldb.Exec("DELETE from ecureuil.DEFEREDCOMMAND WHERE jsonid = $1", packet.Key)

	if err != nil {
		logger.Trace(err.Error())
		return PrepMessageForUser("Error while deleting defered command for object in " + packet.Bucketname + " " + err.Error()), nil
	}

	logger.Trace("Delete command successful")
	return nil, err
}

/*DBUpdate user is asking to store information into a bucket make sure he has access, then perform store and reply to user sucess or failure!
return usermessage and error; Need Username, Password, Bucketname, Key and Data be set in packet.
*/
func DBUpdate(packet *MsgClientCmd, defered bool) ([]byte, error) {

	var err error

	if !defered {

		logger.Trace("request update bucket " + packet.Bucketname)

		// if access if not granted by default then check if the user has rights
		access, err := UserHasRight([]byte(packet.Username), []byte(packet.Password), packet.Bucketname+"-update")
		if err != nil {
			logger.Warn(packet.Username + " update " + packet.Bucketname + " error: " + err.Error())
			return PrepMessageForUser("Error while updating or access denied."), err
		}

		if access == false {
			logger.Warn(packet.Username + " update " + packet.Bucketname + " access denied.")
			return PrepMessageForUser("Access denined."), err
		}

		if float64(packet.Defered) >= UnixUTCSecs() {
			return DBDeferAction(packet)
		}
	}

	logger.Trace("update access granted")

	switch packet.Bucketname {

	case string(UserBUCKET): /* handle special bucket USERS */
		// UserUpdate function does it's own call of logger to log activity.
		// it will also return an error that contain a text for the user.
		err := UserUpdate(packet)
		if err != nil {
			return PrepMessageForUser(err.Error()), err
		}
		return PrepMessageForUser("User saved!"), nil

	default:

		/* chek if status or itemstatus is present in the data json
		do not allow change unless user as admin or statuschange rights
		*/

		jsonParsed, err := gabs.ParseJSON(packet.Data)
		statusexists := jsonParsed.ExistsP("itemstatus") || jsonParsed.ExistsP("status")

		if statusexists {
			// confirm user has admin or statuschange

			access, err := UserHasRight([]byte(packet.Username), []byte(packet.Password), packet.Bucketname+"-statuschange")
			if err != nil {
				logger.Warn(packet.Username + " update " + packet.Bucketname + " error: " + err.Error())
				return PrepMessageForUser("Error while updating or access denied."), err
			}

			if access == false {
				logger.Warn(packet.Username + " no status rights for update " + packet.Bucketname + " error: " + err.Error())
				return PrepMessageForUser("Access denied you can't change the status value."), err
			}
		}

		sqlquery := "UPDATE ecureuil.JSONOBJECTS set data->>'$bucketname' = $5, data->>'$updatedby' = $1, data->>'$updatedtime' = $2, data = $3 WHERE data->>'$id' = $4"
		_, err = sqldb.Exec(sqlquery, packet.Username, uint64(UnixUTCSecs()), SanitizeJSONStrHTML(string(packet.Data)), packet.Key, packet.Bucketname)

		if err != nil {
			logger.Trace(err.Error())
			return PrepMessageForUser("Error  " + err.Error()), nil
		}

		// a broadcast will be emited once the database generate a notify event.

	}

	if err != nil {
		logger.Error(err.Error())
		return []byte(err.Error()), err
	}

	logger.Trace("command " + packet.Action + " successful ")

	return nil, nil

}

/*DBDeferAction ask the server to run an action at a later date and time
 */
func DBDeferAction(packet *MsgClientCmd) ([]byte, error) {

	logger.Trace("Receive defered command: " + packet.Action + " from " + packet.Username + " password = " + packet.Password)

	p, err := json.Marshal(packet)
	if err != nil {
		return PrepMessageForUser("Error unable to marshal: " + err.Error()), nil
	}

	encoded := base64.StdEncoding.EncodeToString(p)
	//decoded, err := base64.StdEncoding.DecodeString(encoded)

	// ID, BUCKETNAME, CREATEDBY, UPDATEDBY, CREATEDTIME, UPDATEDTIME, CREATEDONNETWORK, CREATEDONSERVER, DATA
	sqlquery := "INSERT into ecureuil.DEFEREDCOMMAND (RUNTIME, COMMAND, JSONID) VALUES ($1,$2,$3)"

	_, err = sqldb.Exec(sqlquery, packet.Defered, encoded, packet.Key)

	if err != nil {
		return PrepMessageForUser("Error unable to defer command: " + err.Error()), nil
	}
	return nil, nil
}

/*DBInsert user is asking to store information into a bucket make sure he has access, then perform store and reply to user sucess or failure!
return usermessage and error; Need Username, Password, Bucketname, Key and Data be set in packet.
*/
func DBInsert(packet *MsgClientCmd, defered bool) ([]byte, error) {

	var err error

	if !defered {

		logger.Trace("request update bucket in " + packet.Bucketname)

		// if access if not granted by default then check if the user has rights
		access, err := UserHasRight([]byte(packet.Username), []byte(packet.Password), packet.Bucketname+"-insert")
		if err != nil {
			logger.Warn(packet.Username + " update " + packet.Bucketname + " error: " + err.Error())
			return PrepMessageForUser("Error while updating or access denied."), err
		}

		if access == false {
			logger.Warn(packet.Username + " update " + packet.Bucketname + " access denied.")
			return PrepMessageForUser("Access denined."), err
		}

		if float64(packet.Defered) >= UnixUTCSecs() {
			return DBDeferAction(packet)
		}
	}

	logger.Trace("insert access granted")

	switch packet.Bucketname {

	case string(UserBUCKET): /* handle special bucket USERS */
		// UserUpdate function does it's own call of logger to log activity.
		// it will also return an error that contain a text for the user.
		err := UserUpdate(packet)
		if err != nil {
			return PrepMessageForUser(err.Error()), err
		}
		return PrepMessageForUser("User saved!"), nil

	default:

		var ID string

		if packet.Key != "" {
			if !govalidator.IsUUIDv4(packet.Key) {
				logger.Trace("invalid UUID provided creating a new one")
				ID = uuid.NewV4().String()
			} else {
				logger.Trace("Valid UUID provided for insert")
				ID = packet.Key
			}
		} else {
			ID = uuid.NewV4().String()
		}

		jsonParsed, err := gabs.ParseJSON(packet.Data)
		statusexists := jsonParsed.ExistsP("itemstatus")

		if statusexists {
			// confirm user has admin or statuschange

			access, err := UserHasRight([]byte(packet.Username), []byte(packet.Password), packet.Bucketname+"-statuschange")
			if err != nil {
				logger.Warn(packet.Username + " update " + packet.Bucketname + " error: " + err.Error())
				return PrepMessageForUser("Error while updating or access denied."), err
			}

			if access == false {
				// reset itemstatus to 30
				jsonParsed.SetP("30", "itemstatus")
			}
		}

		statusexists = jsonParsed.ExistsP("status")

		if statusexists {
			// confirm user has admin or statuschange

			access, err := UserHasRight([]byte(packet.Username), []byte(packet.Password), packet.Bucketname+"-statuschange")
			if err != nil {
				logger.Warn(packet.Username + " update " + packet.Bucketname + " error: " + err.Error())
				return PrepMessageForUser("Error while updating or access denied."), err
			}

			if access == false {
				// reset status to 30
				jsonParsed.SetP("30", "status")
			}
		}

		if Configuration.NetworkID == "" {
			Configuration.NetworkID = "00000000-0000-0000-0000-000000000000"
		}

		jsonParsed.SetP(ID, "$id")
		jsonParsed.SetP(packet.Bucketname, "$bucketname")
		jsonParsed.SetP(packet.Username, "$createdby")
		jsonParsed.SetP(packet.Username, "$updatedby")
		jsonParsed.SetP(Configuration.NetworkID, "$createdonnetwork")
		jsonParsed.SetP(Configuration.ID, "$createdonserver")
		jsonParsed.SetP(uint64(UnixUTCSecs()), "$createdtime")
		jsonParsed.SetP(uint64(UnixUTCSecs()), "$updatedtime")

		// ID, BUCKETNAME, CREATEDBY, UPDATEDBY, CREATEDTIME, UPDATEDTIME, CREATEDONNETWORK, CREATEDONSERVER, DATA
		sqlquery := "INSERT into ecureuil.JSONOBJECTS (DATA) VALUES ($1);"

		_, err = sqldb.Exec(sqlquery, SanitizeJSONStrHTML(jsonParsed.String()))

		if err == nil {

			// broadcast will be emit once the database send an event!

		} else {

			logger.Error(err.Error())
			return PrepMessageForUser("Database Error: " + err.Error()), nil

		}
	}

	if err != nil {
		logger.Error(err.Error())
		return []byte(err.Error()), err
	}

	logger.Trace("command " + packet.Action + " successful ")

	return nil, nil

}

/*DropDB  remove database from postgresql usually to rebuild it
 */
func DropDB(host, user, pass *string) string {

	if *host == "" {
		return "Host name not provided -host=xxxx"
	}

	if *user == "" {
		return "Username for postgresql not provided -user=xxxx"
	}

	if *pass == "" {
		return "Password for postfresql not provided -password=xxxx"
	}

	sqldb, err := sql.Open("postgres", "user="+*user+" host="+*host+" password="+*pass+" sslmode=disable")
	if err != nil {
		return err.Error()
	}

	s := "DROP SCHEMA ecureuil CASCADE;"
	fmt.Println(s)
	_, err = sqldb.Exec(s)

	s = "DROP USER ecureuiladmin;"
	_, err2 := sqldb.Exec(s)

	if err != nil {
		return err.Error()
	} else if err2 != nil {
		return err2.Error()
	}

	sqldb.Close()

	return "Schema ecureuil has been removed from the server."
}

/*CreateDB create a database and trigger in postgresql should provide a .sql file instead of functin and should create the user and password.
 */
func CreateDB(host, user, pass *string) string {

	if *host == "" {
		return "Host name not provided -host=xxxx"
	}

	if *user == "" {
		return "Username for postgresql not provided -user=xxxx"
	}

	if *pass == "" {
		return "Password for postfresql not provided -password=xxxx"
	}

	sqldb, err := sql.Open("postgres", "user="+*user+" host="+*host+" password="+*pass+" sslmode=disable")
	if err != nil {
		return err.Error()
	}

	DatabaseUser := "ecureuiladmin"
	AdminPassword := RandomPassword(30)

	fmt.Println("*********************************************************************")
	fmt.Println("username for postgre is: " + DatabaseUser)
	fmt.Println("password for postgre is: " + AdminPassword)
	fmt.Println("*********************************************************************")

	s := "CREATE SCHEMA ecureuil"
	fmt.Println(s)
	_, err = sqldb.Exec(s)
	if err != nil {
		return err.Error()
	}

	s = "CREATE USER " + DatabaseUser + " WITH PASSWORD '" + AdminPassword + "';"
	_, err = sqldb.Exec(s)
	if err != nil {
		fmt.Println(s)
		return err.Error()
	}

	// Set rights

	s = "REVOKE ALL ON SCHEMA ecureuil FROM public;"
	_, err = sqldb.Exec(s)
	if err != nil {
		fmt.Println(s)
		return err.Error()
	}

	s = "GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA ecureuil TO " + DatabaseUser + ";"
	fmt.Println(s)
	_, err = sqldb.Exec(s)
	if err != nil {
		return err.Error()
	}

	s = "ALTER DEFAULT PRIVILEGES IN SCHEMA ecureuil GRANT USAGE, SELECT ON SEQUENCES TO " + DatabaseUser
	fmt.Println(s)
	_, err = sqldb.Exec(s)
	if err != nil {
		return err.Error()
	}

	s = "GRANT ALL PRIVILEGES ON SCHEMA ecureuil TO " + DatabaseUser + ";"
	_, err = sqldb.Exec(s)
	if err != nil {
		fmt.Println(s)
		return err.Error()
	}

	_, err = sqldb.Exec("CREATE TABLE ecureuil.JSONOBJECTS (" +
		"ID uuid NOT NULL primary key," +
		"BUCKETNAME	varchar(64) NOT NULL," +
		"CREATEDBY varchar(64) NOT NULL," +
		"UPDATEDBY varchar(64) NOT NULL," +
		"CREATEDTIME bigint NOT NULL," +
		"UPDATEDTIME bigint NOT NULL," +
		"CREATEDONNETWORK uuid NOT NULL DEFAULT '00000000-0000-0000-0000-000000000000'," +
		"CREATEDONSERVER uuid NOT NULL," +
		"DATA jsonb);")
	if err != nil {
		return err.Error()
	}

	_, err = sqldb.Exec("CREATE TABLE ecureuil.DEFEREDCOMMAND (" +
		"ID BIGSERIAL NOT NULL primary key," +
		"RUNTIME BIGINT NOT NULL," +
		"COMMAND text NOT NULL);") // contain JSON desiralized cmdaction
	if err != nil {
		return err.Error()
	}

	_, err = sqldb.Exec("CREATE INDEX DEFEREDCOMMAND_TIMEIDX ON ecureuil.DEFEREDCOMMAND (RUNTIME);")
	if err != nil {
		return err.Error()
	}

	_, err = sqldb.Exec("CREATE TABLE ecureuil.LOGS (" +
		"ID bigserial NOT NULL primary key," +
		"TIMEOFACTION timestamptz NOT NULL DEFAULT NOW()," +
		"JSONID uuid NOT NULL," +
		"USERNAME varchar(64) NOT NULL," +
		"BUCKETNAME varchar(64)," +
		"ACTION varchar(16) NOT NULL," +
		"PREVIOUSDATA jsonb," +
		"NEWDATA jsonb NOT NULL);")
	if err != nil {
		return err.Error()
	}

	_, err = sqldb.Exec("CREATE INDEX LOGS_TIMEOFACTION ON ecureuil.LOGS (TIMEOFACTION);")
	if err != nil {
		return err.Error()
	}

	_, err = sqldb.Exec("CREATE INDEX LOGS_USERNAME ON ecureuil.LOGS (USERNAME);")
	if err != nil {
		return err.Error()
	}

	_, err = sqldb.Exec("CREATE INDEX LOGS_ACTION ON ecureuil.LOGS (ACTION);")
	if err != nil {
		return err.Error()
	}

	_, err = sqldb.Exec("CREATE INDEX LOGS_JSONID ON ecureuil.LOGS (JSONID);")
	if err != nil {
		return err.Error()
	}

	_, err = sqldb.Exec("CREATE INDEX JSONOBJECTS_BUCKETNAME ON ecureuil.JSONOBJECTS (data->>'$bucketname');")
	if err != nil {
		return err.Error()
	}

	_, err = sqldb.Exec("CREATE INDEX JSONOBJECTS_CREATEDTIME ON ecureuil.JSONOBJECTS (data->>'$createdtime');")
	if err != nil {
		return err.Error()
	}

	_, err = sqldb.Exec("CREATE INDEX JSONOBJECTS_UPDATEDTIME ON ecureuil.JSONOBJECTS (data->>'$updatedtime');")
	if err != nil {
		return err.Error()
	}

	_, err = sqldb.Exec("CREATE INDEX JSONOBJECTS_CREATEDBY ON ecureuil.JSONOBJECTS (data->>'$createdby');")
	if err != nil {
		return err.Error()
	}

	_, err = sqldb.Exec("CREATE INDEX JSONOBJECTS_UPDATEDBY ON ecureuil.JSONOBJECTS (data->>'$updatedby');")
	if err != nil {
		return err.Error()
	}

	_, err = sqldb.Exec("CREATE INDEX JSONOBJECTS_starttime ON ecureuil.JSONOBJECTS ((DATA->>'$starttime'));")
	if err != nil {
		return err.Error()
	}

	_, err = sqldb.Exec("CREATE INDEX JSONOBJECTS_endtime ON ecureuil.JSONOBJECTS ((DATA->>'$endtime'));")
	if err != nil {
		return err.Error()
	}

	_, err = sqldb.Exec("CREATE INDEX JSONOBJECTS_recurrence ON ecureuil.JSONOBJECTS ((DATA->>'$recurrence'));")
	if err != nil {
		return err.Error()
	}

	_, err = sqldb.Exec("CREATE INDEX JSONOBJECTS_status ON ecureuil.JSONOBJECTS ((DATA->>'$status'));")
	if err != nil {
		return err.Error()
	}

	_, err = sqldb.Exec("CREATE INDEX JSONOBJECTS_recurrenceendtime ON ecureuil.JSONOBJECTS ((DATA->>'$recurrenceendtime'));")
	if err != nil {
		return err.Error()
	}

	_, err = sqldb.Exec("CREATE INDEX JSONOBJECTS_template ON ecureuil.JSONOBJECTS (DATA->>'$bucketname', (DATA->>'$status'));")
	if err != nil {
		return err.Error()
	}

	_, err = sqldb.Exec("CREATE INDEX JSONOBJECTS_username ON ecureuil.JSONOBJECTS (DATA->>'$bucketname', (DATA->>'name'));")
	if err != nil {
		return err.Error()
	}

	_, err = sqldb.Exec("CREATE OR REPLACE FUNCTION ecureuil.logtrigger()" +
		"RETURNS trigger AS $$" +
		"DECLARE " +
		"data jsonb;" +
		"notification jsonb;" +
		"bucketname text;" +
		"createdby text;" +
		"updatedby text;" +
		"createdtime bigint;" +
		"updatedtime bigint;" +
		"ID uuid;" +
		"createdonserver uuid;" +
		"createdonnetwork uuid;" +
		"BEGIN " +
		"if (TG_OP = 'DELETE') THEN " +
		"INSERT into ecureuil.LOGS (TIMEOFACTION, BUCKETNAME, JSONID, USERNAME, ACTION, PREVIOUSDATA, NEWDATA) VALUES (NOW(), OLD.data->>'$bucketname', OLD.data->>'$id', OLD.data->>'$updatedby', TG_OP, OLD.data, OLD.data);" +
		"data = OLD.data;" +
		"ID = OLD.data->>'$id';" +
		"bucketname = OLD.data->>'$bucketname';" +
		"createdby = OLD.data->>'$createdby';" +
		"updatedby = OLD.data->>'$updatedby';" +
		"createdtime = OLD.data->>'$createdtime';" +
		"updatedtime = OLD.data->>'$updatedtime';" +
		"createdonserver = OLD.data->>'$createdonserver';" +
		"createdonnetwork = OLD.data->>'$createdonnetwork';" +
		"ELSEIF (TG_OP = 'UPDATE') THEN " +
		"INSERT into ecureuil.LOGS (TIMEOFACTION, BUCKETNAME, JSONID, USERNAME, ACTION, PREVIOUSDATA, NEWDATA) VALUES (NOW(), OLD.data->>'$bucketname', OLD.data->>'$id', NEW.data->>'$updatedby', TG_OP, OLD.data, NEW.data);" +
		"ID = OLD.data->>'$id';" +
		"data = NEW.data;" +
		"bucketname = OLD.data->>'$bucketname';" +
		"createdby = OLD.data->>'$createdby';" +
		"updatedby = NEW.data->>'$updatedby';" +
		"createdtime = OLD.data->>'$createdby';" +
		"updatedtime = NEW.data->>'$updatedby';" +
		"createdonserver = OLD.data->>'$createdonserver';" +
		"createdonnetwork = OLD.data->>'$createdonnetwork';" +
		"ELSIF (TG_OP = 'INSERT') THEN " +
		"INSERT into ecureuil.LOGS (TIMEOFACTION, BUCKETNAME, JSONID, USERNAME, ACTION, NEWDATA) VALUES (NOW(), NEW.data->>'$bucketname', NEW.data->>'$id', NEW.data->>'$updatedby', TG_OP, NEW.data);" +
		"ID = NEW.data->>'$id';" +
		"data = NEW.data;" +
		"bucketname = NEW.data->>'$bucketname';" +
		"createdby = NEW.data->>'$createdby';" +
		"updatedby = NEW.data->>'$updatedby';" +
		"createdtime = NEW.data->>'$createdtime';" +
		"updatedtime = NEW.data->>'$updatedtime';" +
		"createdonserver = NEW.data->>'$createdonserver';" +
		"createdonnetwork = NEW.data->>'$createdonnetwork';" +
		"END IF;" +
		"notification = DATA || jsonb_build_object(" +
		"'action', TG_OP," +
		"'$id', ID," +
		"'$bucketname', BUCKETNAME," +
		"'$createdby', CREATEDBY," +
		"'$updatedby', UPDATEDBY," +
		"'$createdtime', CREATEDTIME," +
		"'$updatedtime', UPDATEDTIME," +
		"'$createdonnetwork', CREATEDONNETWORK," +
		"'$createdonserver', CREATEDONSERVER);" +
		"PERFORM pg_notify('events_ecureuil',notification::text);" +
		"RETURN NULL;" +
		"END;$$ LANGUAGE plpgsql;")

	if err != nil {
		return err.Error()
	}

	_, err = sqldb.Exec("CREATE OR REPLACE FUNCTION ecureuil.useraccess(username text, rightname text) " +
		"RETURNS integer AS " +
		"$BODY$ " +
		"DECLARE " +
		"RightID uuid; " +
		"adminRightID uuid; " +
		"HasRight uuid; " +
		"json_object json; " +
		"item json; " +
		"itemid uuid; " +
		"v_output text[]; " +
		"usergroups jsonb; " +
		"m text[]; " +
		"BEGIN " +
		"select id from ecureuil.jsonobjects where data->>'$bucketname' = 'USERRIGHTS' and data->>'name' = 'admin' into adminRightID; " +
		"IF EXISTS (SELECT 1 FROM ecureuil.jsonobjects WHERE bucketname = 'USERS' AND data->>'name' = username AND  data @> ('{ \"rights\": [ \"' || adminRightID || '\" ] }')::jsonb) THEN " +
		"RETURN 1; " +
		"ELSE " +
		"IF EXISTS (SELECT 1 FROM ecureuil.jsonobjects WHERE bucketname = 'USERS' AND data->>'name' = username AND  data @> ('{ \"rights\": [ \"admin\" ] }')::jsonb) THEN " +
		"	RETURN 1; " +
		"END IF; " +
		"END IF; " +
		"IF EXISTS (SELECT 1 FROM ecureuil.jsonobjects where data->>'$bucketname' = 'USERRIGHTS' and data->>'name' = rightname) THEN " +
		"select id from ecureuil.jsonobjects where data->>'$bucketname' = 'USERRIGHTS' and data->>'name' = rightname into RightID; " +
		"IF EXISTS (SELECT 1 FROM ecureuil.jsonobjects WHERE bucketname = 'USERS' AND data->>'name' = username AND  data @> ('{ \"rights\": [ \"' || RightID || '\" ] }')::jsonb) THEN " +
		"RETURN 1; " +
		"END IF; " +
		"FOR itemid IN SELECT ID FROM ecureuil.jsonobjects WHERE bucketname = 'USERGROUPS' AND data @> ('{ \"rights\": [ \"' || RightID || '\" ] }')::jsonb " +
		"LOOP " +
		"IF EXISTS (SELECT 1 FROM ecureuil.jsonobjects where bucketname = 'USERS' and data->>'name' = username AND data @> ('{ \"group\": [\"' || itemid || '\"]}')::jsonb) THEN " +
		"return 1; -- user has a group that have the right " +
		"END IF; " +
		"END LOOP; " +
		"END IF; " +
		"RETURN 0; " +
		"END; " +
		"$BODY$ " +
		"LANGUAGE plpgsql VOLATILE " +
		"COST 100; ")

	if err != nil {
		return err.Error()
	}

	s = "ALTER FUNCTION ecureuil.logtrigger() OWNER TO postgres;"
	_, err = sqldb.Exec(s)
	if err != nil {
		fmt.Println(s)
		return err.Error()
	}

	s = "CREATE TRIGGER updatejsonId BEFORE INSERT OR UPDATE ON ecureuil.JSONOBJECTS FOR EACH ROW EXECUTE PROCEDURE ecureuil.setjsonid();"
	_, err = sqldb.Exec(s)
	if err != nil {
		fmt.Println(s)
		return err.Error()
	}

	s = "CREATE TRIGGER log_audit AFTER INSERT OR UPDATE OR DELETE ON ecureuil.JSONOBJECTS FOR EACH ROW EXECUTE PROCEDURE ecureuil.logtrigger();"
	_, err = sqldb.Exec(s)
	if err != nil {
		fmt.Println(s)
		return err.Error()
	}

	s = "GRANT SELECT,INSERT,UPDATE,DELETE ON TABLE ecureuil.logs TO " + DatabaseUser + ";"
	_, err = sqldb.Exec(s)
	if err != nil {
		fmt.Println(s)
		return err.Error()
	}

	s = "GRANT SELECT,INSERT,UPDATE,DELETE ON TABLE ecureuil.DEFEREDCOMMAND TO " + DatabaseUser + ";"
	_, err = sqldb.Exec(s)
	if err != nil {
		fmt.Println(s)
		return err.Error()
	}

	s = "GRANT SELECT,INSERT,UPDATE,DELETE ON TABLE ecureuil.jsonobjects TO " + DatabaseUser + ";"
	_, err = sqldb.Exec(s)
	if err != nil {
		fmt.Println(s)
		return err.Error()
	}

	s = "GRANT EXECUTE ON FUNCTION ecureuil.logtrigger() TO " + DatabaseUser + ";"
	_, err = sqldb.Exec(s)
	if err != nil {
		fmt.Println(s)
		return err.Error()
	}

	s = "GRANT EXECUTE ON FUNCTION ecureuil.useraccess(text,text) TO " + DatabaseUser + ";"
	_, err = sqldb.Exec(s)
	if err != nil {
		fmt.Println(s)
		return err.Error()
	}

	s = "ALTER TABLE ecureuil.jsonobjects OWNER TO " + DatabaseUser + ";"
	_, err = sqldb.Exec(s)
	if err != nil {
		fmt.Println(s)
		return err.Error()
	}

	fmt.Println("Disconnecting...")

	sqldb.Close()

	/* Initialize database, create USERS and CONFIGURATION Buckets with  default values */
	ConfigurationINIT()

	return "Database has been successfully created"

}

func runDeferedEvents() {

	sqldb, err := sql.Open("postgres", connstring)
	if err != nil {
		logger.Error(err.Error())
	}

	defer sqldb.Close()

	for {

		// monitor event that need to be executed
		time.Sleep(time.Second * time.Duration(30)) // twice a minute

		v := uint64(UnixUTCSecs())

		// ID, BUCKETNAME, CREATEDBY, UPDATEDBY, CREATEDTIME, UPDATEDTIME, CREATEDONNETWORK, CREATEDONSERVER, DATA

		rows, err := sqldb.Query("SELECT ID, COMMAND from ecureuil.DEFEREDCOMMAND WHERE RUNTIME BETWEEN 1 AND $1", v)

		if err != nil {
			logger.Error(err.Error())
		}

		if rows != nil {

			for rows.Next() {

				var ID uint64
				var COMMAND string

				err = rows.Scan(&ID, &COMMAND)
				if err != nil {
					logger.Error(err.Error())
				}

				// now remove this defered command from the database
				_, err := sqldb.Exec("DELETE FROM ecureuil.DEFEREDCOMMAND WHERE ID = $1", ID)
				if err != nil {
					logger.Error(err.Error())
				}

				// create object
				decoded, err := base64.StdEncoding.DecodeString(COMMAND)

				packet := MsgClientCmd{}

				err = json.Unmarshal([]byte(decoded), &packet)
				if err != nil {
					logger.Error("Error unable to unmarshal: " + err.Error())
				}

				///* only DELETE,UPDATE,INSERT are allowed to be defered

				logger.Trace("Running defered command " + packet.Action + " from " + packet.Username)

				if packet.Action == "UPDATE" {

					_, err := DBUpdate(&packet, true)
					if err != nil {
						logger.Error(err.Error())
					}

				} else if packet.Action == "INSERT" {

					_, err = DBInsert(&packet, true)
					if err != nil {
						logger.Error(err.Error())
					}

				} else if packet.Action == "DELETE" {

					_, err = DBDelete(&packet, true)
					if err != nil {
						logger.Error(err.Error())
					}

				}
			}

			rows.Close()
		}

	}

}

/*

Check for recurrence of item to automatically change their status.

	property.status  			0=pending, 1=active, 2=completed
	property.starttime			unix utc time when this item must become Active.
	property.endtime 			unix utc time when this item must become Inactive
	property.recurrence			see dates.go to get detals about this JSON object
	property.recurrence.StartDate             uint64  `json:"startdate"`             // Date to start Recurrence. Note that time and time zone information is NOT used in calculations
	property.recurrence.Duration              int     `json:"duration"`              // in seconds
	property.recurrence.RecurrencePatternCode string  `json:"recurrencepatterncode"` // D for daily, W for weekly, M for monthly or Y for yearly
	property.recurrence.RecurEvery            int16   `json:"recurevery"`            // number of days, weeks, months or years between occurrences
	property.recurrence.YearlyMonth           *int16  `json:"yearlymonth"`           // month of the year to recur (applies only to RecurrencePatternCode: Y)
	property.recurrence.MonthlyWeekOfMonth    *int16  `json:"monthlyweekofmonth"`    // week of the month to recur. used together with MonthlyDayOfWeek (applies only to RecurrencePatternCode: M or Y)
	property.recurrence.MonthlyDayOfWeek      *int16  `json:"monthlydayofweek"`      // day of the week to recur. used together with MonthlyWeekOfMonth (applies only to RecurrencePatternCode: M or Y)
	property.recurrence.MonthlyDay            *int16  `json:"monthlyday"`            // day of the month to recur (applies only to RecurrencePatternCode: M or Y)
	property.recurrence.WeeklyDaysIncluded    *int16  `json:"weeklydaysincluded"`    // integer representing binary values AND'd for 1000000-64 (Sun), 0100000-32 (Mon), 0010000-16 (Tu), 0001000-8 (W), 0000100-4 (Th), 0000010-2 (F), 0000001-1 (Sat). (applies only to RecurrencePatternCode: M or Y)
	property.recurrence.DailyIsOnlyWeekday    *bool   `json:"dailyisonlyweekday"`    // indicator that daily recurrences should only be on weekdays (applies only to RecurrencePatternCode: D)
	property.recurrence.EndByDate             *uint64 `json:"endbydate"`             // date by which all occurrences must end by. Note that time and time zone information is NOT used in calculations


Search all items that Endtime is > now and there status is not Completed and recurence = 0
	    Set status to Completed

Search all items that Endtime is > now and there status is not Completed and recurrence <> 0 and recurrenceendtime <= now
	    Set status to Completed

Search all items that Endtime is > now and there status is not Completed and recurrence <> 0 and recurrenceendtime > now
		Calculate and set next starttime and endtime and set status to Pending

Search all items that now is between Starttime and Endtime and their status is not active
		Set status to active

Search all items that Starttime is < now and their status is not pending
	- set status to pending

*/

func runMonitorStatusStartEnd() {

	sqldb, err := sql.Open("postgres", connstring)
	if err != nil {
		logger.Error(err.Error())
	}

	defer sqldb.Close()

	for {

		// monitor event that need to be executed
		time.Sleep(time.Second * time.Duration(30)) // twice a minute

		query := "DELETE FROM ecureuil.LOGS WHERE TIMEOFACTION::date < (CURRENT_DATE - INTERVAL '365 days')::date;"
		_, err := sqldb.Exec(query)

		if err != nil {
			logger.Error(query)
			logger.Error(err.Error())
		}

		logger.Trace(" ")

		v := uint64(UnixUTCSecs())

		/* Set Status to 2 (close)
		   IF status is not 2 and endtime is <= now and item is not recurrent
		*/

		query = "UPDATE ecureuil.JSONOBJECTS set DATA = jsonb_set(data, '{$status}', '2', true) "
		query += "WHERE  DATA->>'$status' IS NOT NULL AND CAST(DATA->>'$status' AS INT) <> 2 AND CAST(DATA->>'$endtime' AS BIGINT) <= $1 AND DATA->>'$recurrence' is NULL;"

		_, err = sqldb.Exec(query, v)

		if err != nil {
			logger.Error(query)
			logger.Error(err.Error())
		}

		/* Set Status to 2 (close)
		   IF status is not close, endtime as pass, item is recurrent but recurrence enddate has pass.
		*/

		query = "UPDATE ecureuil.JSONOBJECTS set DATA = jsonb_set(data, '{$status}', '2', true) "
		query += "WHERE DATA->>'$status' IS NOT NULL AND CAST(DATA->>'$status' AS INT) <> 2 AND CAST(DATA->>'$endtime' AS BIGINT) <= $1 AND DATA->>'$recurrence' is NOT NULL AND " +
			"CAST(DATA->'$recurrence'->>'endbydate' AS BIGINT) <= $1;"

		_, err = sqldb.Exec(query, v)

		if err != nil {
			logger.Error(query)
			logger.Error(err.Error())
		}

		/*
		   Search all items that Endtime is > now and there status is not Completed and recurrence <> 0 and recurrenceendtime > now
		   		Calculate and set next starttime and endtime and set status to Pending
		*/
		query = "SELECT ID, CAST(DATA->>'$recurrence' AS TEXT) AS recurrence FROM ecureuil.JSONOBJECTS " +
			"WHERE CAST(DATA->>'$endtime' AS BIGINT) < $1 AND DATA->>'$status' IS NOT NULL AND CAST(DATA->>'$status' AS INT) <> 2 AND DATA->>'$recurrence' is NOT NULL AND CAST(DATA->'$recurrence'->>'endbydate' AS BIGINT) > $1;"

		rows, err := sqldb.Query(query, v)

		if err != nil {
			logger.Error(err.Error())
		} else {

			for rows.Next() {

				var id string
				var recurrence string

				err = rows.Scan(&id, &recurrence)
				if err != nil {
					logger.Error(err.Error())
				}

				/*
					calcule next recurrence return 0,0 if there is no more recurrence

					recurrence should contain a json string:

					StartDate             uint64  `json:"startdate"`             // Date to start Recurrence. Note that time and time zone information is NOT used in calculations
					Duration              int     `json:"duration"`              // in seconds
					RecurrencePatternCode string  `json:"recurrencepatterncode"` // D for daily, W for weekly, M for monthly or Y for yearly
					RecurEvery            int16   `json:"recurevery"`            // number of days, weeks, months or years between occurrences
					YearlyMonth           *int16  `json:"yearlymonth"`           // month of the year to recur (applies only to RecurrencePatternCode: Y)
					MonthlyWeekOfMonth    *int16  `json:"monthlyweekofmonth"`    // week of the month to recur. used together with MonthlyDayOfWeek (applies only to RecurrencePatternCode: M or Y)
					MonthlyDayOfWeek      *int16  `json:"monthlydayofweek"`      // day of the week to recur. used together with MonthlyWeekOfMonth (applies only to RecurrencePatternCode: M or Y)
					MonthlyDay            *int16  `json:"monthlyday"`            // day of the month to recur (applies only to RecurrencePatternCode: M or Y)
					WeeklyDaysIncluded    *int16  `json:"weeklydaysincluded"`    // integer  binary values AND'd together for 1000000-64 (Sun), 0100000-32 (Mon), 0010000-16 (Tu), 0001000-8 (W), 0000100-4 (Th), 0000010-2 (F), 0000001-1 (Sat). (applies only to RecurrencePatternCode: M or Y)
					DailyIsOnlyWeekday    *bool   `json:"dailyisonlyweekday"`    // indicator that daily recurrences should only be on weekdays (applies only to RecurrencePatternCode: D)
					EndByDate             *uint64 `json:"endbydate"`             // date by which all occurrences must end by. Note that time and time zone information is NOT used in calculations


				*/

				// the function take a string and unmarshal it to a struct
				start, end := GetNextDatePeriod(recurrence)

				if start+end == 0 {

					logger.Trace("recurence completed!")

					// there is no more recurrence before the recurrenceendtime
					query = "UPDATE ecureuil.JSONOBJECTS set DATA = jsonb_set(data, '{$status}', '2', true)  WHERE data->>'$id' = $1"
					_, err = sqldb.Exec(query, id)
					if err != nil {
						logger.Error(err.Error())
					}

				} else {

					// first close item to generate email
					query = "UPDATE ecureuil.JSONOBJECTS set DATA = jsonb_set(data, '{$status}', '2', true)  WHERE data->>'$id' = $1"
					_, err = sqldb.Exec(query, id)
					if err != nil {
						logger.Error(err.Error())
					}

					// now change dates
					query = `UPDATE ecureuil.JSONOBJECTS set DATA = DATA || '{"status": 0,"starttime":` +
						strconv.FormatUint(start, 10) + `,"endtime":` + strconv.FormatUint(end, 10) + `}' WHERE data->>'$id' = $1`

					logger.Trace(query)

					_, err = sqldb.Exec(query, id)
					if err != nil {
						logger.Error(err.Error())
					}

					logger.Trace("recurence renewed! now start=" + strconv.FormatUint(start, 10) + " end: " + strconv.FormatUint(end, 10))

				}

			}
		}

		if rows != nil {
			rows.Close()
		}

		/* Set status to 1 (active) for any items that have a now between starting and endtime and status is not equal to 1
		 */

		query = "UPDATE ecureuil.JSONOBJECTS set DATA = jsonb_set(data, '{$status}', '1', true) "
		query += "WHERE DATA->>'$status' IS NOT NULL AND CAST(DATA->>'$status' AS INT) <> 1 AND $1 BETWEEN CAST(DATA->>'$starttime' AS BIGINT) AND CAST(DATA->>'$endtime' AS BIGINT);"

		_, err = sqldb.Exec(query, v)
		if err != nil {
			logger.Error(err.Error())
		}

		/* Set status to 0 (pending)
		   IF items have a starting time > now and status not equal to 0 and endtime is > now
		*/

		query = "UPDATE ecureuil.JSONOBJECTS set DATA = jsonb_set(data, '{$status}', '0', true) "
		query += "WHERE CAST(DATA->>'$status' AS INT) > 0 AND CAST(DATA->>'$starttime' AS BIGINT) > $1 AND CAST(DATA->>'$endtime' AS BIGINT) > $1;"

		_, err = sqldb.Exec(query, v)
		if err != nil {
			logger.Error(err.Error())
		}

	}

}

/*

 */

func waitForNotification(l *pq.Listener) error {

	for {

		select {

		case n := <-l.Notify:
			/* check for data */

			logger.Trace("Received data from POSTGRE channel [", n.Channel, "] :")

			// Validate the payload
			Notification := TNotification{}
			err := json.Unmarshal([]byte(n.Extra), &Notification)
			if err != nil {
				return err
			}

			// Here we know we have a valid notification from POSTGRESQL
			// Only Broadcast to users DELETE, INSERT and UPDATE

			if Notification.Action == "DELETE" || Notification.Action == "UPDATE" || Notification.Action == "INSERT" {

				logger.Trace("Receive event from POSTGRESQL: " + Notification.Action + " for bucket: " + Notification.Bucketname + " " + string(n.Extra))
				BroadcastPut(Notification.Bucketname, n.Extra)

			}

			if Notification.Action == "UPDATE" || Notification.Action == "INSERT" {
				GenerateEmailTemplate(Notification.Bucketname, n.Extra)
			}

			return nil

		case <-time.After(90 * time.Second):
			logger.Trace("Received no events for 90 seconds, checking connection")
			go func() {
				l.Ping()
			}()
			return nil
		}

	}

}
