/*! Ecureuil JSON database  - v0.0.1 - 2017-01-11
* http://Jsonbarn.io
* Copyright 2017 Marc Gauthier; Licensed MIT 

*/

var Jsonbarn = (function(){     

        var Jsonbarn = function () {
            
            this.connected = false;
            this.username =  "";
            this.logged = false;
            this.registerevents =  [];
	        this.serversocket = null;
            
            /* Events */
            this.onconnect = null;
            this.onlogin = null;
            this.onlogout = null;
            this.ondisconnect = null;
            this.onread = null;
            this.onMessage = null;
            this.oninsert = null;
            this.onupdate = null;
            this.ondelete = null;
            this.onerror = null;
            this.onstats = null;
            this.ontime = null;
            this.onindexes = null;
            this.onregisterevent = null;
            this.onunregisterevent = null;
            
           };
        
        return Jsonbarn;

})();


/* Send message to websocket if buffer is not empty then wait 1/100 of seconds and try again. 
*/
Jsonbarn.prototype.queuemsg = function(msg) {
   
    var self = this;
    if (self.serversocket.bufferedAmount == 0) {        
        self.serversocket.send(msg);
        return;
    }

    setTimeout(function()
        {
            self.queuemsg(msg);
        }, 
    10);
  
}

Jsonbarn.prototype.error = function(msg) {
        var self = this;
        if (typeof self.onerror === "function") {
            self.onerror(msg);
        }
}

Jsonbarn.prototype.time =  function(){
    var self = this;
    if (self.serversocket == null || self.connected == false) {
        self.error("There is no active connection.");        
        return;     
    }
	self.queuemsg("{\"action\":\"GETTIME\"}");
};

Jsonbarn.prototype.stats = function(){
    var self = this;
    if (self.serversocket == null || self.connected == false) {
        self.error("There is no active connection.");        
        return;     
    }
    self.queuemsg("{\"action\":\"STATS\" }");
};

Jsonbarn.prototype.getusers = function(){
    var self = this;
    if (self.serversocket == null || self.connected == false) {
        self.error("There is no active connection.");        
        return;     
    }
    self.queuemsg("{\"action\":\"GETUSERS\" }");
};

Jsonbarn.prototype.registerevent = function(bucketname){
    var self = this;
    if (self.serversocket == null || self.connected == false) {
        self.error("There is no active connection.");   
        return;     
    }
    self.queuemsg("{\"action\":\"REGISTEREVENT\", \"bucketname\":\"" + bucketname + "\" }");
};

Jsonbarn.prototype.getconfig = function(){
    var self = this;
    if (self.serversocket == null || self.connected == false) {
        self.error("There is no active connection.");   
        return;     
    }
    self.queuemsg("{\"action\":\"GETCONFIG\" }");
};

Jsonbarn.prototype.putconfig = function(data){
    var self = this;
    if (self.serversocket == null || self.connected == false) {
        self.error("There is no active connection.");   
        return;     
    }
    self.queuemsg("{\"action\":\"PUTCONFIG\", \"data\":" + JSON.stringify(data) + " }");
};

Jsonbarn.prototype.getlogs = function(startime, endtime){
    var self = this;
    if (self.serversocket == null || self.connected == false) {
        self.error("There is no active connection.");   
        return;     
    }    
    self.queuemsg("{\"action\":\"LOGS\", \"key\":\"" + startime  + "\", \"maxkey\":\"" + endtime + "\" }");
};

Jsonbarn.prototype.unregisterevent = function(bucketname){
    var self = this;
    if (self.serversocket == null || self.connected == false) {
        self.error("There is no active connection.");        
        return;     
    }
	self.queuemsg("{\"action\":\"UNREGISTEREVENT\", \"bucketname\":\"" + bucketname + "\" }");
};


Jsonbarn.prototype.setemailalert = function(email, buckets){
    var self = this;
    if (self.serversocket == null || self.connected == false) {
        self.error("There is no active connection.");        
        return;     
    }
	self.queuemsg("{\"action\":\"EMAILALERT\", \"data\": {\"email\":\"" + email + "\", \"buckets\":" + JSON.stringify(buckets) + " }}");
};

