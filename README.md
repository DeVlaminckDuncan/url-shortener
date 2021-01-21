# Go URL Shortener Backend
This is a very fast backend application that shortens long URLs. It's written in Golang and uses Gin for routing and XORM to communicate with the database.

If you would like to test this project out using a UI, download this basic frontend application https://github.com/DeVlaminckDuncan/url-shortener-frontend.

## What you'll need to install to run this application:
- MySQL https://dev.mysql.com/downloads/mysql/ or any other database that XORM supports https://xorm.io/
- Golang https://golang.org/
- Install the Golang packages by running `go get`
- Create a file called `.env` containing:
  - `MYSQL_CONNECTION_STRING:'mysqlUsername:password@tcp(localhost:3306)/URLShortenerDB'`
  - `SECRET_JWT_KEY='yourSecretJWTKey'` if you don't know how to generate one, you can use this website https://www.grc.com/passwords.htm
  - *optional* - `ENABLE_LOGGER='false'`
  - *optional* - `ENABLE_SEED_DATABASE='false'`

## How to run or build the application:
- To run the application do `go run main.go`
- To build the application do `go build main.go`
