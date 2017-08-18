package main

import (
	"github.com/kelseyhightower/envconfig"
	"log"
	"database/sql"
	"fmt"
	"os"
	"net/http"
	"io"
	_ "github.com/go-sql-driver/mysql"
	"strconv"
	"github.com/pkg/errors"
)

type DBConfig struct {
	USER         	string // -> User
	PASS      	    string
	DATANAME     	string
}
type AppConfig struct {
	ID              int
	POSITIONOFDATA  string
	NAMEDIRECTORY   string
}

var dbconf  		DBConfig
var db   			*sql.DB
var conf   		 	AppConfig
var count   		int        = 1
func Init() {
	//configure database
	err := envconfig.Process("DB",&dbconf)
	if err != nil {
		log.Fatal(err.Error())
	}
	log.Println("line 33",dbconf.PASS,dbconf.DATANAME,dbconf.USER)

	err= envconfig.Process("CONFIGURE",&conf)
	if err != nil {
		log.Fatal(err.Error())
	}
	log.Println("line 41",conf.NAMEDIRECTORY,conf.ID,conf.POSITIONOFDATA)

}

//connectdatabase
func ConnectData() *sql.DB {
	var err error
	db, err = sql.Open("mysql", dbconf.USER+":"+dbconf.PASS+"@/"+dbconf.DATANAME)//"user:password@/dbname"
	fmt.Println("Opening connection")
	if err != nil {
		panic(err.Error()) // Just for example purpose. You should use proper error handling instead of panic
	}
	fmt.Println("checked opening connnection")

	// Open doesn't open a connection. Validate DSN data:
	err = db.Ping()
	fmt.Println("Ping database")
	if err != nil {
		panic(err.Error()) // proper error handling instead of panic in your app
	}
	fmt.Println("checked ping database")
	fmt.Println("added database channel")
	return db
}

func main()  {
	Init()

	db    = ConnectData()
	defer   db.Close()


	Row, err := db.Query("SELECT UserState.FbId,UserState.Area,UserState.Province,UserState.Name,UserState.Age, UserState.Gender,Outputs.Inning,Outputs.SampleId,Outputs.UrlRecord FROM InputText, UserState,Outputs WHERE Outputs.FbId=UserState.FbId AND Outputs.SampleId=InputText.Id AND Outputs.State=TRUE AND Outputs.Id> ?", conf.ID)
	if err != nil {
		log.Fatal("err when prepare GetInfo: ", err)
	}

	path :=conf.POSITIONOFDATA
	path  =path+"/"+conf.NAMEDIRECTORY
	if err := os.Mkdir(path,os.ModePerm); err != nil {
		log.Println("directory exist")
	}

	var FbId       string
	var Area       string
	var Province   string
	var Name       string
	var Age        string
	var Gender     string
	var UrlRecord  string
	var Inning     int
	var SampleId   int
	for Row.Next() {
		log.Println("Fetching data from database ")
		Row.Scan(&FbId,&Area,&Province,&Name,&Age,&Gender,&Inning,&SampleId,&UrlRecord)
		dirname  := Area + "_" + Province + "_" + Name + "_" + Age + "_" + Gender + "_" + FbId
		if err   :=os.Mkdir( path+"/"+ dirname, os.ModePerm ); err != nil{
			log.Println(err)
		}
		filename := strconv.Itoa(Inning) + "_" + strconv.Itoa(SampleId)
		if err   := Download(UrlRecord, path+"/"+dirname+"/"+filename); err != nil {
			log.Println("downloadFile failed: ", err)
		}
	}

}


func Download(url string, filepath string) (err error) {
		log.Println("Download: ", url)
		// Get the data
		resp, err := http.Get(url)
		if err != nil {
			return errors.Wrap(err, "http.Get failed. url="+url)
		}
		defer resp.Body.Close()

		// Create the file
		log.Println("downloading file ",count,"...")
		count++
		out, err := os.Create(filepath)
		if err != nil  {
			return err
		}
		defer out.Close()

		// Writer the body to file
		_, err = io.Copy(out, resp.Body)
		if err != nil  {
			return err
		}
		return nil
}