Jsonbarn.prototype.login = function(username, password){
   
    var self = this;

    alert("jsonbarn login: " + username + " " + password);

    if (self.serversocket == null || self.connected == false) {
        self.error("There is no active connection.");        
        return;     
    }
	
    if (username == undefined || username == "") {
		self.error("Username provided was empty");
        return;
    }
     
    if (password == undefined || password == "") {
        self.error("Password provided was empty")
	    return
	}

    self.queuemsg('{"action": "LOGIN", "username":"' + username + '", "password":"' + password + '"}');    
};

Jsonbarn.prototype.logout = function() {
    var self = this;
    if (self.serversocket == null || self.connected == false) {
        self.error("There is no active connection.");        
        return;     
    }
	self.queuemsg('{"action": "LOGOUT"}');
};

Jsonbarn.prototype.insert = function(bucketname, object, defered){
    var self = this;
    if (self.serversocket == null || self.connected == false) {
        self.error("There is no active connection.");        
        return;     
    }

    // value if object.ID must be set in most case, unless a structure with an autoincrement value is set. 
	
	if (bucketname == null || bucketname == undefined || bucketname == "") {
	    self.error("Unable to update object no bucketname was provided.");
		return;
	}

    cmd = "{\"action\":\"INSERT\", \"bucketname\":\"" + bucketname + "\", \"data\":" + JSON.stringify(object)

	if (object.$id != "") {
            cmd += ", \"key\": \"" + object.$id + "\"";
    }

    if (defered != null && defered != undefined && $.isNumeric(defered))  {
            cmd += ", \"defered\": " + defered;            
    }

    cmd += " }";
	
    self.queuemsg(cmd);
      
};

Jsonbarn.prototype.update = function(object, defered){
    
    //alert("update:" +JSON.stringify(object));
    
    var self = this;
    if (self.serversocket == null || self.connected == false) {
        self.error("There is no active connection.");        
         return;     
   }
	
    // value if object.ID must be set in most case, unless a structure with an autoincrement value is set. 
    if (object.$id == null || object.$id == undefined || object.$id == "") {
        self.error("Unable to update object no ID property defined.");
		return;
	}

	if (object.$bucketname == null || object.$bucketname  == undefined || object.$bucketname  == "") {
        self.error("Unable to update object no bucketname was provided.");
		return;
	}

    cmd = "{\"action\":\"UPDATE\", \"bucketname\":\"" + object.$bucketname + "\", \"key\": \"" + object.$id + "\", \"data\":" + JSON.stringify(object); 
       
    if ($.isNumeric(defered))  {
        cmd += ", \"defered\": " + defered;            
    }

    cmd += " }";
	
    alert(cmd);
    self.queuemsg(cmd);
};


Jsonbarn.prototype.updateusersettings = function(object){
    
    //alert("update:" +JSON.stringify(object));
    
    var self = this;
    if (self.serversocket == null || self.connected == false) {
        self.error("There is no active connection.");        
         return;     
   }
	
    cmd = "{\"action\":\"SETUSERSETTING\", \"data\":" + JSON.stringify(object); 
       
    cmd += " }";
	
    //alert(cmd);
    self.queuemsg(cmd);
};


Jsonbarn.prototype.delete = function(bucketname, object, defered){
    
    var self = this;
    if (self.serversocket == null || self.connected == false) {
        self.error("There is no active connection.");        
        return;     
    }

    if (object.$id == null || object.$id == undefined || object.$id == "") {
        self.error("Unable to update object no ID property defined.");
		return;
	}

	if (bucketname == null || bucketname == undefined || bucketname == "") {
  	    self.error("Unable to update object no bucketname was provided.");
		return;
	}

  
 	cmd = "{\"action\":\"DELETE\", \"bucketname\":\"" + bucketname + "\", \"key\": \"" + object.$id + "\"";

    if ($.isNumeric(defered))  {
           cmd += ", \"defered\": " + defered;            
    }

    cmd += " }";

    alert(cmd);

    self.queuemsg(cmd);
};

