/*Package models - users.go

This file contain functions to control the USERS bucket, user Bucket contain
the list of username and password but most important the rights that these
users have.  The rights are defined as follow BUCKETNAME-action.
______________________________________________________________________________

 OWLSO - Overwatch Link and Service Observer.
______________________________________________________________________________

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

examples:

	INCIDENTS-read, INCIDENTS-write, INCIDENTS-delete
	read, write and delete allow to manipulate information within a Bucket.

	db-download allow a user to download the entire database thru HTTPS

	admin allow user to perform all

	password-reset allow user to update USERS and change their password

	The database always need to contain at least on admin user (admin) if
	you delete this user by accident, the next time the program restart,
	it will generate a new user with admin rights using the default
	username and password.

	UNlike other Bucket the USER bucket as a fix predetermine structure
	for the JSON object it store.  Structure sent by the FRONT-END must
	respect the structure define by TUserv1 see bellow.

______________________________________________________________________________


Revision:

	30 Mar 2017 - Change Storm DB to PostgreSQL
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
	uuid "github.com/satori/go.uuid"
	"golang.org/x/crypto/bcrypt"
)

/*UserBUCKET contain the valid name for storing USERS information,
the bucket VALUES is use to store the information but the key is
created with the UserBUCKET name.
*/
var UserBUCKET = []byte("USERS") // Bucket name for users in BoltDB

/*DefaultadminNAME this is the name of the default user that will be
create if there is no admin user present in the database.  So if
the admin user is delete by accident a new one is generated.
*/
var DefaultadminNAME = []byte("owlsoadmin") // default username for admin user

/*DefaultadminPASSWORD default password for the default DefaultadminNAME
 */
var DefaultadminPASSWORD = []byte("p@ssw0rd") // default password for admin user

/*DBLogin check if username and password are OK
 */
func DBLogin(packet *MsgClientCmd) ([]byte, error) {

	logger.Trace("Request for LOGIN credential check for " + packet.Username)

	result, err := VerifyUserPassword([]byte(packet.Username), []byte(packet.Password))

	if !result || err != nil {

		if err == nil {
			err = errors.New("Invalid or incorrect password ")
		}

		// Log error
		logger.Warn(err.Error())

		// Send Response to user!
		return []byte("{ \"action\":\"login\", \"result\":\"failed\", \"username\":\"" + packet.Username + "\"" + ", \"error\":\"" + err.Error() + "\"}"), err

	}

	settings := ""
	rights := ""

	user := userFind(string(packet.Username))
	if user != nil {

		settings = string(user.Settings)
		if settings == "" {
			settings = "{}"
		}

		r, err := json.Marshal(user.Rights)

		if err != nil {
			rights = "[]"
		} else {
			logger.Info("*******************************************************************************" + string(r))
			rights = string(r)
		}

		logger.Info("{ \"action\":\"login\", \"result\":\"success\", \"settings\":" + settings + ", \"username\":\"" + packet.Username + "\"}")

		// sucessfull login sent the good news to the user.
		return []byte("{ \"action\":\"login\", \"result\":\"success\", \"settings\":" + settings + ", \"rights\":" + rights + ", \"username\":\"" + packet.Username + "\"}"), nil
	}

	// Send Response to user!
	return []byte("{ \"action\":\"login\", \"result\":\"failed\", \"username\":\"" + packet.Username + "\"" + ", \"error\":\"" + err.Error() + "\"}"), err

}

/*DBUserSettings check if username and password are OK
 */
func DBUserSettings(packet *MsgClientCmd) ([]byte, error) {

	logger.Trace("Request change user settings for " + packet.Username)

	if packet.Username == "" {
		err := errors.New("No username logged in")
		return nil, err
	}

	result, err := VerifyUserPassword([]byte(packet.Username), []byte(packet.Password))

	if !result || err != nil {

		if err == nil {
			err = errors.New("Invalid or incorrect password ")
		}

		// Log error
		logger.Warn(err.Error())

		// Send Response to user!
		//return []byte("{ \"action\":\"login\", \"result\":\"failed\", \"username\":\"" + packet.Username + "\"" + ", \"error\":\"" + err.Error() + "\"}"), err

	}

	user := userFind(string(packet.Username))
	if user != nil {

		oldsettings := map[string]interface{}{}
		json.Unmarshal(user.Settings, &oldsettings)

		newsettings := map[string]interface{}{}
		json.Unmarshal([]byte(packet.Data), &newsettings)

		for k, v := range newsettings {
			oldsettings[k] = v
		}

		user.Settings, err = json.Marshal(oldsettings)

		saveUser(user, packet.Username)

		// sucessfull SAVE.
		return nil, nil
	}

	return nil, nil // user was not found

}

