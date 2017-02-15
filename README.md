# Ecureuil v0.5 
![](http://owlso.net/img/owlso.png)

[![GoDoc](https://godoc.org/github.com/asdine/storm?status.svg)](https://godoc.org/github.com/asdine/storm)

Ecureuil is a realtime, open-source backend for JavaScript apps.  It serve static content and provide access to database via websockets.  You javascript app can register to automatically receive data when as it change in the database.  This framework is made to build web app that require access to data in real-time.

Rapidly build and deploy web or mobile apps using a simple JavaScript API. 

Built for simplicity it allow you to start developing rapidly without having to learn a complicated API.  

Ecureuil is fully open source and develop in GO-LANG and use postgresql as the backend database.

In addition to the examples below, see also the [examples in the GoDoc](https://godoc.org/github.com/owlso/ecureuil#pkg-examples).



*** Not for production this framework require more testing.

## Example

	<include ecureuil.js>
    connect
    onevent ...
    
    

## Download

Pre configure [VirtualBox] VM with postgre and ecureuil pre-installed 64 bits Centos 
https://owlso.net/ecureuil.ovf
root password is "ecureuil"
postgresql with user postgre password "ecureuil"


Ecureuil source code  
https://github.com/owlso/ecureuil 



## Table of Contents

<!-- TOC depthFrom:2 depthTo:6 withLinks:1 updateOnSave:0 orderedList:0 -->

- [Getting Started](#getting-started)

- [The Javascript API](#simple-orm)
	
    * System Info 
	- [getconfig](#declare-your-structures)
	- [putconfig](#save-your-object)
	- [getusers](#simple-queries)
	- [time](#advanced-queries)
	- [stats](#transactions)

	* Login Management
	- [registerevent](#options)
	- [unregisterevent](#node-options)
	- [login](#node-options)
	- [logout](#node-options)
	- [deletedefered](#node-options)
	- setemailalert
	- connect
	- 
	
    * Read Data
		- [one](#simple-keyvalue-store)
		- [many](#simple-keyvalue-store)
		- [range](#simple-keyvalue-store)
		- [all](#simple-keyvalue-store)
	
    * Write Data

		- [insert](#boltdb)
		- [update](#boltdb)
		- [delete](#boltdb)
	
	* Indexes 

		- [indexcreate](#boltdb)
		- [indexlist](#boltdb)
		- [indexdrop](#boltdb)
		
	* Events 

		- [onupdate](#boltdb)
		- [oninsert](#boltdb)
		- [ondelete](#boltdb)
		- [ondelete](#boltdb)
		- [onconfiguration](#boltdb)
		-   this.onConnect = null;
            this.onLogin = null;
            this.onDisconnect = null;
            this.onRead = null;
            this.onMessage = null;
           this.onUsers = null;
            this.onError = null;
            this.onStats = null;
            this.onTime = null;
            this.onIndexes = null;
            this.onRegisterEvent = null;
         
      properties:
         
          this.connected = false;
            this.username =  "";
            this.logged = false;
            this.registerevents =  [];
	        this.serversocket = null;
            
        
	
- [The Users Rights](#simple-orm)
    
    users rights are not saved in the SQL they are saved in the local database 
    ecureuil.db
    
    each users can be granted rights to read or write in buckets and special rights
    
    - admin
    - download
    - stats-read
    - users-delete
    - xxxxxx-read  to read a specific bucket
    - xxxxxx-write to write in a specific bucket
    - xxxxxx-delete to delete items from a specific bucket 
    - 
    
- [The LOG's](#simple-orm)

	- [Users activiy](#simple-orm)

		All users activities are logged into a specific table in the SQL they are 
        keep for X days where x is the configurable number of days.  This can be set in the system configuration.

	- [System errors](#simple-orm)

		All error that occures in the ecureuil framework are saved into a subfolder call logs/ with rotating log files.
        

    
    
<!-- /TOC -->

## Getting Started

Step 1 install Postgresql here more information if required 
Step 2 install you favorite linux distro, create a folder and copy ecureuil executable in it.
Step 3 run sudo ./ecureuil -createdb -host=192.168.56.101 -user=postgres -password=bitnami

host is the IP of your POSTGRESQL 
user is a user already existing in the database that have the rights to create database 
password is self explanatory.

This user and password will not be saved, ecureuil will generate a username “ecureuiladmin” with a random 30 characters password, this password will be save locally into ecureuil.db this file also contain the access rights for each users and the system configuration.  Only the user running ecureuil and the system admin should have access to this file.

If you want to uninstall ecureuil from your postgresql you can do so with the follwing command:
sudo ./ecureuil -dropdb -host=192.168.56.101 -user=postgres -password=bitnami



## API functions

```go
import "github.com/asdine/storm"
```

## Open a database

Quick way of opening a database
```go
db, err := storm.Open("my.db")

defer db.Close()
```

`Open` can receive multiple options to customize the way it behaves. See [Options](#options) below

## Simple ORM

### Declare your structures

```go
type User struct {
  ID int // primary key
  Group string `storm:"index"` // this field will be indexed
  Email string `storm:"unique"` // this field will be indexed with a unique constraint
  Name string // this field will not be indexed
  Age int `storm:"index"`
}
```

The primary key can be of any type as long as it is not a zero value. Storm will search for the tag `id`, if not present Storm will search for a field named `ID`.

```go
type User struct {
  ThePrimaryKey string `storm:"id"`// primary key
  Group string `storm:"index"` // this field will be indexed
  Email string `storm:"unique"` // this field will be indexed with a unique constraint
  Name string // this field will not be indexed
}
```

Storm handles tags in nested structures with the `inline` tag

```go
type Base struct {
  Ident bson.ObjectId `storm:"id"`
}

type User struct {
	Base      `storm:"inline"`
	Group     string `storm:"index"`
	Email     string `storm:"unique"`
	Name      string
	CreatedAt time.Time `storm:"index"`
}
```
SERVER SIDE SECURITY

Each time a Secure Websocket connection is created  {
	Create Client Object 
	Add to Client User name and Password default nil, nil
	On Receive Login Command Verify Password if valid change Client.Username + Client.Password 
}

The user must provide a user name and password each time he open a connection but the connection can last forever.  

LOGOUT command simply set the client user name and password to nil for the connection object.

There is no sessionID or any credential pass back to the client.

If the WebSocket connection close the session is lost and credential must be provided again over Secure Websocket!


SPECIAL BUCKETS:
	TEMPLATES
		{“bucketname”:”templates”, “bucket”:”name of bucket ie. INCIDENTS”,  “status”: 0,  “body”:”html template”, “subject”:”html template”} 




SPECIAL PROPERTIES:

property.id

property.status  			
0=pending, 1=active, 2=completed this property can change automatically if starttime and endtime are set 

property.starttime			
unix utc time when this item must become Active.

property.endtime 			
unix utc time when this item must become Inactive

property.recurrence			
int 0=not recurrent, 1=every monday, etc.

type Recurrentdate struct {
	StartDate             uint64  `json:"startdate"`             // Date to start Recurrence. Note that time and time zone information is NOT used in calculations
	Duration              int     `json:"duration"`              // in seconds
	RecurrencePatternCode string  `json:"recurrencepatterncode"` // D for daily, W for weekly, M for monthly or Y for yearly
	RecurEvery            int16   `json:"recurevery"`            // number of days, weeks, months or years between occurrences
	YearlyMonth           *int16  `json:"yearlymonth"`           // month of the year to recur (applies only to RecurrencePatternCode: Y)
	MonthlyWeekOfMonth    *int16  `json:"monthlyweekofmonth"`    // week of the month to recur. used together with MonthlyDayOfWeek (applies only to RecurrencePatternCode: M or Y)
	MonthlyDayOfWeek      *int16  `json:"monthlydayofweek"`      // day of the week to recur. used together with MonthlyWeekOfMonth (applies only to RecurrencePatternCode: M or Y)
	MonthlyDay            *int16  `json:"monthlyday"`            // day of the month to recur (applies only to RecurrencePatternCode: M or Y)
	WeeklyDaysIncluded    *int16  `json:"weeklydaysincluded"`    // integer representing binary values AND'd together for 1000000-64 (Sun), 0100000-32 (Mon), 0010000-16 (Tu), 0001000-8 (W), 0000100-4 (Th), 0000010-2 (F), 0000001-1 (Sat). (applies only to RecurrencePatternCode: M or Y)
	DailyIsOnlyWeekday    *bool   `json:"dailyisonlyweekday"`    // indicator that daily recurrences should only be on weekdays (applies only to RecurrencePatternCode: D)
	EndByDate             *uint64 `json:"endbydate"`             // date by which all occurrences must end by. Note that time and time zone information is NOT used in calculations
}

/*Recurrentdate is a type of data structure require for each JSON object that need to have recurrence.

The Recurrence struct is modeled after the recurring schedule data model used by both Microsoft Outlook and Google Calendar for recurring appointments. Just like Outlook, you can pick from Daily ("D"), Weekly ("W"), Monthly ("M") and Yearly ("Y") recurrence pattern codes. Each of those recurrence patterns then require the corresponding information to be filled in.

All recurrences:

StartDateTime - start time of the appointment. Should be set to the first desired occurence of the recurring appointment
RecurrencePatternCode - D: daily, W: weekly, M: monthly or Y: yearly
RecurEvery - number defining how many days, weeks, months or years to wait between recurrences
EndByDate (optional) - date by which recurrences must be done by
NumberOfOccurrences (optional) - data for UI which can be used to store the number of recurrences. Has no effect in calculations though. EndByDate must be calculated based on NumberOfOccurrences
Recurrence Pattern Code D (daily)

DailyIsOnlyWeekday (optional) - ensure that daily occurrences only fall on weekdays (M, T, W, Th, F)
Recurrence Pattern Code W (weekly)

WeeklyDaysIncluded - binary value (converted to int16) to indicate days included (e.g. 0101010 or decimal 42 would be MWF). Each of the individual days are bitwise AND'd together to get the value.
Sunday - 64 (1000000)
Monday - 32 (0100000)
Tuesday - 16 (0010000)
Wednesday - 8 (0001000)
Thursday - 4 (0000100)
Friday - 2 (0000010)
Saturday - 1 (0000001)
Recurrence Pattern Code M (monthly)

MonthlyWeekOfMonth - week of the month to recur on. e.g. Thanksgiving is always on the 4th week of the month. Must be used together with MonthlyDayOfWeek
MonthlyDayOfWeek - day of the week to recur on (0=Sunday, 1=Monday, 2=Tuesday, 3=Wednesday, 4=Thursday, 5=Friday, 6=Saturday). Must be used together with MonthlyWeekOfMonth OR
MonthlyDay - day of the month to recur on. e.g. 5 would recur on the 5th of every month
Recurrence Pattern Code Y (yearly)

YearlyMonth - month of the year to recur on (1=January, 2=February, 3=March, 4=April, 5=May, 6=June, 7=July)
MonthlyWeekOfMonth - week of the month to recur on. e.g. Thanksgiving is always on the 4th week of the month. Must be used together with MonthlyDayOfWeek
MonthlyDayOfWeek - day of the week to recur on (0=Sunday, 1=Monday, 2=Tuesday, 3=Wednesday, 4=Thursday, 5=Friday, 6=Saturday). Must be used together with MonthlyWeekOfMonth OR
MonthlyDay - day of the month to recur on. e.g. 5 would recur on the 5th of every month




Properties:

    connected                                           true if websocket connection eastablish with the Ecureuil server
    username                                            username currently use for the connection (password was previously provided)
    logged                                              true if login has been completed and accepted by the server
    registerevents                                      array list of all the buckets we are monitoring.

Functions:

    connect(host)                                       create a connection to one Ecurreuil server
    time()                                              request time on the server
    stats()                                             request stats about the server such as CPU%, disk space, and Storm DB information
    registerevent(eventname)                            request to monitor change on a bucket
    unregisterevent(eventname)                          request to stop monitoring changes on a bucket
    login(username, password)                           provide credential to the server
    logout()                                            logoff this remove credential and user connection remain with guess priviliege only
    insert(bucketname, object, defered)                 insert an object in the database if propery .id is not set a UUIDv4 will be generated
    update(bucketname, object, defered)                 update an object in the database .id property must be set 
    delete(bucketname, object, defered)                 delete an object in the database .id property will be use to find the object
    all(bucketname)                                     request all data for a specific bucket 
    one(bucketname, searchfield, value, fieldtype)      request one item that match searchfield = value
	many(bucketname, searchfield, value, fieldtype)                request all items that match searchfield = value
	range(bucketname, searchfield, minvalue, maxvalue, fieldtype)  request all items that have a searchfield value between minvalue and maxvalue
	indexcreate(indexname, field)                       request to create an index for a specific fieldname in a bucket
    indexdrop(indexname, field)                         request to drop an index from a specific bucket
    indexlist()                                         request the list of existing indexes
    getconfig()
    putconfig(json)
    getusers()                                          request list of all users need admin login
		
Events:

    onError(string)                                         event call on error return string description of error 
    onDisconnect()                                          event call when websocket is disconnected
    onConnect()                                             event call when websocket is connected
	onLogin(e.response.username, result);                   event call after failed or successful login 
    onMessage(e.response.message);                          event call when Ecureuil want to display a message to the user
    onRead(bucketname, e.data);                             event call after data has been reed from the database
    onStats(e.response.server, e.response.database);        event call when statistics about the server and db are received
    onUpdate(e.response);                                   event call when data has been updated in the database
    onDelete(e.response);                                   event call when data has been deleted from the database
    onInsert(e.response);                                   event call when data has been inserted in the database
    onTime(e.response);                                     event call when receive time on the server
    onIndexes(e.response.indexes);                          event call when reeive the list of existing index on the dastabase
    onRegisterEvent(e.response.bucketname);                 event call after sucessful registration to events of one bucket
    onUnregisterEvent(e.response.bucketname);               event call after succesful unregistration of events for one bucket
    onConfiguration(config)                                 event call when we received the back end configuration  
    onUsers(users)                                          event call when we received the list of users   

*/

## License

MIT

## Credits

- [Marc Gauthier](https://github.com/owlso)
