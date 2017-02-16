# Ecureuil v0.5 
![](http://owlso.net/img/ecureuil.png)

[![GoDoc](https://godoc.org/github.com/asdine/storm?status.svg)](https://godoc.org/github.com/asdine/storm)

Ecureuil is a realtime, open-source backend for JavaScript apps.  It serve static content and provide access to database via websockets.  You javascript app can register to automatically receive data when as it change in the database.  This framework is made to build web app that require access to data in real-time.

Rapidly build and deploy web or mobile apps using a simple JavaScript API. 

Built for simplicity it allow you to start developing rapidly without having to learn a complicated API.  

Ecureuil is fully open source and develop in GO-LANG and use postgresql as the backend database.

In addition to the examples below, see also the [examples in the GoDoc](https://godoc.org/github.com/owlso/ecureuil#pkg-examples).



*** Not for production this framework require more testing.



## Download

* Virtual Box Appliance
	- Pre configure appliance with [Centos 6 - 64 bits](http://owlso.net/ecureuil_centos_64.ovf), postgre and ecureuil pre-installed  
	- **root password** ecureuil
	- **postgres password** ecureuil
	
	
* Source code
	- https://github.com/owlso/ecureuil 



## Table of Contents

<!-- TOC depthFrom:2 depthTo:6 withLinks:1 updateOnSave:0 orderedList:0 -->

- [Getting Started](#getting-started)

- [The Javascript API](#javascriptapi)
	
    * **Functions** 
    
    	* System Info 
			- [getconfig](#getconfig)
			- [putconfig](#putconfig)
			- [getusers](#getusers)
			- [time](#time)
			- [stats](#stats)

		* Login Management
			- [registerevent](#registerevent)
			- [unregisterevent](#unregisterevent)
			- [login](#login)
			- [logout](#logout)
			- [deletedefered](#deletedefered)
			- [setemailalert](#setemailalert)
			- [connect](#connect)
		
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
		
	* **Events** 

		- [onupdate](#boltdb)
		- [oninsert](#boltdb)
		- [ondelete](#boltdb)
		- [ondelete](#boltdb)
		- [onconfiguration](#boltdb)
		- [onconnect](#onconnect)
		- [onlogin](#onlogin)
		- [ondisconnect](#ondisconnect)
		- [onread](#onread)
		- [onmessage](#onmessage)
		- [onusers](#onusers)
		- [onerror](#onerror)
		- [onstats](#onstats)
		- [ontime](#ontime)
		- [onindexes](#onindexes)
		- [onregisterevent](#onregisterevent)
		
	* **Properties** 
		- [connected](#propertyconnected)
		- [username](#propertyusername)
		- [logged](#propertylogged)
		- [registerevents](#propertyregisterevents)
		- [serversocket](#propertyserversocket)	
    
	* **Other info** 
  	- [Reserved Properties within your JSON objects](#simple-orm)
    
- [**Users Rights**](#simple-orm)
    
- [**Log's**](#simple-orm)
    
- [**License**](#simple-orm)
    
- [**Credits**](#simple-orm)
    
    

    
    
<!-- /TOC -->

## Getting Started

The easiest way to get started is to download the pre configure appliance, if you want to build your own server follow theses steps.

1. Install you favorite linux distro, create a folder and copy [ecureuil executable](http://github.com/owlso/ecureuil) in it.
1. Install [Postgre sql database](https://www.postgresql.org/) version 9.6 or higher here more information if required 
1. Build the database: sudo ./ecureuil -createdb -host=192.168.56.101 -user=postgres -password=bitnami

	**Where**
	Host is the IP of your postgre database 
User is a user already existing in the database that have rights to create new database and new user
password for this user.

	*Provided username and password are not going to be saved, ecureuil will generate a new username **ecureuiladmin** with a random 30 characters password, this password will be save locally into ecureuil.db this file also contain the access rights for each users and the system configuration.  Only the user running ecureuil and the system admin should have access to this file.*

	At any point if you want to uninstall ecureuil from your postgres you can do so with the following command:
sudo ./ecureuil -dropdb -host=192.168.56.101 -user=postgres -password=bitnami



## The Javascript API

In order to use ecureuil framework within you javascript application you simply need to import the javascript client.  This file is available in the public/ folder on github.  See bellow for the list of functions, events and properties you can access.
```go
<script src="js/ecureuil.js"></script>
var ecureuil = new Ecureuil();

```

### **function getconfig();**
```go
var ecureuil = new Ecureuil();
ecureuil.connect("wss://yourwebsite.com/wss/");
... once connection is eastablished you can call
ecureuil.getconfig();
```
-	This function will contact the server and request the system configuration.  You must be logged-on with a user that has admin rights in order for the server to respond with the configuration.  The event **onconfiguration** will be fired once it is received.

-	The configuration will be sent as a JSON object see the event **onconfiguration** for more details.

### **function putconfig(configuration);**
```go
var ecureuil = new Ecureuil();
ecureuil.connect("wss://yourwebsite.com/wss/");
... once connection is eastablished you can call
configuration = {};
configuration.smtp.ip = 10.0.0.1;
...
ecureuil.putconfig(configuration);
```
-	This function will contact the server and request to overwrite the system configuration with the object provided.  You must be logged-on with a user that has admin rights in order for the server to accept your request.  See the event **onconfiguration** to view the structure of the configuration object.

-	An acknowledge will be receive over **onmessage** if the new configuration has been accepted.


### **function getusers();**
```go
var ecureuil = new Ecureuil();
ecureuil.connect("wss://yourwebsite.com/wss/");
... once connection is eastablished you can call
...
ecureuil.getusers();
```
-	This function will contact the server and request the list of all users and their properties.  This allowed you to list all the user in the system and modify their configuration and access rights.  You must be logged-on with a user that has admin rights in order for the server to accept your request.  The event **onusers** will be fired once the server reply with the list of users.  See the event **onusers** to view the structure of the users object.

### **function time();**
```go
var ecureuil = new Ecureuil();
ecureuil.connect("wss://yourwebsite.com/wss/");
... once connection is eastablished you can call
...
ecureuil.time();
```
-	This function will contact the server and request the current time in UTC+0 in unix EPOCH.

### **function stats();**
```go
var ecureuil = new Ecureuil();
ecureuil.connect("wss://yourwebsite.com/wss/");
... once connection is eastablished you can call
...
ecureuil.stats();
```
-	This function will contact the server and request statistics about the ecureuil server.  It will return physical information about the hardware such as disk free space, cpu utilization and information about the local database that contain users rights.  You must be logged-on with a user that has admin rights in order for the server to accept your request.  The event **onstats** will be fired once the server reply with the statistics.  See the event **onstats** to view the structure of the statistics object return by the server.


### **function login(username, password);**
```go
var ecureuil = new Ecureuil();
ecureuil.connect("wss://yourwebsite.com/wss/");
... once connection is eastablished you can call
ecureuil.login(username, password);
```
-	This function provide the backend server with credential, once the credential are verified you will be granted access rights.  The event **onlogin** will be fired once the server reply to confirmed you have provided correct credentials.  Once the login is accepted it will remain until the websocket connection is closed.  If the websocket connection close you will have to provide credential again.


### **function logout();**
```go
var ecureuil = new Ecureuil();
ecureuil.connect("wss://yourwebsite.com/wss/");
... once connection is eastablished you can call
ecureuil.login(username, password);
...
ecureuil.logout();
```
-	This function tell the backend to release all access rights granted to this websocket connection and grant access rights to a guess user.  The connection is not lost, only the access rights are discarded.  The function does not generate an event.

### **function registerevent(bucketname);**
```go
var ecureuil = new Ecureuil();
ecureuil.connect("wss://yourwebsite.com/wss/");
... once connection is eastablished you can call
ecureuil.login(username, password);
...
ecureuil.registerevent(bucketname);
```
-	This function tell the backend to generate an event every time data in the bucket "bucketname" is updated, deleted or added.  Bucket are equivalent or collection in the NoSQL world or table in SQL database.  You must be logged-on with a user that have read access to the bucket your are requesting access to.  Events fired are **onupdate**, **oninsert** and **ondelete**.

### **function unregisterevent(bucketname);**
```go
var ecureuil = new Ecureuil();
ecureuil.connect("wss://yourwebsite.com/wss/");
... once connection is eastablished you can call
ecureuil.login(username, password);
...
ecureuil.registerevent(bucketname);
ecureuil.unregisterevent(bucketname);
```
-	This function work with registerevent, once you no longer want to receive event about changes inside a specific bucket you can unregister.  This function does not generate any event.

### **function setemailalert(emailaddress, bucketnames);**
```go
var ecureuil = new Ecureuil();
ecureuil.connect("wss://yourwebsite.com/wss/");
... once connection is eastablished you can call
ecureuil.login(username, password);
...
ecureuil.setemailalert("marc.gauthier3@gmail.com", ["CHAT-MESSAGES", "GENERAL-INFO"]);
```
-	Ecureuil broadcast change to the database via websocket but also send email alert when data is inserted or updated.  This function allow to request that a specific email receive alert about changes in specific buckets.  If the database already have a list of buckets to monitor for this specific email address they will be replaced with this new list of buckets.  The change will occur in two phase.  The call to setemailalert will generate a confirmation email to the user.  Once the user open his email and click on the confirmation link than the changes become permanent.


### **function connect(url);**
```go
var ecureuil = new Ecureuil();
ecureuil.connect("wss://yourwebsite.com/wss/");
... once connection is eastablished you can call
ecureuil.login(username, password);
...
ecureuil.one("INCIDENTS", "starttime", 2321232, "BIGINT");
```
-	This will generate a websocket connection between the Ecureuil backend and the client library.  Once the connection is eastablished the event **onconnect** will be fired.  Your url must always start with wss Ecureuil only support secure connection and must end with /wss/ this is the path that the mux on the backend is expecting to indicate that a websocket connection is requested..


### **function one(bucketname, searchfield, value, fieldtype);**
```go
var ecureuil = new Ecureuil();
ecureuil.connect("wss://yourwebsite.com/wss/");
... once connection is eastablished you can call
ecureuil.login(username, password);
...
ecureuil.one("INCIDENTS", "starttime", 2321232, "BIGINT");
```
-	This function will search the database for the first entry that will match the request.  You need to provide the bucketname where to search, the fieldname the value you are looking for and what type of field the fieldname represent; valid options are INT, BIGINT, TEXT, DECIMAL, DOUBLE. Once data is found and return the event **onread** will be fired.  




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
		

### Events

		- [onupdate](#boltdb)
		- [oninsert](#boltdb)
		- [ondelete](#boltdb)
		- [ondelete](#boltdb)
		- [onconfiguration](#boltdb)
		- [onconnect](#onconnect)
		- [onlogin](#onlogin)
		- [ondisconnect](#ondisconnect)
		- [onread](#onread)
		- [onmessage](#onmessage)
		- [onusers](#onusers)
		- [onerror](#onerror)
		- [onstats](#onstats)
		- [ontime](#ontime)
		- [onindexes](#onindexes)
		- [onregisterevent](#onregisterevent)
		
        
### Properties
		
		- [connected](#propertyconnected)
		- [username](#propertyusername)
		- [logged](#propertylogged)
		- [registerevents](#propertyregisterevents)
		- [serversocket](#propertyserversocket)
	




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

    [**Reserved Properties within your JSON objects**](#simple-orm)
    
    [**Users Rights**](#simple-orm)
    
    [**Log's**](#simple-orm)
    
    [**License**](#simple-orm)
    
    [**Credits**](#simple-orm)
    
    s

- [The LOG's](#simple-orm)

	- [Users activiy](#simple-orm)

		All users activities are logged into a specific table in the SQL they are 
        keep for X days where x is the configurable number of days.  This can be set in the system configuration.

	- [System errors](#simple-orm)

		All error that occures in the ecureuil framework are saved into a subfolder call logs/ with rotating log files.
        


## License

MIT

## Credits

- [Marc Gauthier](https://github.com/owlso)
