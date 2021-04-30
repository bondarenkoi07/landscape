package routines

import (
	"app/landscape/Model"
	"fmt"
	"github.com/gorilla/websocket"
	"io/ioutil"
	"log"
	"net/http"
)

type broadcastTransaction struct {
	value Model.ChangeData
	conn  *websocket.Conn
}

// подключенные клиенты
var clients = make(map[*websocket.Conn]bool)

// канал передачи данных между горутинами
var broadcast = make(chan broadcastTransaction, 1)

var canvas Model.Transaction

// настройка Upgrader
var upgrader = websocket.Upgrader{
	ReadBufferSize:  8192,
	WriteBufferSize: 8192,
}

// HandleConnections функция принимает и обрабатывает входящий запрос
func HandleConnections(w http.ResponseWriter, r *http.Request) {
	// Upgrade initial GET request to a websocket
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal(err)
	}
	//отправляем клиенту массив с операциями, сделанныи до его подключения

	err = ws.WriteJSON(canvas)
	if err != nil {
		log.Fatal(err)
	}

	defer ws.Close()
	//добавляем нового клиента
	clients[ws] = true

	for {
		var msg Model.ChangeData
		// считываем данные, полученные поо вебсокету
		err := ws.ReadJSON(&msg)
		validation := true
		if err != nil {
			log.Printf("error: %v", err)
			delete(clients, ws)
			break
		}
		files, err := ioutil.ReadDir("static/assets/models")
		if err != nil {
			log.Fatalf("error: %v", err)
		}
		validator := Model.NewModels(files)

		if !validator.Find(msg.Model) && (msg.Mode == "DeleteModel" || msg.Mode == "PlaceModel") {
			validation = false
		}
		//тут будем передавать сообщения другим  горутинам
		if validation {
			var transaction broadcastTransaction
			transaction.value = msg
			transaction.conn = ws
			broadcast <- transaction
			//if error occurred during validation, sends repaired canvas to user
			//whom message cause error
		} else {
			msg.Mode = "DeleteModel"
			err = ws.WriteJSON(msg)
			if err != nil {
				log.Printf("error: %v", err)
				ws.Close()
				delete(clients, ws)
			}
		}
	}
}

// HandleMessages
func HandleMessages() {
	for {
		// получаем сообщение из канала
		data := <-broadcast
		// отправляем сообщение каждому клиенту
		output := make(map[int]Model.ChangeData)
		output[0] = data.value
		for client := range clients {
			if client != data.conn {
				err := client.WriteJSON(output)
				if err != nil {
					log.Printf("error: %v", err)
					client.Close()
					delete(clients, client)
				}
			}
		}
	}
}

// MatrixTestHandler here we would receive json serialization of matrix
func MatrixTestHandler(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer ws.Close()
	for {
		// считываем данные, полученные по вебсокету
		var lastCanvas Model.Transaction
		fmt.Print("received canvas to " + ws.RemoteAddr().String() + "\n")
		err := ws.ReadJSON(&lastCanvas)
		if err != nil {
			log.Print(err)
			delete(clients, ws)
			break
		}
		files, err := ioutil.ReadDir("static/assets/models")
		if err != nil {
			log.Fatal(err)
		}
		var status = make(map[string][]string)
		var validation = true
		validator := Model.NewModels(files)
		for i := range lastCanvas.Object {
			for j := range lastCanvas.Object[i] {
				item := lastCanvas.Object[i][j]
				//log.Print("name =  " + item)
				if item != "" {
					//log.Print(item)
					if !validator.Find(item) {
						log.Print("wow" + item)
						validation = false
						item = ""
						status["error"] = append(status["error"], "element "+item+" is restricted")
					}
				}
			}
		}
		if validation {
			status["status"] = append(status["error"], "true")
		} else {
			status["status"] = append(status["error"], "false")
			err = ws.WriteJSON(canvas)
			if err != nil {
				log.Print(err)
				delete(clients, ws)
				break
			}
		}
		canvas = lastCanvas
		fmt.Print("send canvas to " + ws.RemoteAddr().String() + "\n")
	}
}

//TODO: create routine to validate data
//TODO: implement ErrorHandler for errors, raising during the game (however, now errors come only from @MatrixTestHandler via var @status)
