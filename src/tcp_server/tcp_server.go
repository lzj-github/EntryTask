package main

import (
	_ "github.com/go-sql-driver/mysql"
	"database/sql"
	"os"
	"log"
	"os/exec"
	"hash/fnv"
	"strconv"
	"strings"
	"net"
	"net/rpc"
	"util"
    )

var conf = make(map[string]interface{})

type mysqlCli struct{
	db *sql.DB
}

var mCli *mysqlCli

func init(){
	dir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	conf = util.ConfReader(dir + "/../../conf/setting.conf")
	logDir := conf["log_file_dir"].(string)
	f, err := os.OpenFile( dir + "/" + logDir + "/tcp_server.log", os.O_RDWR | os.O_CREATE | os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer f.Close()
	log.SetOutput(f)
	mCli = &mysqlCli{db:nil}
}

func (my *mysqlCli ) Connect() {
	if  my.db == nil{
		var err error
		dbDriver := "mysql"
		dbUser := conf["db_user"].(string)
		dbPass := conf["db_pass"].(string)
		dbName := conf["db_name"].(string)
		dbAddr := conf["mysql_host"].(string) + ":" + conf["mysql_port"].(string) 
		my.db, err = sql.Open(dbDriver, dbUser+":"+dbPass+"@tcp(" + dbAddr +")/"+dbName)
		if err != nil {
			log.Println(err.Error())
		}
	}
}

func (my *mysqlCli) Close(){
	if my.db != nil {
		my.db.Close()
	}
}

func (my *mysqlCli) Inquery(sql string, paras ... string ) bool{
	my.Connect()
	stmt, err := my.db.Prepare(sql)
	if len(paras) == 1 {
		_, err =stmt.Exec(paras[0])
	}else if len(paras) == 2 {
		_, err =stmt.Exec(paras[0], paras[1])
	}else if len(paras) == 3 {
		_, err =stmt.Exec(paras[0],paras[1] , paras[2])
	}else if len(paras) == 4 {
		_, err =stmt.Exec(paras[0],paras[1] , paras[2], paras[3])
	}

	if err == nil{
		return true
	}else{
		return false
	}
}

func uuID() string {
    out, err := exec.Command("uuidgen").Output()
    if err != nil {
        log.Fatal(err)
    }
    return strings.Replace(string(out), "\n", "", -1)
}

func hash(s string) string {
	h := fnv.New64a()
	h.Write([]byte(s))
	return strconv.FormatUint(h.Sum64(), 10)
}

func insertUser( realname string, nickname string, pwd string, avatar string) string {
	//redis format :  username:realname
	resp := util.RedisGet( "user:" + realname)
	if resp != "" {
		log.Println(realname +" already exists!")
		return "{\"code\":1,\"msg\":\"should NOT overwrite existing data\",\"uuid\":\"\"}"
	}

	uuid := uuID()
	hashedPwd :=string(hash(pwd))
	mCli.Inquery("INSERT user SET uuid=?,realname=?,nickname=?,pwd=?",uuid, realname,nickname,hashedPwd)
	if mCli.db == nil{
		log.Println("mysql client is nil")
	}

	util.RedisPut("user:"+realname, uuid + "_"+ hashedPwd + "_" + nickname)
	util.RedisPut("uuid:"+uuid, uuid + "_"+ hashedPwd + "_" + nickname+ "_" + realname)
	return login(realname ,pwd)
}

func login(realname string, pwd string) string {
	hashedPwd :=string(hash(pwd))
	resp := util.RedisGet("user:"+realname)
	if resp == "" {
		return "{\"code\":1,\"msg\":\"fail\",\"uuid\":\"\"}"
	}
	uuid_pwd_nickname := strings.Split(resp,"_")
	log.Println("check upn: " + resp)
	log.Println("check uuid: " + uuid_pwd_nickname[0] )
	log.Println("check pwd: " + uuid_pwd_nickname[1] )
	log.Println("check nn: " + uuid_pwd_nickname[2] )

	if hashedPwd != uuid_pwd_nickname[1]{
		return "{\"code\":1,\"msg\":\"failed\",\"uuid\":\"\"}"
	}
	return "{\"code\":0,\"msg\":\"success\",\"uuid\":\"" + uuid_pwd_nickname[0] + "\"}"
}

func lookup(uuid string) string {
	// lookup the redis cache first
	photoID := util.RedisGet("uuid_pid:"+uuid)
	if photoID ==""{
		return "{\"code\":2,\"msg\":\"failed\",\"nickname\":\"\",\"photoid\":\"" + photoID + "\"}"
	}
	resp := util.RedisGet("uuid:"+uuid)
	if resp == "" {
		return "{\"code\":3,\"msg\":\"failed\",\"nickname\":\"\",\"photoid\":\"" + photoID + "\"}"
	}
	id_pwd_pid_nn_rn := strings.Split(resp,"_")
	return "{\"code\":0,\"msg\":\"success\",\"nickname\":\"" + id_pwd_pid_nn_rn[2] +"\",\"photoid\":\"" + photoID + "\"}"
}

func lookupAvatar(uuid string) string {
    resp := util.RedisGet("uuid_pid:" +uuid)
	log.Println("lookup avatar : " + resp)
	//return "{code:0,msg :'success',data:'{uuid:" + uuid + "}'}"
	return "{\"code\":0,\"msg\":\"success\",\"photoid\":\"" + resp + "\"}"
}

func updateNickname( uuid string, nickname string) string {
	mCli.Inquery("update user set nickname=? where uuid=?",nickname, uuid)
	uuid_pid_nn_rn := util.RedisGet("uuid:"+uuid)
	log.Println(uuid_pid_nn_rn)
	upnr := strings.Split(uuid_pid_nn_rn,"_")
	util.RedisPut("uuid:"+uuid, uuid + "_" + upnr[1] + "_" + nickname+ "_" + upnr[3])
	log.Println(upnr[0])
	log.Println(upnr[1])
	log.Println(upnr[2])
	log.Println(upnr[3])

	uuid_pwd_nn := util.RedisGet("uuid:" + upnr[3])
	upn := strings.Split(uuid_pwd_nn,"_")
	log.Println("upn: " + uuid_pwd_nn)
	_=upn
	return "{\"code\":0,\"msg\":\"\"}";
}

func insertAvatar( uuid string, pid string) string {
	sql := "insert into  avatar (uuid,pid)  values (?,?)"
	affect := mCli.Inquery(sql, uuid, pid)
	if affect  {
		util.RedisPut("uuid_pid:"+uuid,pid)
		return "{\"code\":0,\"msg\":\"success\",\"data\":\"\"}";
	} else {
		return "{\"code\":2,\"msg\":\"failed to insert avatar\"}";
	}
}

func updateAvatar( uuid string, pid string) string {
	affect := mCli.Inquery("update avatar set pid=? where uuid=?",pid, uuid)
	if affect {
		//update redis cache	
		return "{\"code\":0,\"msg\":\"success\"}";
	} else {
		return "{\"code\":1,\"msg\":\"failed to update avatar\"}";
	}
}

type Query string

func (t *Query) SignUp( args *util.Args4, reply *string) error{
	*reply = insertUser(args.A, args.B, args.C, args.D)
	return nil
}

func (t *Query) SignIn( args *util.Args2, reply *string) error{
	*reply = login(args.A, args.B)
	return nil
}

func (t *Query) Lookup( args *util.Args2, reply *string) error{
	*reply = lookup(args.A)
	return nil
}

func (t *Query) LookupAvatar( args *util.Args2, reply *string) error{
	*reply = lookupAvatar(args.A)
	return nil
}

func (t *Query) InitAvatar( args *util.Args2, reply *string) error{
	*reply = insertAvatar(args.A, args.B)
	return nil
}

func (t *Query) ChangeAvatar( args *util.Args2, reply *string) error{
	*reply = updateAvatar(args.A, args.B)
	return nil
}

func (t *Query) ChangeNickname( args *util.Args2, reply *string) error{
	*reply = updateNickname(args.A, args.B)
	return nil
}

func main() {
	/*
	dir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	conf = util.ConfReader(dir + "/../../conf/setting.conf")
	logDir := conf["log_file_dir"].(string)
	f, err := os.OpenFile( dir + "/" + logDir + "/tcp_server.log", os.O_RDWR | os.O_CREATE | os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer f.Close()
	log.SetOutput(f)
*/
    teller := new(Query)
    rpc.Register(teller)

	tcp_host := conf["tcp_server_host"].(string)
	tcp_port := conf["tcp_server_port"].(string)
	tcp_addr := tcp_host + ":" + tcp_port
    tcpAddr, err := net.ResolveTCPAddr("tcp", tcp_addr)
    listener, err := net.ListenTCP("tcp", tcpAddr)
	_ = err
    for {
        conn, err := listener.Accept()
        if err != nil {
            continue
        }
        rpc.ServeConn(conn)
	}
}
