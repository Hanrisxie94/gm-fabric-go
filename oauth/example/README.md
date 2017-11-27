# OAuth Integration Example Service
A basic example service that integrates GM Fabric's OAuth middleware

## Prerequisites
1.  [Dex](https://github.com/coreos/dex)
2.  [Go](https://golang.org)
3.  An HTTP REST Client

## Configuration
1.  Follow the steps to clone dex and build the binary
2.  Start the dex server:
```bash
./bin/dex serve examples/config-dex.yaml
```
3.  Start the dex example app:
```bash
./bin/example-app
```
4.  Sign in with the example credentials
```bash
username: admin@example.com
password: password
```
5.  Copy the token created by the user authorization

## Running The Example Service
1.  Fetch dependencies:
```bash
dep ensure -v
```
2.  Build the binary:
```bash
go build
```
3.  Run the binary:
```bash
./example
```
4.  Make a GET request to the `/movies` endpoint with the following headers:
```http
Authorization: Bearer $TOKEN_YOU_COPIED_FROM_DEX
```
5.  If the token was verified successfully, you will see a JSON list of movies, and a user object will be printed out in the console
