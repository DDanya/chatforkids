package user

import (
	"database/sql"
	"errors"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"log"
	"mysql"
	"net/http"
	"os"
	"strconv"
)

type User struct {
	Id       int
	Login    string
	password string
	Nickname string
	Avatar   int
}

func (u User) Token() string {
	return "testoken"
}

func Get(wr http.ResponseWriter, r *http.Request, db *sql.DB) User {
	var u User
	id, err1 := r.Cookie("id")
	token, err2 := r.Cookie("token")
	if err1 != nil || err2 != nil {
		if err1 != http.ErrNoCookie || err2 != http.ErrNoCookie {
			log.Print(err1)
			log.Fatal(err2)
			return User{}
		} else {
			return User{}
		}
	}
	i, err := strconv.Atoi(id.Value)
	if err != nil {
		log.Print(err)
		return User{}
	}
	err = db.QueryRow("SELECT login, pass, nick, avatar FROM users WHERE id = ?", i).Scan(&u.Login, &u.password, &u.Nickname, &u.Avatar)
	if err != nil {
		log.Print(err)
		return User{}
	}
	u.Id = i
	if token.Value == u.Token() {
		return u
	}
	return User{}
}

func ChangeAvatar(wr http.ResponseWriter, r *http.Request, db *sql.DB, path string) error {
	//Check token
	token := r.FormValue("token")
	u := Get(wr, r, db)
	if token != u.Token() {
		return errors.New("token1: " + token + " != token2: " + u.Token())
	} else {
		//Parse Form
		err := r.ParseMultipartForm(5242880)
		if err != nil {
			return err
		}
		file, handler, err := r.FormFile("avatar")
		if err != nil {
			return err
		}
		defer file.Close()
		//Logging
		log.Print("new file: " + handler.Filename)
		log.Print(handler.Header)

		//Working with file
		var (
			conv bool
			dec  image.Image
		)
		switch handler.Header.Get("Content-Type") {
		case "image/jpeg":
			dec, err = jpeg.Decode(file)
			if err != nil {
				return err
			}
			conv = true
		case "image/png":
			conv = false
		}

		//Save image and change avatar
		f, err := os.OpenFile(path+"/files/a/"+u.Login+".png", os.O_WRONLY|os.O_CREATE, 0666)
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()
		if conv {
			err = png.Encode(f, dec)
			if err != nil {
				return err
			}
		} else {
			io.Copy(f, file)
		}
		row, err := db.Prepare("UPDATE users SET avatar = 1 WHERE id = ?")
		if err != nil {
			log.Fatal(err)
		}
		res, err := row.Exec(u.Id)
		if err != nil {
			log.Fatal(err)
		}
		log.Print(res.LastInsertId())
		return nil
	}
}

func DeleteAvatar(wr http.ResponseWriter, r *http.Request, db *sql.DB, path string) error {
	token := r.FormValue("token")
	u := Get(wr, r, db)
	if token != u.Token() {
		return errors.New("token1: " + token + " != token2: " + u.Token())
	} else {
		err := os.Remove(path + "/files/a/" + u.Login + ".png")
		if err != nil {
			return err
		}
		row, err := db.Prepare("UPDATE users SET avatar = 0 WHERE id = ?")
		if err != nil {
			log.Fatal(err)
		}
		res, err := row.Exec(u.Id)
		if err != nil {
			log.Fatal(err)
		}
		log.Print(res.LastInsertId())
	}
	return nil
}

func Auth(rw http.ResponseWriter, req *http.Request, db *sql.DB) {
	var r User
	login := req.FormValue("login")
	r.password = req.FormValue("password")
	err := db.QueryRow("SELECT id FROM users WHERE login = ? AND pass = ?", login, r.password).Scan(&r.Id)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Redirect(rw, req, "/#logerr", 301)
		} else {
			log.Fatal(err)
		}
	}
	if r.Id != 0 {
		http.SetCookie(rw, &http.Cookie{
			Name:  "id",
			Value: strconv.Itoa(r.Id),
		})
		http.SetCookie(rw, &http.Cookie{
			Name:  "token",
			Value: r.Token(),
		})
	}
	c, err := req.Cookie("id")
	log.Print(c)
	http.Redirect(rw, req, "/#logsuc", 301)
}

func SignUp(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	var u User
	u.Login = r.FormValue("login")
	u.password = r.FormValue("password")
	u.Nickname = r.FormValue("nickname")
	row, err := db.Prepare("INSERT INTO users(login, pass, nick) VALUES (?,?,?)")
	if err != nil {
		log.Fatal(err)
	}
	res, err := row.Exec(u.Login, u.password, u.Nickname)
	if driverErr, ok := err.(*mysql.MySQLError); ok {
		if driverErr.Number == 1062 {
			w.WriteHeader(http.StatusForbidden)
			w.Write([]byte("FUCK YOU DUDE"))
		} else {
			log.Fatal(err)
		}
	} else {
		a, err := res.LastInsertId()
		if err != nil {
			log.Print(err)
		}
		c := strconv.FormatInt(a, 10)
		http.SetCookie(w, &http.Cookie{
			Name:  "id",
			Value: c,
		})
		http.SetCookie(w, &http.Cookie{
			Name:  "token",
			Value: u.Token(),
		})
		http.Redirect(w, r, "/#regsuc", 301)
	}
}

func UniqueLogin(login string, db *sql.DB) bool {
	var i int
	err := db.QueryRow("SELECT id FROM users WHERE login = ?", login).Scan(&i)
	if err != nil {
		if err == sql.ErrNoRows {
			return true
		} else {
			log.Fatal(err)
		}
	}
	return false
}

func UniqueNick(nick string, db *sql.DB) bool {
	var i int
	err := db.QueryRow("SELECT id FROM users WHERE nick = ?", nick).Scan(&i)
	if err != nil {
		if err == sql.ErrNoRows {
			return true
		} else {
			log.Fatal(err)
		}
	}
	return false
}
