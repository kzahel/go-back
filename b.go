package main

import(
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"log"
"code.google.com/p/go.net/websocket"
"io"
"net/http"
"encoding/json"
"code.google.com/p/go-uuid/uuid"
)


// Echo the data received on the Web Socket.
func EchoServer(ws *websocket.Conn) {
	log.Printf("echo server init")
	io.Copy(ws, ws);
}

func decodeMessage(buf []byte) interface{} {
	var m interface{}
	json.Unmarshal(buf, &m)

	msg := m.(map[string]interface{})
	for k,v := range msg {
		switch vv := v.(type) {
		case string:
			log.Println("got json key string",k,vv);

		default:
			log.Println("not handling this json thing");
		}
		
	}

	log.Println("received ws message %q", buf, msg);

	return m;
}


type GameState struct {
	MasterPlayer string
	SlavePlayer string
	Started bool
	MasterStream *websocket.Conn
	SlaveStream *websocket.Conn
}




var AllGames = make(map[string]GameState)

func GameServer(ws *websocket.Conn) {
	authenticated := false;

	log.Printf("have websocket conn", ws);

	for {
		var buf []byte
		err := websocket.Message.Receive(ws, &buf)
		if (err != nil) {
			log.Println("websocket receieve error", err);
			break;
		}

		type GameCommand struct {
			Command string
			Data string
			Invite string
		}
		var request GameCommand
		json.Unmarshal(buf, &request);

		log.Println("websocket message", buf)

		if (request.Command != "Authenticate" && ! authenticated) {
			websocket.Message.Send(ws, "authenticate please")
			break;
		}

		if (request.Command == "NewGame") {

			if (request.Invite == "") {
				websocket.Message.Send(ws, "need invitee")
				break;
			}

			var s = GameState{}
			var roomid = uuid.New()
			AllGames[roomid] = s
			type Response struct {
				Roomid string
			}
			r := Response{roomid}

			AllGames[roomid] = s
			s.MasterStream = ws;
			s.Started = false;

			b, _ := json.Marshal(r)
			websocket.Message.Send(ws, b)
		} else if (request.Command == "JoinGame") {

			val, ok := AllGames[request.Data]

			log.Println("join game, ok", ok,val)
		} else if (request.Command == "Authenticate") {
			websocket.Message.Send(ws, "thanks!")
			authenticated = true;
		} else {
			websocket.Message.Send(ws, "what?")
		}
		

	}

}

func servewebsocket() {
	log.Printf("serve websocket")
	http.Handle("/echo", websocket.Handler(GameServer));
	err := http.ListenAndServe(":12345", nil);
	if err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}


func main() {
	db, err := sql.Open("sqlite3", "/home/ubuntu/backgammon.sqlite")
	if (err != nil) {
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