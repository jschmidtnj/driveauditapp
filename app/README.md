# driveauditappcode

## The main code for the drive audit application
* Can be exported in Docker by changing the file `Rename Dockerfile` to `Dockerfile`, and configuring it to your specifications.  
* Dockerfile.scratch could also be used, but it has been having mixed results.  
* Logs can be found in the logs folder, if running on a VM. app.yaml is used to configure app engine.  
* k8s is used to configure Kubernetes with Jenkins, along with Jenkinsfile, tests, and jenkins.   
* docs just contains pictures that are used by Jenkins.  
* config is used for the configuration of the program, and templates is used as a template for the email file.  
* html.go is used with frontEndApp.go to create a front end user experience, which is just for testing purposes right now, and requires port 8080.  
* httprequests.go is used for get, post, and delete requests for the database, but requires port 3000 to be open to use.  
