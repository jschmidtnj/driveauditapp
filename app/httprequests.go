package main

import (
  "encoding/json"
  "net/http"
  "bytes"
  "fmt"

  "gopkg.in/mgo.v2/bson"
  contextScope "github.com/gorilla/context"
  "gopkg.in/mgo.v2"
)

//REST http handler for requests, and finds data from both collections in database

type Adapter func(http.Handler) http.Handler

func Adapt(h http.Handler, adapters ...Adapter) http.Handler {
  for _, adapter := range adapters {
    h = adapter(h)
  }
  return h
}

func handle(w http.ResponseWriter, r *http.Request) {
  switch r.Method {
  case "GET":
    getUsers(w, r)
  case "POST":
    postUsers(w, r)
  case "DELETE":
    deleteUsers(w, r)
  case "GETRAW":
    handleRead(w, r)
  case "POSTRAW":
    handleInsert(w, r)
  case "DELETERAW":
    handleDelete(w, r)
  default:
    http.Error(w, "Not supported", http.StatusMethodNotAllowed)
  }
}

type baduser struct {
  ID     bson.ObjectId `json:"id" bson:"_id"`
  Email string        `json:"email" bson:"email"`
  Type   string        `json:"type" bson:"type"`
  Count   int     `json:"count" bson:"count"`
}

func deleteUsers(w http.ResponseWriter, r *http.Request) {
  db := contextScope.Get(r, "database").(*mgo.Session)

  // load the users
  buf := new(bytes.Buffer)
  buf.ReadFrom(r.Body)
  q := buf.String()
  var query map[string]interface{}
  json.Unmarshal([]byte(q), &query)
  var driveusers []*baduser
  if err := db.DB(getConfig.Database_Name).C(getConfig.Suspect_Data_Collection).
  Find(query).Sort("-Time").Limit(1000).All(&driveusers); err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
    return
  }
  if err := db.DB(getConfig.Database_Name).C(getConfig.Suspect_Data_Collection).Remove(query); err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
    return
  }

  // write it out
  if err := json.NewEncoder(w).Encode(driveusers); err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
    return
  }
}

func postUsers(w http.ResponseWriter, r *http.Request) {
  db := contextScope.Get(r, "database").(*mgo.Session)

  // decode the request body
  var c baduser
  if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
    http.Error(w, err.Error(), http.StatusBadRequest)
    return
  }

  // give the user a unique ID
  c.ID = bson.NewObjectId()

  // insert it into the database
  if err := db.DB(getConfig.Database_Name).C(getConfig.Suspect_Data_Collection).Insert(&c); err != nil {
    http.Error(w, err.Error(), http.StatusBadRequest)
    return
  }

  // redirect to it
  http.Redirect(w, r, "/driveusers/"+c.ID.Hex(), http.StatusTemporaryRedirect)
}

func getUsers(w http.ResponseWriter, r *http.Request) {
  db := contextScope.Get(r, "database").(*mgo.Session)
  buf := new(bytes.Buffer)
  buf.ReadFrom(r.Body)
  q := buf.String()
  var query map[string]interface{}
  json.Unmarshal([]byte(q), &query)
  // load the users
  var driveusers []*baduser
  if err := db.DB(getConfig.Database_Name).C(getConfig.Suspect_Data_Collection).
  Find(query).Sort("-Time").Limit(1000).All(&driveusers); err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
    return
  }

  // write it out
  if err := json.NewEncoder(w).Encode(driveusers); err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
    return
  }
}

type driveuser struct {
  ID     bson.ObjectId `json:"id" bson:"_id"`
  Email string        `json:"email" bson:"email"`
  Type   string        `json:"type" bson:"type"`
  Time   string     `json:"time" bson:"time"`
}

func handleDelete(w http.ResponseWriter, r *http.Request) {
  db := contextScope.Get(r, "database").(*mgo.Session)

  // load the users
  buf := new(bytes.Buffer)
  buf.ReadFrom(r.Body)
  q := buf.String()
  var query map[string]interface{}
  json.Unmarshal([]byte(q), &query)
  var driveusers []*driveuser
  if err := db.DB(getConfig.Database_Name).C(getConfig.User_Data_Collection).
  Find(query).Sort("-Time").Limit(1000).All(&driveusers); err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
    return
  }
  if err := db.DB(getConfig.Database_Name).C(getConfig.User_Data_Collection).Remove(query); err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
    return
  }

  // write it out
  if err := json.NewEncoder(w).Encode(driveusers); err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
    return
  }
}

func handleInsert(w http.ResponseWriter, r *http.Request) {
  db := contextScope.Get(r, "database").(*mgo.Session)

  // decode the request body
  var c driveuser
  if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
    http.Error(w, err.Error(), http.StatusBadRequest)
    return
  }

  // give the user a unique ID
  c.ID = bson.NewObjectId()

  // insert it into the database
  if err := db.DB(getConfig.Database_Name).C(getConfig.User_Data_Collection).Insert(&c); err != nil {
    http.Error(w, err.Error(), http.StatusBadRequest)
    return
  }

  // redirect to it
  http.Redirect(w, r, "/driveusers/"+c.ID.Hex(), http.StatusTemporaryRedirect)
}
func handleRead(w http.ResponseWriter, r *http.Request) {
  db := contextScope.Get(r, "database").(*mgo.Session)
  buf := new(bytes.Buffer)
  buf.ReadFrom(r.Body)
  q := buf.String()
  var query map[string]interface{}
  json.Unmarshal([]byte(q), &query)
  // load the users
  var driveusers []*driveuser
  if err := db.DB(getConfig.Database_Name).C(getConfig.User_Data_Collection).
  Find(query).Sort("-Time").Limit(1000).All(&driveusers); err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
    return
  }

  // write it out
  if err := json.NewEncoder(w).Encode(driveusers); err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
    return
  }
}

func withDB(db *mgo.Session) Adapter {
  fmt.Println("Running Http requests")
  // return the Adapter
  return func(h http.Handler) http.Handler {

    // the adapter (when called) should return a new handler
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

      // copy the database session
      dbsession := db.Copy()
      defer dbsession.Close() // clean up

      // save it in the mux context
      contextScope.Set(r, "database", dbsession)

      // pass execution to the original handler
      h.ServeHTTP(w, r)

    })
  }
}