Jsonbarn.prototype.query = function(bucketname, pattern){
    var self = this;
    if (self.serversocket == null || self.connected == false) {
        self.error("There is no active connection.");        
         return;     
    }
    self.queuemsg("{\"action\":\"QUERY\", \"bucketname\":\"" + bucketname + "\", \"data\":" + JSON.stringify(pattern) + "}");
};


Jsonbarn.prototype.one = function(bucketname, searchfield, value, fieldtype){
    var self = this;
    if (self.serversocket == null || self.connected == false) {
        self.error("There is no active connection.");        
         return;     
    }
    if (fieldtype != "BIGINT" && fieldtype != "TEXT" && fieldtype != "INT" && fieldtype != "DECIMAL" && fieldtype != "DOUBLE") {
       self.error("Invalid field type, must be either INT, BIGINT, TEXT, DECIMAL or DOUBLE");        
         return;     
    }
    self.queuemsg("{\"action\":\"READONE\", \"bucketname\":\"" + bucketname + "\", \"key\":\"" + value + "\", \"searchfield\":\"" + searchfield + "\", \"field\":\"" + fieldtype +"\" }");
};

Jsonbarn.prototype.many = function(bucketname, searchfield, value, fieldtype){
    var self = this;
    if (self.serversocket == null || self.connected == false) {
        self.error("There is no active connection.");        
         return;     
    }
    if (fieldtype != "BIGINT" && fieldtype != "TEXT" && fieldtype != "INT" && fieldtype != "DECIMAL" && fieldtype != "DOUBLE") {
       self.error("Invalid field type, must be either INT, BIGINT, TEXT, DECIMAL or DOUBLE");        
         return;     
    }
	self.queuemsg("{\"action\":\"READFIND\", \"bucketname\":\"" + bucketname + "\", \"key\":\"" + value + "\", \"searchfield\":\"" + searchfield + "\", \"field\":\"" + fieldtype +"\" }");
};


Jsonbarn.prototype.range = function(bucketname, searchfield, minvalue, maxvalue, fieldtype){
    var self = this;
    if (self.serversocket == null || self.connected == false) {
        self.error("There is no active connection.");        
         return;     
    }
    if (fieldtype != "BIGINT" && fieldtype != "TEXT" && fieldtype != "INT" && fieldtype != "DECIMAL" && fieldtype != "DOUBLE") {
       self.error("Invalid field type, must be either INT, BIGINT, TEXT, DECIMAL or DOUBLE");        
         return;     
    }
	self.queuemsg("{\"action\":\"READRANGE\", \"bucketname\":\"" + bucketname + "\", \"key\":\"" + minvalue + "\", \"searchfield\":\"" + searchfield + "\", \"maxkey\":\"" + maxvalue + "\", \"field\":\"" + fieldtype +"\"  }");
};

Jsonbarn.prototype.all = function(bucketname){
    var self = this;
    if (self.serversocket == null || self.connected == false) {
        self.error("There is no active connection.");        
         return;     
   }
    self.queuemsg("{\"action\":\"READALL\", \"bucketname\":\"" + bucketname + "\" }");
};

Jsonbarn.prototype.indexcreate = function(indexname, field){
    var self = this;
    if (self.serversocket == null || self.connected == false) {
        self.error("There is no active connection.");        
        return;     
    }
    self.queuemsg("{\"action\":\"INDEXCREATE\", \"key\":\"" + indexname + "\", \"searchfield\":\"" + field + "\" }");
};


Jsonbarn.prototype.indexdrop = function(indexname, field){
    var self = this;
    if (self.serversocket == null || self.connected == false) {
        self.error("There is no active connection.");        
        return;     
    Trace("sending login to server");

    }
    self.queuemsg("{\"action\":\"INDEXDROP\", \"key\":\"" + indexname + "\" }");
};


