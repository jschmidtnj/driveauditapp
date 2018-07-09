# dailydriveaudit
this software looks for users who have downloaded or shared too many files on google drive, and creates a log report, informing the admin of potential problems

## Setup
start with this setup: https://developers.google.com/drive/api/v3/quickstart/go  
run `wget` followed by the download url to get the download, on linux systems (I usually use Ubuntu 18 for testing)  
create the directory `~/go/src/` and git clone the repository in there  
make sure to create a github token by navigating to account settings => developer settings => Personal Access Tokens  
`git clone https://<github username>:<github personal access token>@github.com/colpal/driveaudit.git`  
```
cd ~/go/src/driveaudit/app  
export GOPATH=~/go  
export PATH=$PATH:$GOPATH/bin  
go get ./...  
go run *.go  
go build  
```  
See [this](https://stackoverflow.com/questions/20628918/cannot-download-gopath-not-set) for problems with gopath  

## Deployment
get into the GCP cloud console  
navigate to the directory `~/gopath/src/driveaudit/app`  
run `go get ./...`  
make sure everything works on a separate machine with the "setup" steps above  
delete everything from the container registry => images => `us.gcr.io/<project name>/appengine` folder  
then use `gcloud app deploy`  
Follow [this](https://cloud.google.com/appengine/docs/flexible/go/quickstart) if you run into any problems  
change the app.yaml configuration file if needed  
for Kubernetes deployment, follow [this](https://cloud.google.com/kubernetes-engine/docs/tutorials/hello-app) or [this](https://cloud.google.com/solutions/jenkins-on-kubernetes-engine-tutorial) and [this](https://github.com/GoogleCloudPlatform/continuous-deployment-on-kubernetes) with Jenkins  
when deploying a new version, if you changed the time expiration for the drive data, make sure to stop all services first, and manually delete the table in mongodb console (`db.driveusers.drop()`) or the update will not work.  

## Accessing Data
There are two options. The first is to use the data directly from the database, using the mongodb shell (see below).<br/><br/>
The second is to use the data from the email, specifically the file names that were downloaded. To generate a csv, copy all of the names of the files. Then open Notepad or any text-editor and paste the file names in. Finally, save the file as `<name>.csv`. When finished, this file can be opened in Excel or Google Sheets as a spreadsheet, and can be formatted however is necessary.

## Using MongoDB
login to MongoDB like this from terminal (linux): `mongo "mongodb+srv://cluster0-ug6sv.gcp.mongodb.net/test" --username admin`  
`show dbs`  
`use <database name>`  
`show tables`  
`db.<table name>.find()` finds all documents in table  
`db.<table name>.drop()` drops all documents in table  
```
db.driveusers.aggregate([
  {$match:{"email":"example@example.com", "type":"download"}},
  {"$group":{"_id": null, files:{$push:"$name"}}},
  {"$project":{files:true, _id:false}}]);
```
This command finds all file names from documents that include the email example@example.com and are downloads, and outputs the result as a list.

## The code
The code is documented in-line, and the file that you should look at to gain an understanding of the program is `app/main.go`. This is the main program.  

## Miscellaneous Notes
curl -i -X POST -d '{ "email":"asdf@gmail.com", "type":"download", "time":"1:43"}' http://localhost:3000/driveusers  
