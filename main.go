package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"sort"
	"time"

	"github.com/gorilla/websocket"
	"github.com/johnny-debt/instascrap"
	"github.com/sociafill/gorillas"
	"github.com/sociafill/sp2pt"
)

var hub gorillas.Gorillas
var broker sp2pt.Broker
var consumer = hashtagConsumer{}
var maxIDs = make(map[string]string)

func processCommand(conn *websocket.Conn) bool {
	// Read raw bytes from the connection
	messageType, payload, err := conn.ReadMessage()
	if err != nil {
		log.Printf("Payload read error: %v\n", err)
		return false
	}
	log.Printf("Payload read [%d]: %s\n", messageType, payload)
	// Parse raw bytes to the internal command struct
	var command ClientCommand
	err = json.Unmarshal(payload, &command)
	if err != nil {
		log.Printf("Payload parsing error: %v\n", err)
		return true
	}
	log.Printf("Command parsed: %v\n", command)
	if command.Command == "watch" {
		hashtag := watchedHashtag{slug: command.Hashtag}
		hub.Subscribe(conn, gorillas.Topic(hashtag.Identifier()))
		broker.Watch(&hashtag)
	}
	return true
}

// Configure the upgrader
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type message interface{}

// ClientCommand defines our message object
type ClientCommand struct {
	Command string `json:"command"` // "watch" or "unwatch"
	Hashtag string `json:"hashtag"`
}

func main() {
	hub = gorillas.NewGorillas()
	broker = sp2pt.NewBroker(consumer)
	// Configure WebSocket route
	http.HandleFunc("/ws", handleConnections)
	port := os.Getenv("SOCIAFILL_WATCHER_PORT")
	if port == "" {
		port = "8000"
	}
	// Start the server on localhost port 8000 and log any errors
	log.Println("http server started on :" + port)
	err := http.ListenAndServe(":"+port, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

func handleConnections(w http.ResponseWriter, r *http.Request) {
	// Upgrade initial GET request to a WebSocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal(err)
	}
	// Make sure we close the connection when the function returns
	defer hub.RemoveConnection(conn)

	// Register our new client
	hub.AddConnection(conn)

	for {
		if !processCommand(conn) {
			break
		}
	}
}

type watchedHashtag struct {
	slug  string
	maxID string
}

func (hashtag watchedHashtag) Identifier() sp2pt.Identifier {
	return sp2pt.Identifier(hashtag.slug)
}

func (hashtag watchedHashtag) Poll() []interface{} {
	medias, _ := instascrap.GetHashtagMedia(hashtag.slug)
	// Sort medias be ascending ID
	sort.SliceStable(medias, func(i, j int) bool {
		return medias[i].ID < medias[j].ID
	})

	var items []interface{}
	for _, v := range medias {
		maxID, exists := maxIDs[hashtag.slug]
		if !exists || v.ID > maxID {
			maxIDs[hashtag.slug] = v.ID
			items = append(items, v)
		}
	}
	return items
}

func (hashtag watchedHashtag) GetInterval() time.Duration {
	return time.Second * 2
}

type hashtagConsumer struct {
}

func (consumer hashtagConsumer) Consume(object sp2pt.Observable, item interface{}) {
	switch item.(type) {
	case instascrap.Media:
		log.Printf("Media #%s received for source %s\n", item.(instascrap.Media).ID, object.Identifier())
		hub.SendJSON(gorillas.Topic(object.Identifier()), item)
	default:
		log.Printf("Unknown object received (%T)\n", item)
	}
}