Jsonbarn.prototype.indexlist = function(){
    var self = this;
    if (self.serversocket == null || self.connected == false) {
        self.error("There is no active connection.");        
        return;     
    }
    self.queuemsg("{\"action\":\"INDEXLIST\" }");
};


Jsonbarn.prototype.connect = function(host) {

            var self = this;
            
            self.serversocket = new WebSocket(host);
        
	    	self.serversocket.onopen = function(evt) {

                self.connected = true;
           
                if (typeof self.onconnect === "function") {
                    self.onconnect();
                }

			    return false;
		    }

		    self.serversocket.onclose = function(evt) {
	            
                alert("disconnect:" + JSON.stringify(evt));

                self.connected = false;

			    if (typeof self.ondisconnect === "function") {
                    self.ondisconnect();
                }

            }

		    self.serversocket.onerror = function (error) {
			
                alert("ecureuil error:" + JSON.stringify(error) + " " + error);
                
    			if (typeof self.onerror === "function") {
                    self.onerror(JSON.stringify(error));
                }

				return false;
			}



			/*************************************************************************************

			Received and analyse all the reply from the Websocket server.
			i.e. : Login result, list of locations, list of incidents, etc.

			message from the server are always lowercase message sent to server are uppercase

			**************************************************************************************/

		    self.serversocket.onmessage = function(e) {

                /* check the type of message returned! */
			
 
                try {
                    e.response = JSON.parse(e.data);
                }
                    catch(err) {
                    showAlert(e.data)        
                }

			    if (e.response.action == "login") {

					if (e.response.result == "success") {

                        self.username = e.response.username;
                        self.logged = true;
                        result = true;

                    } else {
            
                        result = false;
                    }

				    if (typeof self.onlogin === "function") {
                        self.onlogin(e.response.username, result, e.response.rights, e.response.settings);
                    }

                } else if (e.response.action == "logout") {

                        self.username = "guess";
                        self.logged = false;
                        result = true;
                        
                        if (typeof self.onlogout === "function") {
                            self.onlogout();
                        }
                
                } else if (e.response.action == "message") {

                    if (typeof self.onmessage === "function") {
                        self.onmessage(e.response.message);
                    }
                
 			    } else if (e.response.action == "read") {  
                    
                 	/* this event is fired when you read data or another user change data */
                    if (typeof self.onread === "function") {
                        // always sent response even if items count is zero    
                        self.onread(e.response.bucketname, e.response.items);                        
                    }
          
    		    } else if (e.response.action == "stats") {  

                    if (typeof self.onstats === "function") {
                        self.onstats(e.response.server, e.response.database);
                    }
             					
    		    } else if (e.response.action == "UPDATE") {  

                    if (typeof self.onupdate === "function") {
                        self.onupdate(e.response);
                    }

          		} else if (e.response.action == "DELETE") {

                    if (typeof self.ondelete === "function") {
                        self.ondelete(e.response);
                    }
        
          		} else if (e.response.action == "INSERT") {

                    if (typeof self.oninsert === "function") {
                        self.oninsert(e.response);
                    }

			    } else if (e.response.action == "gettime") {

                    if (typeof self.ontime === "function") {
                    	// receive current time on server in seconds 
				        self.ontime(e.response);
                    }

    		    } else if (e.response.action == "readindexes") {

                    if (typeof self.onindexes === "function") {
                        self.onindexes(e.response.indexes);
                    }
            
            	} else if (e.response.action == "registerevent") {
				
                    self.registerevents.push(e.response.bucketname);
                    if (typeof self.onregisterevent === "function") {
                        self.onregisterevent(e.response.bucketname);
                    }


		    	} else if (e.response.action == "unregisterevent") {

                    var index = registerevents.indexOf(e.response.bucketname);
                    if (index !== -1) {
                        self.registerevents.splice(index, 1);
                    }

                    if (typeof self.onunregisterevent === "function") {
                        self.onunregisterevent(e.response.bucketname);
                    }
				

			}

			return false;
		}

};



