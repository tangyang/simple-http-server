# a simple http server for tantan app

This is a simple restful http server for simplified tantan backend, which provide the features below:

* add a new user 
* get all users
* establish a new relationship with another person
* get all existed relationship for a specified user


# Table of contents

* [Installation](#installation)
* [Get started](#getstarted)
* [Api documents](#documents)

## Installation

Install:

```
go get github.com/tangyang/simple-http-server
```
We assume that you have a PostgreSQL database server already, otherwise you can install a PostgreSQL server locally following the  steps below.

```
mv Dockerfile.postgresql Dockerfile
docker build --rm=true -t mypostgresql:9.4 .
docker run -i -t  -d -p 5432:5432 mypostgresql:9.4

``` 

## Getstarted

```
simple-http-server -init //create database schema when you start the server the first time

simple-http-server // start the server
```

We also provide a few configuration parameters which are supposed to be in a config.toml file in the same directory of ***simple-http-server***. Here is an example of these configuration parameters: 

```
http-port="8001"   //http server port
pg-address = "192.168.56.101:5432"  //PostgreSQL database address
pg-username = "pger"      //PostgreSQL database user name
pg-password = "pger"      //PostgreSQL database password
pg-db-name = "pgerdb"     //database name
pg-poolsize = 50          //database connection pool size
pg-readtimeout = 3        //read timeout in seconds for PostgreSQL
pg-writetimeout = 4       //write timeout in seconds for PostgreSQL
pg-idletimeout = 5        //the amount of time in seconds after which client closes idle db connections

```
## documents

### add a new user 

```
curl -XPOST -d '{"name":"Alice1"}' "http://localhost:8000/users"
 
{"Code":200,"Message":"","Data":{"Id":2,"Name":"Alice1","Type":"user"}}

```

### get all users 

```
curl -XGET "http://localhost:8000/users"

{"Code":200,"Message":"","Data":[{"Id":1,"Name":"Alice","Type":"user"},{"Id":2,"Name":"Alice1","Type":"user"}]}

```

### establish a new relationship

```
curl -XPUT -d '{"state":"liked"}' "http://localhost:8000/users/12/relationships/10"

{"Code":200,"Message":"","Data":{"UserId":10,"State":"matched","Type":"relationship"}}
```

### get all relationships of a user
```
curl -XGET "http://localhost:8000/users/10/relationships"

{"Code":200,"Message":"","Data":[{"UserId":11,"State":"liked","Type":"relationship"},{"UserId":13,"State":"disliked","Type":"relationship"},{"UserId":12,"State":"matched","Type":"relationship"}]}

```