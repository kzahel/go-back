package main

import (
	"code.google.com/p/go-uuid/uuid"
	"code.google.com/p/go.net/websocket"
	"database/sql"
	"encoding/json"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"io"
	"log"
	"net/http"
	"os"
)

// Echo the data received on the Web Socket.
func EchoServer(ws *websocket.Conn) {
	log.Printf("echo server init")
	io.Copy(ws, ws)
}

func decodeMessage(buf []byte) interface{} {
	var m interface{}
	json.Unmarshal(buf, &m)

	msg := m.(map[string]interface{})
	for k, v := range msg {
		switch vv := v.(type) {
		case string:
			log.Println("got json key string", k, vv)

		default:
			log.Println("not handling this json thing")
		}

	}

	log.Println("received ws message %q", buf, msg)

	return m
}

type GameState struct {
	MasterPlayer string
	SlavePlayer  string
	Started      bool
	MasterStream *websocket.Conn
	SlaveStream  *websocket.Conn
}

var AllGames = make(map[string]GameState)
var AllWebSockets = make(map[string]bool)

func GameServer(ws *websocket.Conn) {
	wsptrstr := fmt.Sprintf("%p", ws)

	AllWebSockets[wsptrstr] = true
	authenticated := false
	var authed_fbid string

	log.Printf("have websocket conn", ws)

	for {
		var buf []byte
		err := websocket.Message.Receive(ws, &buf)
		if err != nil {
			log.Println("websocket receieve error", err)
			break
		}

		type GameCommand struct {
			Command string
			Data    string
			Invite  string
		}
		var request GameCommand
		json.Unmarshal(buf, &request)

		log.Println("websocket message", buf)

		if request.Command != "Authenticate" && !authenticated {
			websocket.Message.Send(ws, "authenticate please")
			break
		}

		if request.Command == "NewGame" {

			if request.Invite == "" {
				websocket.Message.Send(ws, "need invitee")
				break
			}

			var s = GameState{}
			var roomid = uuid.New()
			AllGames[roomid] = s
			type Response struct {
				Roomid string
			}
			r := Response{roomid}

			AllGames[roomid] = s
			s.MasterPlayer = authed_fbid
			s.MasterStream = ws
			s.Started = false

			b, _ := json.Marshal(r)
			websocket.Message.Send(ws, b)
		} else if request.Command == "JoinGame" {
			val, ok := AllGames[request.Data]
			log.Println("join game, ok", ok, val)
		} else if request.Command == "Authenticate" {
			websocket.Message.Send(ws, "thanks!")
			authed_fbid = request.Data
			authenticated = true
		} else {
			websocket.Message.Send(ws, "what?")
		}

	}

	delete(AllWebSockets, wsptrstr)

}

type DebugHandler struct{}

func (m DebugHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	//for k,v := range AllGames {

	//w.Write("thanks!")

	fmt.Fprintf(w, "allwebsockets:!!\n")

	b2, err := json.Marshal(AllWebSockets)
	if err != nil {
		fmt.Fprintf(w, err.Error())
	} else {
		w.Write(b2)
	}

	fmt.Fprintf(w, "\nallgames!!\n")

	b, _ := json.Marshal(AllGames)
	w.Write(b)

	//}
	//fmt.Fprintf(w, "Hi there, I love %s!", r.URL.Path[1:])
}

func servewebsocket() {
	cwd, _ := os.Getwd()
	//http.Handle("/static", http.FileServer(http.Dir(cwd))) // not sure why, but this does not work!
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir(cwd))))
	var mh = DebugHandler{}
	http.Handle("/debug/", mh)
	http.Handle("/echo", websocket.Handler(GameServer))
	err := http.ListenAndServe(":12345", nil)
	log.Printf("listening!")
	if err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}

func main() {
	db, err := sql.Open("sqlite3", "backgammon.sqlite")
	if err != nil {
		log.Fatal("error opening sqlite %v", err)
	}

	//log.Printf("opened sqlite db %v", db);

	rows, err := db.Query("select * from user")
	log.Printf("did query")
	if err != nil {
		log.Fatal(err)
		return
	}
	defer rows.Close()
	log.Printf("la la la la")
	for rows.Next() {
		var id int
		var fbid string
		var userdata string
		rows.Scan(&id, &fbid, &userdata)
		log.Println("got stuff %v %v", id, fbid, userdata)
	}
	log.Printf("all rows done!")
	rows.Close()

	servewebsocket()
}
