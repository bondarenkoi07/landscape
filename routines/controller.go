package routines

import (
	"app/landscape/Model"
	"github.com/gorilla/websocket"
	"log"
	"math"
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

func init() {

	canvas.Landscape = make(map[int64][]int64)
	canvas.Object = make(map[int64][]string)

	intOfCols := int64(math.Floor(
		math.Sqrt(float64(
			Model.CanvasWidth*Model.CanvasWidth+Model.CanvasHeight*Model.CanvasHeight),
		) / Model.GridCount,
	),
	)
	intOfRows := int64(math.Floor(Model.CanvasHeight/Model.GridCount) + 2)

	log.Printf("cols: %d, rows: %d", intOfCols, intOfRows)

	for y := -intOfRows; y < intOfRows; y++ {
		canvas.Landscape[y] = make([]int64, 0)
		canvas.Object[y] = make([]string, 0)

		var x int64

		for x = 0; x < intOfCols; x++ {
			canvas.Landscape[y] = append(canvas.Landscape[y], 0)
			canvas.Object[y] = append(canvas.Object[y], "")
		}
	}
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
		err = ws.ReadJSON(&msg)
		validation := true
		if err != nil {
			log.Printf("error: %v", err)
			delete(clients, ws)
			break
		}
		validation, err = canvas.Check(msg) //TODO: implements error sending

		if err != nil {
			if err.Error() == "unknown model" {
				delete(clients, ws)
				break
			}
		}

		//тут будем передавать сообщения другим  горутинам
		if validation {
			var transaction broadcastTransaction
			transaction.value = msg
			transaction.conn = ws
			broadcast <- transaction
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

// MatrixTestHandler here we would receive json serialization of matrix and test it
func MatrixTestHandler(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer ws.Close()
	for {
		// считываем данные, полученные по вебсокету
		var lastCanvas Model.Transaction
		err = ws.ReadJSON(&lastCanvas)
		if err != nil {
			log.Print(err)
			delete(clients, ws)
			break
		}
		err = lastCanvas.CheckMap()
		if err != nil {
			err = ws.WriteJSON(canvas)
			if err != nil {
				log.Print(err)
				delete(clients, ws)
				break
			}
		} else {
			canvas = lastCanvas
		}
	}
}
