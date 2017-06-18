package main

import (
	"database/sql"
	"html/template"
	"io/ioutil"
	"log"
	_ "mysql"
	"net/http"
	"router"
	"strings"
	"user"
)

var (
	DB   *sql.DB
	Path string
)

func main() {
	//init()
	Path = "C:/Users/Д/Go/fls"
	var err error
	log.SetFlags(log.Lshortfile)
	DB, err = sql.Open("mysql", "root:root@tcp(127.0.0.1:3306)/chat")
	if err != nil {
		log.Fatal(err)
	}
	defer DB.Close()

	//router settings
	router := httprouter.New()
	router.GET("/", mainHandler)
	router.POST("/auth", authHandler)
	router.POST("/signup", signupHandler)
	router.POST("/signup/uniquelogin", uLoginHandler)
	router.POST("/signup/uniquenick", uNickHandler)
	router.POST("/auth/changeavatar", changeAvatarHandler)
	router.NotFound = http.HandlerFunc(fsHandler)
	//start server
	http.ListenAndServe(":80", router)
}

//Main Handler
func mainHandler(wr http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	User := user.Get(wr, r, DB)
	tPath := Path + "/templates"
	t, err := template.ParseFiles(tPath+"/index.html", tPath+"/chat.html", tPath+"/login.html")
	if err != nil {
		log.Fatal(err)
	}
	t.Execute(wr, User)
}

//Authorization Handler
func authHandler(wr http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	user.Auth(wr, r, DB)
}

//Register Handler
func signupHandler(wr http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	user.SignUp(wr, r, DB)
}

//It's Unique?
func uLoginHandler(wr http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	login := r.FormValue("login")
	is := user.UniqueLogin(login, DB)
	if is {
		wr.Write([]byte("yes"))
	} else {
		wr.Write([]byte("no"))
	}
}

func uNickHandler(wr http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	nick := r.FormValue("nick")
	is := user.UniqueLogin(nick, DB)
	if is {
		wr.Write([]byte("yes"))
	} else {
		wr.Write([]byte("no"))
	}
}

//Change Avatar Handler
func changeAvatarHandler(wr http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	del := r.FormValue("delete")
	if del != "Delete avatar" {
		//Change avatar
		err := user.ChangeAvatar(wr, r, DB, Path)
		if err != nil {
			log.Print(err)
			http.Redirect(wr, r, "/#caerr", 301)
		} else {
			http.Redirect(wr, r, "/#casuc", 301)
		}
	} else {
		//Delete avatar
		err := user.DeleteAvatar(wr, r, DB, Path)
		if err != nil {
			log.Print(err)
			http.Redirect(wr, r, "/#delava", 301)
		} else {
			http.Redirect(wr, r, "/#delava", 301)
		}
	}
}

//File System Handler
func fsHandler(wr http.ResponseWriter, r *http.Request) {
	file, err := ioutil.ReadFile(Path + "/files/" + r.URL.Path)
	if err != nil {
		notFound(wr, r)
	} else {
		var ct string
		if strings.HasSuffix(r.URL.Path, ".css") {
			ct = "text/css"
		} else if strings.HasSuffix(r.URL.Path, ".js") {
			ct = "application/javascript"
		} else if strings.HasSuffix(r.URL.Path, ".png") {
			ct = "image/png"
		} else if strings.HasSuffix(r.URL.Path, "jpg") || strings.HasSuffix(r.URL.Path, "jpeg") {
			ct = "image/jpeg"
		} else {
			ct = "text/plain"
		}
		wr.Header().Set("Content-Type", ct)
		wr.Write(file)
	}
}

//notfound handler
func notFound(wr http.ResponseWriter, r *http.Request) {
	wr.WriteHeader(http.StatusNotFound)
	wr.Write([]byte("I love animegirls"))
}
