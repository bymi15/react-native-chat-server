package main

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

var (
	upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
)

type message struct {
	Id        string
	Text      string
	Image     string
	TimeStamp int
	By        string
}

func validateMessage(data []byte) (message, error) {
	var msg message

	if err := json.Unmarshal(data, &msg); err != nil {
		return msg, errors.Wrap(err, "Unmarshaling message")
	}

	if msg.Id == "" || msg.Text == "" {
		return msg, errors.New("Message has no Id or Text")
	}

	return msg, nil
}

// handleWebsocket connection.
func handleWebsocket(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		m := "Unable to upgrade to websockets"
		log.WithField("err", err).Println(m)
		http.Error(w, m, http.StatusBadRequest)
		return
	}

	rr.register(ws)

mainLoop:
	for {
		mt, data, err := ws.ReadMessage()
		l := log.WithFields(logrus.Fields{"mt": mt, "data": data, "err": err})
		if err != nil {
			if websocket.IsCloseError(err, websocket.CloseGoingAway) || err == io.EOF {
				l.Info("Websocket closed!")
				break
			}
			l.Error("Error reading websocket message")
		}
		switch mt {
		case websocket.TextMessage:
			msg, err := validateMessage(data)
			if err != nil {
				l.WithFields(logrus.Fields{"msg": msg, "err": err}).Error("Invalid Message")
				break
			}
			rw.publish(data)
		case -1:
			l.Error("Connection lost!")
			break mainLoop
		default:
			l.Warning("Unknown Message!")
		}
	}

	rr.deRegister(ws)

	ws.WriteMessage(websocket.CloseMessage, []byte{})
}
