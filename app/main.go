package main

import (
	"fmt"
	//"io/ioutil"
	"flag"
	"net/http"
	"os"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"

	//"reflect"

	//"golang.org/x/oauth2/jwt"
	//"golang.org/x/net/context"
	//"golang.org/x/oauth2/google"
	//"google.golang.org/api/admin/reports/v1"
	contextScope "github.com/gorilla/context"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

const (
	VERSION = "1.0.0"
)

// imports the config.go file immediately, setting the constants to the getConfig variable, as defined in the config/config.toml file
var getConfig = Config{}

// This data structure is for user drive data to be added to the working database for api calls
type datapush struct {
	Email string
	Type  string
	Time  time.Time
	Name  string
}

//This function runs the google reports api and sends the data to the mongodb database
func runDriveCheck() {
	checkDatabasePeriodically()
	t := time.Now()
	//Log logs to a file, and mongologger logs to mongodb
	fmt.Printf("Input Connected to %v, %s\n", mainsession.LiveServers(), t.Format("2006-01-02 15:04:05"))
	Log.Printf("Input Connected to %v, %s\n", mainsession.LiveServers(), t.Format("2006-01-02 15:04:05"))
	mongoLogger.Printf("Input Connected to %v, %s\n", mainsession.LiveServers(), t.Format("2006-01-02 15:04:05"))
	//coll is the working mongodb collection of user data
	coll := mainsession.DB(getConfig.Database_Name).C(getConfig.User_Data_Collection)
	//session expire is the amount of time the data has before it expires, in minutes
	session_expire := time.Minute * time.Duration(getConfig.Expire)
	//this sets the index information for the working data
	index := mgo.Index{
		Key:         []string{"time"},
		Unique:      false,
		DropDups:    false,
		Background:  true,
		ExpireAfter: session_expire}
	//and this adds the index to the collection
	if err := coll.EnsureIndex(index); err != nil {
		mongoLogger.Printf("problem creating index", err)
		Log.Printf("problem creating index", err)
		panic(err)
	}
	//these regex commands will be used later to compare to data from the api calls
	matchregex, err1 := regexp.Compile("download|move|change_user_access")
	shareregex, err1 := regexp.Compile("shared_externally|public_on_the_web|unknown")
	whitelistregex, err1 := regexp.Compile(getConfig.Whitelisted_Emails_Regex)
	if err1 != nil {
		mongoLogger.Printf("problem with regex", err1)
		Log.Printf("problem with regex", err1)
		panic(err1)
	}
	//var checkDB = true
	//count iterations is a counter for the number of times the api call occurred
	var count_iterations = 0
	for {
		count_iterations++
		//token refresh count is the number of iterations of api calls before a new token needs to be generated, via oauth2
		if count_iterations%getConfig.Token_Refresh_Count == 0 {
			connectToApi()
		}
		//queryTimeDuration := time.Now().Add(-1 * time.Second * 5)

		//this is the main reports api call, to get the drive data
		r, err := reportsapiservice.Activities.List("all", "drive").MaxResults(getConfig.MaxResults).Do()
		if err != nil {
			mongoLogger.Printf("Unable to retrieve data from the domain. ", err)
			Log.Printf("Unable to retrieve data from the domain. ", err)
			panic(err)
		}
		var count = 0
		//iterate over the number of items in the response
		if len(r.Items) == 0 {
			fmt.Println("No Data found.")
		} else {
			//fmt.Println("Data:")
			for _, a := range r.Items {
				count++
				//log the current time to the database, per api request
				time, err := time.Parse(time.RFC3339Nano, a.Id.Time)
				if err != nil {
					mongoLogger.Printf("time parse error", err)
					Log.Printf("time parse error", err)
					panic(err)
					//fmt.Println("Unable to parse data time.")
					// Set time to zero.
					//time = time.Time{}
				}
				//fmt.Println(a.IpAddress) //ip address
				//fmt.Println(a.Events[0].Parameters[1].Value)

				var shared bool
				//check if the visibility is set to shared_externally
				if shareregex.MatchString(a.Events[0].Parameters[5].Value) {
					shared = true
				} else {
					shared = false
				}
				//testing data from api calls:
				//name := a.Events[0].Parameters[3].Value + "." + a.Events[0].Parameters[2].Value
				//email is parameter 6, only goes up to 6, primary_event is 0, doc_id is 1, doc_type is 2, visibility is 4, doc_name is 3
				//originating_app_id is 5
				//fmt.Printf(a.Events[0].Parameters[5].Name + ": " + name + ", ")
				//If regex matches add to working database
				if matchregex.MatchString(a.Events[0].Name) || shared {
					//time := t.Format(time.RFC3339Nano)
					var doc_name string
					if a.Events[0].Name == "download" {
						doc_name = a.Events[0].Parameters[3].Value + "." + a.Events[0].Parameters[2].Value
					} else {
						//fmt.Println(a.Events[0].Parameters[0].Name,a.Events[0].Parameters[1].Name,a.Events[0].Parameters[2].Name,a.Events[0].Parameters[3].Name,a.Events[0].Parameters[4].Name,a.Events[0].Parameters[5].Name,a.Events[0].Parameters[6].Name,a.Events[0].Parameters[7].Name,a.Events[0].Parameters[8].Name,a.Events[0].Parameters[9].Name,a.Events[0].Parameters[10].Name)
						//folders are listed:
						if a.Events[0].Parameters[7].Value == "folder" {
							//take care of folders listed in a for loop
							//fmt.Println(a.Events[0].Parameters[8].Value,a.Events[0].Parameters[8].Value[0])
						}
						doc_name = a.Events[0].Parameters[8].Value + "." + a.Events[0].Parameters[7].Value
					}
					email := a.Actor.Email

					//if the email is whitelisted, don't add it
					if !(whitelistregex.MatchString(email)) {
						var event string = ""
						if shared {
							event = "shared externally"
						} else {
							event = a.Events[0].Name
						}
						data := datapush{
							Email: email,
							Type:  event,
							Time:  time,
							Name:  doc_name,
						}
						//values := map[string]{"email": email, "time": time, "action": "download"}
						//if the data is not already in the database, add it there:
						count_in_db, err := coll.Find(bson.M{"time": time, "email": email, "type": event}).Count()
						if err != nil {
							mongoLogger.Printf("mongo lookup count error", err)
							Log.Printf("mongo lookup count error", err)
							panic(err)
						} else if count_in_db == 0 {
							if err := coll.Insert(data); err != nil {
								mongoLogger.Printf("mongo user insert error", err)
								Log.Printf("mongo user insert error", err)
								panic(err)
							}
						}
						//fmt.Println("Document Inserted Successfully")
					}
				}
				//fmt.Printf("%s: %s %s   %d\n", t.Format(time.RFC3339Nano), a.Actor.Email, a.Events[0].Name, count)
			}
		}
		//fmt.Println(count_iterations)
		//check the database for negligent users every given amount (periodically configured in config.go)
		if count_iterations%getConfig.Count_Between_Check == 0 {
			go checkDatabasePeriodically()
		}
		//remove all logs after a certain number of iterations (configurable)
		if count_iterations%getConfig.Count_Between_Log_Reset == 0 {
			coll := mainsession.DB(getConfig.Database_Name).C(getConfig.Log_Data_Collection)
			coll.RemoveAll(nil)
		}
		//optional functionality to check database every minute or hour
		/*
		   timeStampString := time.Now().Format("2006-01-02 15:04:05")
		   layout := "2006-01-02 15:04:05"
		   timestamp, err := time.Parse(layout, timeStampString)
		   if err != nil && checkDB {
		     Log.Printf("problem with current time lookup", err)
		     fmt.Println(err)
		   }
		   var every = 1
		   //hr, min, sec
		   _, min, _ := timestamp.Clock()
		   if min % every == 0 && checkDB {
		     checkDatabasePeriodically()
		     checkDB = false
		   } else if min % every != 0 {
		     checkDB = true
		   }
		*/
		//delay the api calls by a certain amount of time to prevent too many api calls per day
		time.Sleep(time.Duration(getConfig.Seconds_Between_Api) * time.Second)
	}
}

//structure for a negligent user in the database
type negligentUser struct {
	Email    string   `json:"email" bson:"email"`
	Type     string   `json:"type" bson:"type"`
	Count    int      `json:"count" bson:"count"`
	Data     []bson.M `json:"data" bson:"data"`
	FileData string   `json:"filedata" bson:"filedata"`
}

//checks the working database for negligent users, and adds it to another collection in the database
func checkDatabasePeriodically() {
	//time.Sleep(5000 * time.Millisecond)
	// connect to the database
	t := time.Now()
	//logging
	fmt.Printf("Periodic Check Connected to %v, %s\n", mainsession.LiveServers(), t.Format("2006-01-02 15:04:05"))
	Log.Printf("Periodic Check Connected to %v, %s\n", mainsession.LiveServers(), t.Format("2006-01-02 15:04:05"))
	mongoLogger.Printf("Periodic Check Connected to %v, %s\n", mainsession.LiveServers(), t.Format("2006-01-02 15:04:05"))

	//get the two collections - one for negligent users and one for the working drive data
	coll := mainsession.DB(getConfig.Database_Name).C(getConfig.User_Data_Collection)
	badcoll := mainsession.DB(getConfig.Database_Name).C(getConfig.Suspect_Data_Collection)

	//this pipeline finds data by email and type, and counts the data. Then it sorts and matches based on whether it meets the
	//preconfigured threshold, and if so, returns the data as a new object
	getUsersPipe := []bson.M{
		{"$group": bson.M{"_id": bson.M{"email": "$email", "type": "$type"}, "count": bson.M{"$sum": 1}}},
		{"$sort": bson.M{"count": -1}},
		{"$match": bson.M{"count": bson.M{"$gte": getConfig.Threshold}}},
		{"$group": bson.M{"_id": "$_id.email", "type": bson.M{"$first": "$_id.type"}, "count": bson.M{"$first": "$count"}}},
	}
	//notes on other structures tried
	//{"$sort": bson.M{"count":-1}},
	//{$project:{email:"$email", count:{$cond: {if: {$gte: ["$count", 50]}, then: "$count", else: 0}}}},
	//{$redact:{$cond:{if:{$eq:["$count", 0]}, then: "$$PRUNE", else:"$$DESCEND"}}}
	//{$group:{_id:"$_id.email", $push:{count:{$each: "$count_in_db", $position: 0}}}}
	//fmt.Println(time.Now().UTC())

	//start the pipeline with mgo
	getUsersPipeline := coll.Pipe(getUsersPipe)
	resp := []bson.M{}
	geterr := getUsersPipeline.All(&resp)
	if geterr != nil {
		mongoLogger.Printf("problem with pipeline", geterr)
		Log.Printf("problem with pipeline", geterr)
		panic(geterr)
	}
	//iterate over responses and send emails for the negligent users
	for i := 0; i < len(resp); i++ {
		email := resp[i]["_id"].(string)
		action := resp[i]["type"].(string)
		count := resp[i]["count"].(int)
		getAllDataPipe := []bson.M{
			{"$match": bson.M{"email": email, "type": action}},
			{"$group": bson.M{"_id": "$time", "type": bson.M{"$first": "$type"}}},
		}
		//all of the requests for the particular negligent individual - downloads or shared_externally, is stored in the negligent user collection database
		getAllDataPipeline := coll.Pipe(getAllDataPipe)
		alldata := []bson.M{}
		newerr := getAllDataPipeline.All(&alldata)
		if newerr != nil {
			mongoLogger.Printf("problem with pipeline2", newerr)
			Log.Printf("problem with pipeline2", newerr)
			panic(newerr)
		}
		fmt.Println("get file names " + email + action)
		//db.driveusers.aggregate([{$match:{"email":"joshua_schmidt@gapps-dev.colpal.com", "type":"download"}},{"$group":{"_id": null, files:{$push:"$name"}}},{"$project":{files:true, _id:false}}])
		getDataPipe := []bson.M{
			{"$match": bson.M{"email": email, "type": action}},
			{"$group": bson.M{"_id": "null", "names": bson.M{"$push": "$name"}}},
			//{"$project": bson.M{"names": true, "_id": false}},
		}
		getDataPipeline := coll.Pipe(getDataPipe)
		res1 := []bson.M{}
		err1 := getDataPipeline.All(&res1)
		if err1 != nil {
			mongoLogger.Printf("problem with pipeline", geterr)
			Log.Printf("problem with pipeline", geterr)
			panic(geterr)
		}
		files := res1[0]["names"].([]interface{})
		arrayfiles := make([]string, len(files))
		for i, v := range files {
			arrayfiles[i] = fmt.Sprint(v)
		}
		//this is just comma-seperated string values, not a csv file or csv data per say
		var csvdata string
		for _, element := range arrayfiles {
			csvdata = csvdata + "," + element
		}
		var filedata = csvdata
		user := negligentUser{
			Email:    email,
			Type:     action,
			Count:    count,
			Data:     alldata,
			FileData: filedata,
		}
		//fmt.Println(user.Files)
		badUserQuery := bson.M{
			"email": email,
			"type":  action,
		}
		//this section ensures that there is only one of each use in the negligent user collection
		count_in_db, err := badcoll.Find(badUserQuery).Count()
		if err != nil {
			mongoLogger.Printf("problem with find in mongo", err)
			Log.Printf("problem with find in mongo", err)
			panic(err)
		} else if count_in_db > 1 {
			_, err := badcoll.RemoveAll(bson.M{"email": email, "type": action})
			if err != nil {
				mongoLogger.Printf("Error deleting multiple of the same bad users objects", err)
				Log.Printf("Error deleting multiple of the same bad users objects", err)
				panic(err)
			} else {
				t := time.Now()
				fmt.Printf("Deleted multiple bad user objects %s\n", t.Format("2006-01-02 15:04:05"))
				Log.Printf("Deleted multiple bad user objects %s\n", t.Format("2006-01-02 15:04:05"))
				mongoLogger.Println("Deleted multiple bad user objects %s\n", t.Format("2006-01-02 15:04:05"))
				count_in_db = 0
			}
		} else {
			if count_in_db == 0 {
				//fmt.Println(user.FileData)
				if err := badcoll.Insert(user); err != nil {
					mongoLogger.Printf("problem with mongo insert", err)
					Log.Printf("problem with mongo insert", err)
					panic(err)
				} else {
					t := time.Now()
					fmt.Printf("Inserted: %s, %s, %s, %s\n", email, action, count, t.Format("2006-01-02 15:04:05"))
					Log.Println("Inserted: %s, %s, %s, %s\n", email, action, count, t.Format("2006-01-02 15:04:05"))
					mongoLogger.Println("Inserted: %s, %s, %s, %s\n", email, action, count, t.Format("2006-01-02 15:04:05"))
					//SEND EMAIL HERE!!!
					//this section gets all the file names, and saves them to be sent in the email
					var actionstr string = ""
					countstr := strconv.Itoa(count)
					subject := getConfig.Subject
					destination := getConfig.Recipient
					r := NewRequest([]string{destination}, subject)
					if strings.Compare(action, "download") == 0 {
						actionstr = "downloaded"
					} else {
						actionstr = action
					}
					//fmt.Println(csvdata)
					//the email is then sent here via the mailjet api, in the mailer.go file
					fmt.Println("send email")
					r.Send("templates/template.html", map[string]string{"mongo": getConfig.Mongo_Connect, "docnames": csvdata, "username": getConfig.Recipient_Name, "email": email, "count": countstr, "type": actionstr})
				}
			} else {
				//update saved users if they are in the database already
				//optional, working, but as entries expire from the database it is updated with old files.
				//maybe make it so that it counts the number of files in the update and if greater, updates, otherwise does not
				if err := badcoll.Update(badUserQuery, user); err != nil {
					mongoLogger.Printf("problem with mongo update", err)
					Log.Printf("problem with mongo update", err)
					panic(err)
				} else {
					t := time.Now()
					fmt.Printf("Updated: %s, %s, %d, %s\n", email, action, count, t.Format("2006-01-02 15:04:05"))
					Log.Printf("Updated: %s, %s, %d, %s\n", email, action, count, t.Format("2006-01-02 15:04:05"))
					mongoLogger.Printf("Updated: %s, %s, %d, %s\n", email, action, count, t.Format("2006-01-02 15:04:05"))
				}
				//this code was to check if it was necessary to update the negligent user, but it was more trouble then it was worth
				//it does not work perfectly yet
				/*
				   //This sees if it needs to update
				   fmt.Println("checking to update")
				   //if the data is the same don't update it
				   past_user_data := bson.M{}
				   current_user_data := alldata
				   err := badcoll.Find(badUserQuery).One(&past_user_data)//["data"].(string)
				   if err != nil {
				     fmt.Println(err)
				     Log.Printf("error with past data read for bad user", err)
				     panic(err)
				   }
				   //fmt.Println(past_user_data)
				   //DEEP EQUAL IS NOT WORKING.
				   fmt.Println(current_user_data, "\n\n\n\n", past_user_data)
				   eq := reflect.DeepEqual(current_user_data, past_user_data)
				   if eq {
				     fmt.Println("Did not update:", email, action, count)
				     Log.Println("Did not update:", email, action, count)
				   } else {
				     if err := badcoll.Update(badUserQuery, user); err != nil {
				       Log.Printf("problem with mongo update", err)
				       panic(err)
				     } else {
				       fmt.Println("Updated:", email, action, count)
				       Log.Println("Updated:", email, action, count)
				     }
				   }
				*/
			}
		}
		//fmt.Println("Document Inserted Successfully")
		//time.Sleep(10 * 1000 * time.Millisecond)
	}
	//log that the check was finished
	fmt.Println("Finished periodic check")
	Log.Println("Finished periodic check")
	mongoLogger.Println("Finished periodic check")
}

//this connects to the database and starts the web server for REST requests
func runServer() {
	// connect to the database
	db := mainsession
	t := time.Now()
	fmt.Printf("Web Connected to %v, %s\n", db.LiveServers(), t.Format("2006-01-02 15:04:05"))
	Log.Printf("Web Connected to %v, %s\n", db.LiveServers(), t.Format("2006-01-02 15:04:05"))
	mongoLogger.Printf("Web Connected to %v, %s\n", db.LiveServers(), t.Format("2006-01-02 15:04:05"))
	// Adapt our handle function using withDB
	h := Adapt(http.HandlerFunc(handle), withDB(db))
	// add the handler
	http.Handle(getConfig.REST_Handler, contextScope.ClearHandler(h))
	// start the server
	//changed from 8080 originally
	if err := http.ListenAndServe(getConfig.REST_Port, nil); err != nil {
		mongoLogger.Printf("error starting server", err)
		Log.Printf("error starting server", err)
		panic(err)
	}
}

//the first function that runs, and initializes the program
func init() {
	//first get the configuration
	getConfig.Read()
	fmt.Println("read config")
	connectToApi()
	fmt.Println("connected to api")
	mongoconnect()
	fmt.Println("connected to mongodb")
	createMainLogger(mainsession)
	flag.Parse()
	//filelocation := "logs/log_" + time.Now().Format(time.RFC3339)
	var logpath = flag.String("logpath", "logs/main.log", "Log Path")
	NewLog(*logpath)
	//Log.SetOutput(f)
	Log.Println("\n-----------------------------------------------------------------------------------------------------------------------")
	Log.Printf("Server v%s pid=%d started with processes: %d\n", VERSION, os.Getpid(), runtime.GOMAXPROCS(runtime.NumCPU()))
	fmt.Println("Created Log File")
	mongoLogger.Println("\n-----------------------------------------------------------------------------------------------------------------------")
	mongoLogger.Printf("Server v%s pid=%d started with processes: %d\n", VERSION, os.Getpid(), runtime.GOMAXPROCS(runtime.NumCPU()))
}

//the second function that runs, the main function, necessary in golang programs
func main() {
	go runDriveCheck()
	go frontEndApp()
	runServer()
}