func userFind(name string) *TUser {

	sqlquery := "select data from ecureuil.jsonobjects WHERE data->>'$bucketname' = '" + string(UserBUCKET) + "' AND data->>'name' = $1;"

	logger.Trace(sqlquery)

	rows, err := sqldb.Query(sqlquery, name)

	defer rows.Close()

	if err != nil {
		logger.Error(err.Error())
		return nil
	}

	for rows.Next() {

		data := ""
		err = rows.Scan(&data)

		if err != nil {
			logger.Error(err.Error())
			return nil
		}

		user := TUser{}
		err := json.Unmarshal([]byte(data), &user)

		if err != nil {
			logger.Error(err.Error())
			return nil
		}
		return &user

	}
	return nil
}

/*GetUsers this function is use to update the user information in the database.

Handle the request to Update user information.
if user is an admin only an admin can update the account.
if newpassword is set user requesting must have admin or password-reset right, or
requesting a password reset for himself.
if update is trying to add the following special rights admin, db-download, password-reset
the user must have admin account.
*/
func GetUsers(packet *MsgClientCmd) ([]byte, error) {

	// check if user requesting the action has admin right.
	_, err := UserHasRight([]byte(packet.Username), []byte(packet.Password), "admin")

	if err != nil {
		logger.Warn("Unable to verify rights of " + packet.Username + " " + err.Error())
		return PrepMessageForUser("Unable to verify rights of current user"), errors.New("Unable to verify rights of current user")
	}

	buffer := new(bytes.Buffer)

	// what type of information user want to extract?

	buffer.WriteString("{\"action\":\"read\", \"bucketname\": \"" + string(UserBUCKET) + "\", \"items\" : [")

	// get all users information except for password info.

	sqlquery := "select data - 'passwordhash' - 'newpassword' from ecureuil.jsonobjects where data->>'$bucketname' = '" + string(UserBUCKET) + "';"

	logger.Trace(sqlquery)

	rows, err := sqldb.Query(sqlquery)

	defer rows.Close()

	if err != nil {
		logger.Error(err.Error())
		return nil, nil
	}

	start := 0

	for rows.Next() {

		if start > 0 {
			buffer.WriteString(",")
		}

		data := ""
		err = rows.Scan(&data)

		if err != nil {
			logger.Error(err.Error())
			return nil, nil
		}

		start = 1

		buffer.WriteString(data)

	}

	buffer.WriteString("]}")

	return buffer.Bytes(), nil

}

