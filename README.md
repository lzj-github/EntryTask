# Requirements of the System

## Functional Requirements
1. User signup :real name as a global unique name , for login use 
2. User login : If the login is successful, redirect to /home page , user information will be displayed, session info will be saved, otherwise an error message will be shown.
3.  Edit profile : After a successful login, a user can edit the following information: \
  (1) Upload a picture as his/her profile picture\
  (2) Change his/her nickname (support unicode characters with utf-8 encoding)
User information includes: username (cannot be changed), nickname, profile picture

4. Separate HTTP server and TCP server and put the main logic on TCP server.Backend authentication logic should be done in the TCP server

5. User information must be stored in a MySQL database. Connect by MySQL Go client.
We can use redis as a user info cache ,for performance improvement purpouse

## Non-Functional Requirements & Considerations:

#### Robustness
1. Use standard library whenever possible.
2. Horizontal scalablity when traffic surge
3. Code extensibility when needed
4. Avoid single point of failure

#### Security
1. Web interface will not directly connect to MySQL. For each HTTP request, web interface will send a TCP request to the TCP server, which will handle business logic and query the database.

2. Login frequency check and IP check (such as MaxMind GeoIP2 Database),todo
3. Login abnormal behavior warning ( such as abnormal region or device ),todo
4. Login behavior logging and analysis ( such as hacking),todo


## Performance Requirement 

1. Supports up to 1000 login requests per second (from at least 200 unique users)
2. Supports up to 1000 concurrent http requests
3. For test, the initial user data can be directly insert into database. Make sure there are at least 10 million user accounts in the test database.


###  Extendable Functionality ( todos ):
1. Roles and privileges enhancement and management, todo
2. SSO ( Single Sign On ) Central Authentication Service for cross domain login, todo
3. Sign in by OpenId such as QQ, Wechat, google or facebook account, todo
4. OpenId Authentication for other applications, todo


#### Environment Requirements

Server: Virtual Machine on Working PC\
OS: CentOS 7 x64\
DB: MySQL 5.7.23\
Client: Chrome and Firefox
  

# Design of the System

## System Archecture

#### Final Overlook (just show the idea)
After improvements , the system should look like:

![image](http://alejandroseaah.com:4869/4fe9982280f58404f88f4ab8fec783a1?h=600&w=500)

#### Real Picture
But now , it is:

![image](http://alejandroseaah.com:4869/ff75e5165d164bac7f55cb75b2aeebf9?w=400&h=500)

## Database Design
All tables are in a database named UserDB, there are three tables:
1. user table : storing uuid (universally unique identifier, as unique user id), real name , nick name and password
2. avatar table : storeing uuid and photo id
3. login talbe : storing login records, for security and user behavior study purpose
    
![image](http://alejandroseaah.com:4869/98336e55522fac37af942a20de1e5655?w=600&h=600)


Every user has a uuid, and all the tables share the same unique user id, this is the only field shared between all tables , data consistency is ensured by outside applications, not mysql itself.

With the simple connection between them, it is easy to split database when necessary.

* uuid is generated by /usr/bin/uuidgen on linux, with 36 letters.


## System APIs


### TCP Server
tcp server provide rpc service for web server rpc call:

1.func (t *Query) SignUp( args *Args4, reply *string) error

call function:
 func insertUser( realname string, nickname string, pwd string, avatar string) string


parameters:

| para  |type | required  | max len| desc | example|
| ----- |:----:|:----:|:----:|:----:|:----:|
| realname | string| yes |1024 |||
| username | string| yes |1024 |||
| pwd1 | string| yes |32 |password||
| pwd2 | string| yes |32 |confirmed password||

response:

| para  |type | required  | max len| desc | example|
| ----- |:----:|:----:|:----:|:----:|:----:|
| code | string| yes | 1 |0 for success，otherwise fail||
| msg | string| no |  |for detail info |user or password mismatch|
| uuid | string| no | | return uuid after success login, or NULL after failure| |


2.func (t *Query) SignIn( args *Args2, reply *string) error
call function:
func login(realname string, pwd string) string 

3. func (t *Query) Lookup( args *Args2, reply *string) error
call function :
func lookup(uuid string) string

4. func (t *Query) InitAvatar( args *Args2, reply *string) error
call function:
func insertAvatar( uuid string, pid string) string

5. func (t *Query) ChangeAvatar( args *Args2, reply *string) error
call function:
func updateAvatar( uuid string, pid string) string

### http web server

1. /signup
2. /login
3. /edit
4. /home
5. /upload
6. /logout

## Redis Cache
## LoadBalance
We can add a Load balancing layer at some places in our system:
  1. Between http web server and tcp servers
  2. Between tcp Servers and database servers
  3. Between http Servers and redis  cache servers
  4. Between http Servers 
  ...

  
## Telemetry
How many times a real name  has been used within an hour, what were the users? Any abnormal user behavior? Some statistics worth tracking: country of the visitor, date and time of access, web page that refers the click, browser or platform from where the page was accessed, how long they stayed, etc....


## Other considerations
replace mysql with MariaDB, Cassandra , Green plum or other RMDB ?


# repo

https://github.com/pkusnail/EntryTask.git


# Reference:
1. https://www.datastax.com/dev/blog/2012-in-review-performance
2. https://mariadb.com/sites/default/files/A_Quick_Start_Guide_to_Backup_Technologies_-_MariaDB_White_Paper_-_08_26_13_001.pdf
3. http://www.eandbsoftware.org/wp-content/uploads/2015/03/MariaDB_vs_MySQL_-_MariaDB_White_Paper_-_08_26_13_001.pdf
4. https://dev.maxmind.com/geoip/
5. http://alejandroseaah.com:4869
6. https://golang.org/doc/articles/wiki/
7. 

