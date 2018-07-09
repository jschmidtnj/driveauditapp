package main

import (
  //"fmt"
  //"reflect"
  //"io/ioutil"
  "golang.org/x/oauth2"
  "golang.org/x/oauth2/jwt"
  "golang.org/x/oauth2/google"
  "google.golang.org/api/admin/reports/v1"
  //"log"
  //"google.golang.org/api/drive/v3"
  //"encoding/json"
)

var (
  reportsapiconf *jwt.Config
  reportsapiservice *admin.Service
)

//Google documentation for getting the token from a json file
/*
//getting the token from the google oauth file
func ExampleJWTConfigFromJSON() {
// Your credentials should be obtained from the Google
// Developer Console (https://console.developers.google.com).
// Navigate to your project, then see the "Credentials" page
// under "APIs & Auth".
// To create a service account client, click "Create new Client ID",
// select "Service Account", and click "Create Client ID". A JSON
// key file will then be downloaded to your computer.
data, err := ioutil.ReadFile("/path/to/your-project-key.json")
if err != nil {
log.Fatal(err)
}
conf, err := google.JWTConfigFromJSON(data, "https://www.googleapis.com/auth/bigquery")
if err != nil {
log.Fatal(err)
}
// Initiate an http.Client. The following GET request will be
// authorized and authenticated on the behalf of
// your service account.
client := conf.Client(oauth2.NoContext)
client.Get("...")
}
*/

//current impletementation is much simpler. saves to a global variable reportsapiservice
func connectToApi() {
  reportsapiconf = &jwt.Config{
    Email: getConfig.Service_Account_Email,
    // The contents of your RSA private key or your PEM file
    // that contains a private key.
    // If you have a p12 file instead, you
    // can use `openssl` to export the private key into a pem file.
    //
    //    $ openssl pkcs12 -in key.p12 -passin pass:notasecret -out key.pem -nodes
    //
    // The field only supports PEM containers with no passphrase.
    // The openssl command will convert p12 keys to passphrase-less PEM containers.
    PrivateKey: []byte(getConfig.Service_Account_Private_Key),
    Scopes: getConfig.Service_Account_Scopes,
    TokenURL: google.JWTTokenURL,
    // If you would like to impersonate a user, you can
    // create a transport with a subject. The following GET
    // request will be made on the behalf of user@example.com.
    // Optional.
    Subject: getConfig.Admin_Email,
  }

  client := reportsapiconf.Client(oauth2.NoContext)

  srv, err := admin.New(client)
  if err != nil {
    Log.Fatalf("Unable to retrieve reports Client %v", err)
    panic(err)
  }
  reportsapiservice = srv
}
