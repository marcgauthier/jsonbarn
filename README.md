# JsonBarn v1.0 
![](http://owlso.net/img/jsonbarn.png)

[![GoDoc](https://godoc.org/github.com/marcgauthier/JsonBarn?status.svg)](https://godoc.org/github.com/marcgauthier/jsonbarn)

**JsonBarn** is a **light** self-hosted backend as a service for building realtime JavaScript (web) apps, Open Source and free.  It was build to allow developper to quickly build Single page App with real-time support without having to write a single line of backend code.  JsonBarn is developped in golang and relied on postgre Sql for it's database backend.  JsonBarn is a light framework it does not have many features just what you need!

- Serve static content over https, http is automatically redirected to https
- Provide access to database via client library api over secure websockets.
- Real-time support, clients can select to register to receive changes commit to the database in near real-time.
- Built for simplicity it allow you to start developing rapidly without having to learn a complicated API.  
- JsonBarn is fully open source and develop in GOLANG.


**Dependencies**
- logger https://github.com/antigloss/go
- govalidator https://github.com/asaskevich/govalidator
- Gods https://github.com/emirpasic/gods
- Gorilla http://www.gorillatoolkit.org/pkg/websocket
- PQ https://github.com/lib/pq
- Bluemonday https://github.com/microcosm-cc/bluemonday
- gocode https://github.com/nsf/gocode
- Calendar https://github.com/EndFirstCorp/calendar
- UUID https://github.com/satori/go.uuid
- GABS https://github.com/Jeffail/gabss

## Download
	
	
* Source code
	- https://github.com/marcgauthier/jsonbarn 



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

		* Login Management
			- [registerevent](#registerevent)
			- [unregisterevent](#unregisterevent)
			- [login](#login)
			- [logout](#logout)
			- [setemailalert](#setemailalert)
			- [connect](#connect)
		
    	* Read Data
			- [one](#one)
			- [many](#many)
			- [range](#range)
			- [all](#all)
			- [query](#query)
	
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
		- [onconnect](#onconnect)
		- [onlogin](#onlogin)
		- [ondisconnect](#ondisconnect)
		- [onread](#onread)
		- [onmessage](#onmessage)
		- [onerror](#onerror)
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
  	- [Serving static content](#static-content)
  	- [Certificates](#certificates)
  	- [Command line flags](#command-line)
    
- [**Users Rights**](#simple-orm)
    
- [**Log's**](#simple-orm)
    
- [**License**](#simple-orm)
    
- [**Credits**](#simple-orm)
    
    

    
    
<!-- /TOC -->

## Getting Started


1. Install you favorite linux distro, create a folder and copy [jsonbarn executable](http://github.com/owlso/jsonbarn) in it.
1. Install [Postgre sql database](https://www.postgresql.org/) version 9.6 or higher.
1. Build the database: sudo ./jsonbarn -createdb -host=xxxxx -user=postgres -password=xxxxx 

	**Where**
	Host is the IP of your postgre database 
User is a user already existing in the database that have rights to create new database and new user
password for this user.

	At any point if you want to uninstall JsonBarn from your postgres you can do so with the following command:
sudo ./JsonBarn -dropdb -host=sql-ip -user=postgres -password=bitnami



## The Javascript API

In order to use JsonBarn framework within you javascript application you simply need to import the javascript client.  This file is available in the public/ folder on github.  See bellow for the list of functions, events and properties you can access.
```go
<script src="js/JsonBarn.js"></script>
var JsonBarn = new JsonBarn();

```

### **function getconfig();**
```go
var JsonBarn = new JsonBarn();
JsonBarn.connect("wss://yourwebsite.com/wss/");
... once connection is eastablished you can call
JsonBarn.getconfig();
```
-	This function will contact the server and request the system configuration.  You must be logged-on with a user that has admin rights in order for the server to respond with the configuration.  The event **onread** will be fired once it is received.

-	The configuration will be sent as a JSON object see the event **onread** for more details.

### **function putconfig(configuration);**
```go
var JsonBarn = new JsonBarn();
JsonBarn.connect("wss://yourwebsite.com/wss/");
... once connection is eastablished you can call
configuration = {};
configuration.smtp.ip = 10.0.0.1;
...
JsonBarn.putconfig(configuration);
```
-	This function will contact the server and request to overwrite the system configuration with the object provided.  You must be logged-on with a user that has admin rights in order for the server to accept your request.  

-	An acknowledge will be receive over **onmessage** if the new configuration has been accepted.


### **function getusers();**
```go
var JsonBarn = new JsonBarn();
JsonBarn.connect("wss://yourwebsite.com/wss/");
... once connection is eastablished you can call
...
JsonBarn.getusers();
```
-	This function will contact the server and request the list of all users and their properties.  This allowed you to list all the user in the system and modify their configuration and access rights.  You must be logged-on with a user that has admin rights in order for the server to accept your request.  The event **onread** will be fired once the server reply with the list of users.  See the event **onread** to view the structure of the users object.

### **function time();**
```go
var JsonBarn = new JsonBarn();
JsonBarn.connect("wss://yourwebsite.com/wss/");
... once connection is eastablished you can call
...
JsonBarn.time();
```
-	This function will contact the server and request the current time in UTC+0 in unix EPOCH.

### **function login(username, password);**
```go
var JsonBarn = new JsonBarn();
JsonBarn.connect("wss://yourwebsite.com/wss/");
... once connection is eastablished you can call
JsonBarn.login(username, password);
```
-	This function provide the backend server with credential, once the credential are verified you will be granted access rights.  The event **onlogin** will be fired once the server reply to confirmed you have provided correct credentials.  Once the login is accepted it will remain until the websocket connection is closed.  If the websocket connection close you will have to provide credential again.


### **function logout();**
```go
var JsonBarn = new JsonBarn();
JsonBarn.connect("wss://yourwebsite.com/wss/");
... once connection is eastablished you can call
JsonBarn.login(username, password);
...
JsonBarn.logout();
```
-	This function tell the backend to release all access rights granted to this websocket connection and grant access rights to a guess user.  The connection is not lost, only the access rights are discarded.  The function does not generate an event.

### **function registerevent(bucketname);**
```go
var JsonBarn = new JsonBarn();
JsonBarn.connect("wss://yourwebsite.com/wss/");
... once connection is eastablished you can call
JsonBarn.login(username, password);
...
JsonBarn.registerevent(bucketname);
```
-	This function tell the backend to generate an event every time data in the bucket "bucketname" is updated, deleted or added.  Bucket are equivalent or collection in the NoSQL world or table in SQL database.  You must be logged-on with a user that have read access to the bucket your are requesting access to.  Events fired are **

**, **oninsert** and **ondelete**.

### **function unregisterevent(bucketname);**
```go
var JsonBarn = new JsonBarn();
JsonBarn.connect("wss://yourwebsite.com/wss/");
... once connection is eastablished you can call
JsonBarn.login(username, password);
...
JsonBarn.registerevent(bucketname);
JsonBarn.unregisterevent(bucketname);
```
-	This function work with registerevent, once you no longer want to receive event about changes inside a specific bucket you can unregister.  This function does not generate any event.

### **function setemailalert(emailaddress, bucketnames);**
```go
var JsonBarn = new JsonBarn();
JsonBarn.connect("wss://yourwebsite.com/wss/");
... once connection is eastablished you can call
JsonBarn.login(username, password);
...
JsonBarn.setemailalert("marc.gauthier3@gmail.com", ["CHAT-MESSAGES", "GENERAL-INFO"]);
```
-	JsonBarn broadcast change to the database via websocket but also send email alert when data is inserted or updated.  This function allow to request that a specific email receive alert about changes in specific buckets.  If the database already have a list of buckets to monitor for this specific email address they will be replaced with this new list of buckets.  The change will occur in two phase.  The call to setemailalert will generate a confirmation email to the user.  Once the user open his email and click on the confirmation link than the changes become permanent.


### **function connect(url);**
```go
var JsonBarn = new JsonBarn();
JsonBarn.connect("wss://yourwebsite.com/wss/");
... once connection is eastablished you can call
JsonBarn.login(username, password);
...
JsonBarn.one("INCIDENTS", "starttime", 2321232, "BIGINT");
```
-	This will generate a websocket connection between the JsonBarn backend and the client library.  Once the connection is eastablished the event **onconnect** will be fired.  Your url must always start with wss JsonBarn only support secure connection and must end with /wss/ this is the path that the mux on the backend is expecting to indicate that a websocket connection is requested..


### **function one(bucketname, searchfield, value, fieldtype);**
```go
var JsonBarn = new JsonBarn();
JsonBarn.connect("wss://yourwebsite.com/wss/");
... once connection is eastablished you can call
JsonBarn.login(username, password);
...
JsonBarn.one("INCIDENTS", "starttime", 2321232, "BIGINT");
```
-	This function will search the database for the first entry that will match the request.  You need to provide the bucketname where to search, the fieldname the value you are looking for and what type of field the fieldname represent; valid options are INT, BIGINT, TEXT, DECIMAL, DOUBLE. Once data is found and return the event **onread** will be fired.  

### **function many(bucketname, searchfield, value, fieldtype);**
```go
var JsonBarn = new JsonBarn();
JsonBarn.connect("wss://yourwebsite.com/wss/");
... once connection is eastablished you can call
JsonBarn.login(username, password);
...
JsonBarn.many("INCIDENTS", "starttime", 2321232, "BIGINT");
```
-	This function will search the database for all the entries that will match the request.  You need to provide the bucketname where to search, the fieldname the value you are looking for and what type of field the fieldname represent; valid options are INT, BIGINT, TEXT, DECIMAL, DOUBLE. Once data is found and return the event **onread** will be fired.  

### **function all(bucketname);**
```go
var JsonBarn = new JsonBarn();
JsonBarn.connect("wss://yourwebsite.com/wss/");
... once connection is eastablished you can call
JsonBarn.login(username, password);
...
JsonBarn.all("INCIDENTS");
```
-	This function will return all items in the bucketname.  Note that the system configuration contain a limit on the number of items that can be returned default to 1,000,000 items. Once data is found and return the event **onread** will be fired.  

### **function range(bucketname, searchfield, minvalue, maxvalue, fieldtype)**
```go
var JsonBarn = new JsonBarn();
JsonBarn.connect("wss://yourwebsite.com/wss/");
... once connection is eastablished you can call
JsonBarn.login(username, password);
...
JsonBarn.range("INCIDENTS", "status", 0, 1, "INT");
```
-	This function will return all items in the bucketname that match a range values.  Note that the system configuration contain a limit on the number of items that can be returned default to 1,000,000 items. Once data is found and return the event **onread** will be fired.  

### **function query(bucketname, searchpattern);**
```go
var JsonBarn = new JsonBarn();
JsonBarn.connect("wss://yourwebsite.com/wss/");
... once connection is eastablished you can call
...
JsonBarn.query("BULLETINS", 
[{"property":"Title", "type":"TEXT", "st":"EQ", "values":["Daily Scan"], "logic":"AND"}, 
 {"property":"Source", "type":"TEXT", "st":"EQ", "values":["Z"], "logic":"AND"},
 {"property":"status", "type":"INT", "st":"EQ", "values":["1"], "logic":""}
] );
```
-	This function will allow to make more advance query it allow to make a search using multiples conditions against a bucket.  

	You must provide the bucketname against which the query will be run and 
    an array that contain the conditions that need to be evaluated.
	[property, type, st, values, logic]    

	property is the property of any object in the database you want to check

	type represent the type of data for this property valid type are TEXT, INT, BIGINT, DOUBLE, DECIMAL 

	st (searchtype) represent the type of comparaison you are doing
	valid st are "EQ" equal, "GT" greater then, "GTE" greater than or equal, "LT" less then,  "LTE" less than or equal, "BETWEEN" range of values 
    
	logic represent the logic operator to add following the query valid logic are {"AND", "OR", ""}
    the last item must have a logic equal to "" or the query will not execute.
    
    For example the example above would translate to 
    	Search items with Title = "Daily Scan" AND Source = "Z" AND status = 1


### **function indexcreate(indexname, field)**
```go
var JsonBarn = new JsonBarn();
JsonBarn.connect("wss://yourwebsite.com/wss/");
... once connection is eastablished you can call
JsonBarn.login(username, password);
...
JsonBarn.indexname("idx_status", "status");
```
-	This function will create an index in the postgre database using the property provided.  If an index already exist a second one will be created.  Use indexlist to view the list of indexes already created in the database.  You need to be logged-on with admin privilege to be able to create indexes.


### **function indexlist()**
```go
var JsonBarn = new JsonBarn();
JsonBarn.connect("wss://yourwebsite.com/wss/");
... once connection is eastablished you can call
JsonBarn.login(username, password);
...
JsonBarn.indexlist();
```
-	This function will return the list of all indexes that exists in the database. You need to be logged-on with admin privilege to be able to create indexes.  The event onindexes will be fired once the list has been received.


### **function indexdrop(indexname)**
```go
var JsonBarn = new JsonBarn();
JsonBarn.connect("wss://yourwebsite.com/wss/");
... once connection is eastablished you can call
JsonBarn.login(username, password);
...
JsonBarn.indexdrop("idx_status");
```
-	This function will remove an index from the database.  Use indexlist to view the list of indexes already created in the database.  You need to be logged-on with admin privilege to be able to drop indexes.


### **function insert(bucketname, object, defered)**
```go
var JsonBarn = new JsonBarn();
JsonBarn.connect("wss://yourwebsite.com/wss/");
... once connection is eastablished you can call
JsonBarn.login(username, password);
...
JsonBarn.insert("BULLETINS", {"message":"Hello don't forget...", "time":1232141}, 4323243);
```
-	This function is use to insert database into the database.  You must provide the bucketname where to put the information an a json object.  The last parameter defered is optional, defered is use if you want to insert the item automatically at a later time.  It must be an EPOCH value number of seconds after 1970 Jan 01 in UTC time zone. If defered is small than current time or not present the data is inserted right away.

	if user does not have statuschange rights then property status and itemstatus will be set to 30 (draft).  This is usefull if you want user to create draft item that need to be confirmed or approved.


	- Special case if the bucket is "USERS" you will be entering a new user in the database.  Users are not saved in the SQL but in a local file supported by STORM database.

	- In any case you must have write privilege to the bucketname to be able to insert data.

	- If you do not set the parameter id a UUIDv4 will be generated and the id property will be assigned automatically.  It is suggested that you let the system generate the id value for each new item.

	- Once the data is inserted in the database an event oninsert will be generate.  This is the confirmation that the data has been saved and broadcasted to all users listening on that bucket.

Saving users, unlike other object USER must follow a specific structure, any properties that are not part of the structure will be ignored.  See special bucket for more details.



### **function update(bucketname, object, defered)**
```go
var JsonBarn = new JsonBarn();
JsonBarn.connect("wss://yourwebsite.com/wss/");
... once connection is eastablished you can call
JsonBarn.login(username, password);
...
JsonBarn.update("BULLETINS", {"id": "84555e5f-4272-44d2-ac2f-92635876d16f", message":"Hello !!! don't forget...", "time":44444});
```
-	This function is use to update data into the database.  You must provide an object that contain at a minimum the property id and any other properties that you want to save in the object.  The last parameter defered is optional, defered is use if you want to insert the item automatically at a later time.  It must be an EPOCH value number of seconds after 1970 Jan 01 in UTC time zone. If defered is smaller than current time or not present the data is inserted right away.

	- Special case if the bucket is "USERS" you will be entering a new user in the database.  Users are not saved in the SQL but in a local file supported by STORM database.

	- In any case you must have write privilege to the bucketname to be able to update data.

	- If you do not set the parameter id no update will be performed.

	- Once the data is update in the database an event **onupdate** will be generated.  This is the confirmation that the data has been saved and broadcasted to all users listening on that bucket.

	if user does not have statuschange right associate with his account then user can't change the status or itemstatus properties.  User can change all other properties.  If user try to update status without th right the backend server will respond with an error message.

Saving users, unlike other object USER must follow a specific structure, any properties that are not part of the structure will be ignored. See special bucket for more details.



### **function delete(bucketname, object, defered)**
```go
var JsonBarn = new JsonBarn();
JsonBarn.connect("wss://yourwebsite.com/wss/");
... once connection is eastablished you can call
JsonBarn.login(username, password);
...
JsonBarn.delete("BULLETINS", {"id": "84555e5f-4272-44d2-ac2f-92635876d16f"});
```
-	This function is use to delete item from the database.  You must provide an object that contain a property id.  The last parameter defered is optional, defered is use if you want to insert the item automatically at a later time.  It must be an EPOCH value number of seconds after 1970 Jan 01 in UTC time zone. If defered is smaller than current time or not present the data is deleted right away.

	- Special case if the bucket is "USERS" you will be entering a new user in the database.  Users are not saved in the SQL but in a local file supported by STORM database.

	- In any case you must have write privilege to the bucketname to be able to delete data.

	- If you do not set the parameter id no action will be performed.

	- Once the data is deleted in the database an event **ondelete** will be generated.  This is the confirmation that the data has been removed and broadcasted to all users listening on that bucket.


### **Event onupdate(object)**
```go
var JsonBarn = new JsonBarn();
JsonBarn.connect("wss://yourwebsite.com/wss/");
...
JsonBarn.onupdate = function (object) {
	alert(JSON.stringify(object));
}

```
-	This event is called every time an object is updated in a bucket that you have requested to received event from using **registerevent** function.

Object return will have the following properties:
- **bucketname**			name of the modified bucket
- **createdby**			name of the user that first created this item 
- **updatedby**			name of the user that modified this item 
- **createdtime **		time in seconds EPOCH when the item was created 
- **updatedtime** 		time in seconds EPOCH when the item was modified 
- **createdonnetwork** 	networkid where the item was created 
- **createdonserver**	 	serverid wehere the item was created 
- **action**		 	UPDATE
- **data**			 	the actual object 

    
 
### **Event oninsert(object)**
```go
var JsonBarn = new JsonBarn();
JsonBarn.connect("wss://yourwebsite.com/wss/");
...
JsonBarn.oninsert = function (object) {
	alert(JSON.stringify(object));
}

```
-	This event is called every time an object is inserted in a bucket that you have requested to received event from using **registerevent** function.

Object return will have the following properties:
- **bucketname**			name of the modified bucket
- **createdby**			name of the user that first created this item 
- **updatedby**			name of the user that modified this item 
- **createdtime **		time in seconds EPOCH when the item was created 
- **updatedtime** 		time in seconds EPOCH when the item was modified 
- **createdonnetwork** 	networkid where the item was created 
- **createdonserver**	 	serverid wehere the item was created 
- **action**		 	INSERT
- **data**			 	the actual object 


### **Event ondelete(object)**
```go
var JsonBarn = new JsonBarn();
JsonBarn.connect("wss://yourwebsite.com/wss/");
...
JsonBarn.ondelete = function (object) {
	alert(JSON.stringify(object));
}

```
-	This event is called every time an object is deleted from a bucket that you have requested to received event from using **registerevent** function.

Object return will have the following properties:
- **bucketname**			name of the modified bucket
- **createdby**			name of the user that first created this item 
- **updatedby**			name of the user that modified this item 
- **createdtime **		time in seconds EPOCH when the item was created 
- **updatedtime** 		time in seconds EPOCH when the item was modified 
- **createdonnetwork** 	networkid where the item was created 
- **createdonserver**	 	serverid wehere the item was created 
- **action**		 	DELETE
- **data**			 	the actual object 

    
    


### **Event onconnect()**
```go
var JsonBarn = new JsonBarn();
JsonBarn.connect("wss://yourwebsite.com/wss/");
...
JsonBarn.onconnect = function () {
	alert("A websocket connection is now eastablished with JsonBarn backend server.");
}

```
-	This event is called when a websocket is created between the browser and the JsonBarn backend server. No data is return by this event.



### **Event ondisconnect()**
```go
var JsonBarn = new JsonBarn();
JsonBarn.connect("wss://yourwebsite.com/wss/");
...
JsonBarn.ondisconnect = function () {
	alert("Your websocket connection has been terminated!");
}

```
-	This event is called when a websocket has disconnected between the browser and the JsonBarn backend server. No data is return by this event.

### **Event onlogin(username, result)**
```go
var JsonBarn = new JsonBarn();
JsonBarn.connect("wss://yourwebsite.com/wss/");
...
JsonBarn.onlogin = function (username, result) {
	if (result == "success") {
	    alert("you are now logged in with " + username); 
    } else {
    	alert("loggin attemp for " + username + " has failed");
    }
}

```
-	This event is called to indicate if a loggin attemp was succesful or not.  Username is the name of the user that you use to try to login.  Login is successful if result is equal to "success" otherwise the login has failed.

### **Event onmessage(msg)**
```go
var JsonBarn = new JsonBarn();
JsonBarn.connect("wss://yourwebsite.com/wss/");
...
JsonBarn.onmessage = function (msg) {
	    alert(msg);
 }

```
-	This event is called when the backend server is return a message that need to be display to the user.  It might be an error or a success message.

### **Event ontime(seconds)**
```go
var JsonBarn = new JsonBarn();
JsonBarn.connect("wss://yourwebsite.com/wss/");
...
JsonBarn.ontime = function (seconds) {
	    alert("current time on the server is: " + moment.unix(seconds).format());
 }

```
-	This event is generated when the backend server return the current time on the server in EPOCH format UTC timezone.


### **Event onread(bucketname, items)**
```go
var JsonBarn = new JsonBarn();
JsonBarn.connect("wss://yourwebsite.com/wss/");
...
JsonBarn.onread = function (bucketname, items) {
	    items.forEach(item) {
 			alert("receive:" + item.id + " from bucket: " + bucketname);       
        }
 }

```
-	This event is generated when the backend server return the result of a query you have made.
-	Note USERS and CONFIGURATION are also return here so you should check the bucketname to know what type of information is being returned.

### **Event onerror(msg)**
```go
var JsonBarn = new JsonBarn();
JsonBarn.connect("wss://yourwebsite.com/wss/");
...
JsonBarn.onerror = function (msg) {
	    alert("The following error just happen: " + msg);
 }

```
-	This event is generated when the backend server return the result of a query you have made.
-	Note USERS and CONFIGURATION are also return here so you should check the bucketname to know what type of information is being returned.



### **Event onindexes(server, database)**
```go
var JsonBarn = new JsonBarn();
JsonBarn.connect("wss://yourwebsite.com/wss/");
...
JsonBarn.onindexes = function (indexes) {
   		indexes.forEach(item) {
 			alert("Index: " + item);
        }
}

```
-	This event is generated when the backend server return the list of indexes you have requested.  The list is an array of string.
	
### **Event onregisterevent(bucketname)**
```go
var JsonBarn = new JsonBarn();
JsonBarn.connect("wss://yourwebsite.com/wss/");
...
JsonBarn.onregisterevent = function (bucketname) {
   		alert("You are now register to get changes for " + bucketname);
}

```
-	This event is generated when the backend confirm you have register to receive changes for a specific bucket.


### **Event onunregisterevent(bucketname)**
```go
var JsonBarn = new JsonBarn();
JsonBarn.connect("wss://yourwebsite.com/wss/");
...
JsonBarn.onunregisterevent = function (bucketname) {
   		alert("You are now unregister to get changes for " + bucketname);
}

```
-	This event is generated when the backend confirm you have unregister from receiving changes for a specific bucket.


### PROPERTIES

- [connected](#propertyconnected) return true if you have a websocket connection
```go
var JsonBarn = new JsonBarn();
if (JsonBarn.connected == true) {
 do something.
}
```
- [username](#propertyusername) return the name of the username 
```go
var JsonBarn = new JsonBarn();
alert(JsonBarn.username + " is the current user");
}
```
- [logged](#propertylogged) return true if you have provided a username and password
```go
var JsonBarn = new JsonBarn();
if (JsonBarn.logged == true) {
 alert("you are currently loggedin");
}
```
- [registerevents](#propertyregisterevents) list of event you have registered.
```go
var JsonBarn = new JsonBarn();
	forEach.JsonBarn.registereventsconnected(function(item) { 
    	alert(item);
    });
}
```
- [serversocket](#propertyserversocket) websocket object if you need to access it directly.
```go
var JsonBarn = new JsonBarn();
	JsonBarn.serversocket.send("this info");
}
```

### SERVER SIDE SECURITY

JsonBarn only support secure connections any transaction started as HTTP are redirected to a HTTPS connection.  The backend does not support unsecured websocket connections.

Once the websocket connection is established client can transmit their username and password.  Since the websocket is persistent the server will remember the username and password until the websocket connection is disconnected.  There is no requirement to resend the username and password unless the connection need to be reastablished.  The Javascript JsonBarn client **does not** store the password in memory, it is not recomanded to do so since your password would not be considered secured.

Because the websocket is persistent there is no sessionID that can be stolen.


###SPECIAL BUCKETS:

- Theses buckets have structure that can't be changed.  You can still read and update information in them but you must respect the structure, any other properties you add to objects will be discarded. 


- [USERS](#specialuserbucket) 
	- Use to store the user and their access rights, the same insert and update function can be user to access user bucket.  But the data is not saved in the SQL database but in a local database supported by STORM/BoltDB.
```go
ID string  // user name
Contact string  // how to contact this user.
PasswordHash []byte  // password hash value
Rights []string  // Rights["INCIDENTS-read"]
NewPassword string  // Use only to do a password change
Groups []string  // What group the user is part of you can use this to add more information storing data in GROUPS bucket.
EmailAlert []string  // What group the user is part of
EmailAddr string  // What is the email address 
```

- [TEMPLATES](#specialtemplatesbucket) 
	- Use to store the html templates for generating alert email.  The templates need to contain a body and a subject templates.  The format is the same use by the golang template engine.
```go
bucketname string // name of bucket as per the database
bucket string // name of bucket 
status int // 0,1,2 
body string // html template for the body
subject string // html template for the subject
```



### SPECIAL PROPERTIES:

The following properties are either required or are generating action by the database.

-	property.$id
	-	Each document in the bucket must have a unique id to be able to save and retrive information. You can manualy set the id or let JsonBarn generate the id automatically the later is suggested. 		

-	property.$starttime			
	-	Each document must have a starttime this is in most case the time that the document was created.  In some case you want data to be activated by time.  According to the starttime and endtime properties the status will be changed.
	
-	property.$endtime 			
	-	Same as starttime except this is for endtime.  
	
-	property.$status  			
	-	Each document have a status 0 is pending the docuement data is not "activated" 1 = active and 2 = completed.  You do not have to change the value of this propertie it will be change automatically by a background task running in the server.  When starttime is reach status will change to 1 and when endtime is reached status will change to 2.  

-	property.$recurrence				
	-	Item can have a starttime and endtime but they can also be reccurrent if you set a recurrence object within you object when the endtime is reached a new starttime and endtime will be generated and status will change to either pending or active.

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

Example of recurrence that start now and reoccured every 5 days at the same time.

obj.recurrence = {};
b.starttime = moment().add(10, 'seconds').utc().unix(); // JsonBarn want seconds!
obj.recurrence.startdate = moment().utc().unix();	// start now
obj.recurrence.duration = 60;
obj.recurrence.recurrencepatterncode = "D";
obj.recurrence.recurevery = 5



###USERS RIGHTS
- Each users can be granted rights to read write or delete in buckets 

    - admin			// allow user to do all actions
    - download		// allow user to download the configuration and users database for backup
    - stats-read	// allow user to read statistics
    - users-delete	// allow user to delet users 
    - xxxxxx-read  	// allow to read a specific bucket
    - xxxxxx-write 	// allow to write in a specific bucket
    - xxxxxx-delete // allow to delete items from a specific bucket 


### The LOG's
-	[Users activiy](#simple-orm)

	- All users activities are logged into a specific table in the SQL they are keep for X days where x is the configurable number of days.  This can be set in the system configuration.  LOG register all actions with a copy of the data prior and after being changed.  Default is 365 days retention.

- [System errors](#simple-orm)

	-	All error that occures in the JsonBarn framework are saved into a subfolder call logs/ with rotating log files.  Logs also contain Info, Warning and Trace information.
        


## License

MIT

## Credits

- [Marc Gauthier](marc.gauthier3@gmaile.com)