/*UserUpdate this function is use to update the user information in the database.

Handle the request to Update user information.
if user is an admin only an admin can update the account.
if newpassword is set user requesting must have admin or password-reset right, or
requesting a password reset for himself.
if update is trying to add the following special rights admin, db-download, password-reset
the user must have admin account.
*/
func UserUpdate(packet *MsgClientCmd) error {

	item := TUser{}

	// deserialize object to confirm it's valid
	//************************************************
	errcode := json.Unmarshal(packet.Data, &item)
	if errcode != nil {
		logger.Error("Unable to unmarshal data provided by User: " + packet.Username + " for bucket " + packet.Bucketname + " error: " + errcode.Error())
		return errors.New("Unreadable data")
	}

	// check if user requesting the action has admin right.
	admin, err := UserHasRight([]byte(packet.Username), []byte(packet.Password), "admin")

	if err != nil {
		logger.Warn("Unable to verify rights of " + packet.Username + " " + err.Error())
		return errors.New("Unable to verify rights of current user")
	}

	/* Check if the user we are going to modify has admin rights */
	//************************************************************

	if err := VerifyUserHasRight([]byte(item.ID), "admin"); err == nil {

		// The user we are trying to update has admin rights make sure the
		// user actioning the request also had admin rights.

		if admin == false {
			logger.Warn(packet.Username + " try to modify " + string(packet.Key) + " (admin) access denied")
			return errors.New("You required admin rights to modify this user")
		}

	} else {
		// user to be updated does not have admin right fine we can edit him!
	}

	logger.Trace("Checking for new password")

	/* Check if a password reset is requested.
	//************************************************/
	if item.NewPassword != "" {

		/* check for self password reset */
		if packet.Username != packet.Key {

			// we are trying to reset someone else password /* do we have admin rights?  */

			if admin == false {

				/* do we have password-reset right? */
				passwordreset, err := UserHasRight([]byte(packet.Username), []byte(packet.Password), "password-reset")

				if err != nil {
					logger.Warn(packet.Username + " try to reset password of " + string(packet.Key) + " error: " + err.Error())
					return errors.New("You do not have the rights to reset this user password")
				}

				if passwordreset == false {
					logger.Warn(packet.Username + " try to reset password of " + string(packet.Key) + " access denied")
					return errors.New("You do not have the rights to reset this user password")
				}
			}

		}

		// ok we are allow to reset the password, continue...

	}

	logger.Trace("Checking for special rights")

	/* Check if user is trying to add special rights: admin, password-reset, db-download */
	//************************************************************************************
	if IsStrInArray("admin", item.Rights) || IsStrInArray("db-download", item.Rights) || IsStrInArray("password-reset", item.Rights) {

		/* do we have admin rights?  */
		if admin == false {

			logger.Warn(packet.Username + " try to give special rights to " + string(packet.Key) + " access denied")
			return errors.New("You do not have the rights to set special rights")

		}

	}

	logger.Trace("Saving user")

	// save user in database.
	//************************************************
	err = UserSave(&item, item.NewPassword != "", packet.Username)

	if err == nil {
		logger.Info(packet.Username + " has modify " + string(item.ID))

		//DBLog("USERS", packet.Username, packet.Action, []byte(""), packet.Data)

	} else {
		logger.Trace("Save error: " + err.Error())
	}
	return err

}

/*UserSave this function is use to save a user structure in the database,

  Save or update a user in the database, important in order to preserve user password hash,
  item must be reed from the database and all item except password hash need to be
  overwritten.

  This function does not verify if rights to make the change are valid, this process need
  to be perform before this function is called.
*/
func UserSave(user *TUser, PasswordHasChanged bool, Username string) error {

	if PasswordHasChanged {

		// Password Hash is in clear in the struct, replace it with an Hash value.

		var err error
		user.PasswordHash, err = bcrypt.GenerateFromPassword([]byte(user.NewPassword), bcrypt.DefaultCost)
		if err != nil {
			return err
		}

		// because we just changed the password we can overwrite all fields
		return saveUser(user, Username)

	}

	item := userFind(string(user.ID))

	if item != nil {
		// Copy the user password into the new user information
		copy(user.PasswordHash, item.PasswordHash)
		user.NewPassword = "" // do not save the password in clear in the database.
	}

	return saveUser(user, Username)

}

/*VerifyUserHasRight internal function to verify if a user a a right does not check for password.
 */
func VerifyUserHasRight(username []byte, rightname string) error {

	// call postgre function UserAccess "select UserAccess ('username', 'name of right');" return 1 if has right else 0

	sqlquery := "select ecureuil.UserAccess ('" + string(username) + "', '" + rightname + "');"

	logger.Trace(sqlquery)

	rows, err := sqldb.Query(sqlquery)

	defer rows.Close()

	if err != nil {
		return err
	}

	for rows.Next() {

		data := 0
		err = rows.Scan(&data)

		if err != nil {
			logger.Error(err.Error())
			return err
		}

		if data == 1 {
			logger.Trace("============access granted!")
			return nil
		}
		break
	}

	logger.Trace("============access denied!")

	return errors.New("User " + string(username) + " does not have access to " + rightname)

}

