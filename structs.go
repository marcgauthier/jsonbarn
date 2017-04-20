package models

import "encoding/json"

/*

  Storm Tag Keywords:

  	index		: index the field
  	id			: field is the primary key
  	unique		: index is a unique index
  	inline  	: use to concatenate nested structs
  	increment 	: auto increase the value for int's

*/

/*TUserGroup USERGROUP bucket contain information about groups, groups can have rights
and are use for planned incident.  Each group can be selected to approved Planned Incident
This bucket should not sync with other serverStat
*/
type TUserGroup struct {
	ID string `storm:"id,unique" json:"name"` // name of the group

	Rights []string `json:"rights"` // user of this group are givens theses extra rights

	RequireForPlannedIncident bool `json:"requireforplannedincident"` // this group approved planned incident

	Description string `json:"description"` // descrption

	POC string `json:"poc"` // who to contact to approve access

	AltPOC string `json:"altpoc"` // alternate contact for approving access

}

/*TUser USERS bucket structure use to store information about each user that will

have access to this database

This bucket should not sync with other server

*/
type TUser struct {
	ID string `storm:"id,unique" json:"name"` // user name

	Contact string `json:"contact"` // how to contact this user.

	PasswordHash []byte `json:"passwordhash"` // password hash value

	Rights []string `json:"rights"` // Rights["INCIDENTS-read"]

	NewPassword string `json:"newpassword"` // Use only to do a password change

	Groups []string `json:"group"` // What group the user is part of

	Settings []byte `json:"settings"` // save user setting for client-side
}

/*TNotification contain the structure to hold item received from postgresql
 */
type TNotification struct {
	Bucketname       string          `json:"bucket"`
	Action           string          `json:"action"`
	CreatedBy        string          `json:"createdby"`
	UpdatedBy        string          `json:"updatedby"`
	CreatedTime      uint64          `json:"createdtime"`
	UpdatedTime      uint64          `json:"updatedtime"`
	CreatedonNetwork string          `json:"createdonnetwork"`
	CreatedonServer  string          `json:"createdonserver"`
	Data             json.RawMessage `json:"data"` // contain the JSON serialized object to be saved, it will be HTML Sanitized
}