/*UserHasRight verify if user password is correct and if user has rights.
 */
func UserHasRight(username, password []byte, rightname string) (bool, error) {

	if username == nil {
		username = []byte("guess")
	}

	logger.Trace("verifiying password")

	access, err := VerifyUserPassword(username, password)
	if err != nil || !access {
		return false, err
	}

	logger.Trace("password verified")

	err = VerifyUserHasRight(username, rightname)
	if err != nil {
		return false, err
	}

	return true, nil

}

/*VerifyUserPassword verifiy if a username and password are correct.
 */
func VerifyUserPassword(username, password []byte) (bool, error) {

	if len(password) <= 0 {
		return false, errors.New("No password provided")
	}

	user := userFind(string(username))
	if user == nil {
		logger.Error("user not found")
		return false, nil
	}

	errcode := bcrypt.CompareHashAndPassword(user.PasswordHash, password)
	if errcode != nil {
		// password do not matched!
		return false, errcode
	}

	return true, nil

}

/*UsersINIT
 */
func UsersINIT() {

}

func saveUser(u *TUser, Username string) error {

	newuser, err := json.Marshal(u)
	if err != nil {
		return err
	}

	user := userFind(u.ID)

	if user == nil {
		// insert
		logger.Trace("insert")

		// start with empty json
		jsonParsed, _ := gabs.ParseJSON(newuser)

		id := uuid.NewV4().String()

		jsonParsed.SetP(id, "$id")
		jsonParsed.SetP("USERS", "$bucketname")
		jsonParsed.SetP(Username, "$updatedby")
		jsonParsed.SetP(Configuration.NetworkID, "$createdonnetwork")
		jsonParsed.SetP(Configuration.ID, "$createdonserver")
		jsonParsed.SetP(uint64(UnixUTCSecs()), "$createdtime")
		jsonParsed.SetP(uint64(UnixUTCSecs()), "$updatedtime")

		sqlquery := "INSERT INTO ecureuil.JSONOBJECTS (data) values ($1)"

		_, err = sqldb.Exec(sqlquery, jsonParsed.String())

	} else {
		// update
		logger.Trace("update")

		sqlquery := "UPDATE ecureuil.JSONOBJECTS set data = $2 WHERE ecureuil.JSONOBJECTS.data->>'$bucketname' = '" + string(UserBUCKET) + "' AND ecureuil.JSONOBJECTS.DATA->>'name' = $1;"

		_, err = sqldb.Exec(sqlquery, u.ID, string(newuser)) // u.ID contain name!

	}

	return err
}

/*UserDelete this function is use to delete a user from the database.

    Username to Delete is in packet.Key.
	User need either USERS-delete or admin righs to be able to delete users.
	if User to be deleted has admin rights than user requesting action must be an admin as well.
*/
func UserDelete(packet *MsgClientCmd) error {

	// check if user has delete users rights.
	access, err := UserHasRight([]byte(packet.Username), []byte(packet.Password), "USERS-delete")

	// check if user has admin rights.
	admin, err := UserHasRight([]byte(packet.Username), []byte(packet.Password), "admin")

	if !admin && !access {
		logger.Warn(packet.Username + " try to delete " + string(packet.Key) + " access denied.")
		return errors.New("You do not have rights to delete USERS")
	}

	// check if the user to be deleted is an admin, only admin can delete admin!

	if err = VerifyUserHasRight([]byte(packet.Key), "admin"); err == nil {

		// The user we are trying to delete has admin rights
		if !admin {
			logger.Warn("User modification blocked, " + packet.Username + " want to modified  " + string(packet.Key) + " (admin) access denied.")
			return errors.New("You required admin rights to modify this user")
		}
	}

	// Proceed with the deletion of the user.

	sqlquery := "DELETE FROM ecureuil.JSONOBJECTS WHERE ecureuil.JSONOBJECTS.data->>'$bucketname' = '" + string(UserBUCKET) + "' AND ecureuil.JSONOBJECTS.DATA->>'name' = $1;"

	_, err = sqldb.Exec(sqlquery, string(packet.Key)) // Key contain name!

	return err

}
