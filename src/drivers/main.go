package main

import (
	"log"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"github.com/streadway/amqp"
)

const (
	version 		 = "0.1.0"
	author		   = "Gribouille"
	serverURL    = "127.0.0.1:3000"
	defaultURI   = "amqp://guest:guest@localhost:5672/"
	exchangeName = "drivers"
	routingKey   = "drivers.update"
	helloMessage = `Hello drivers

Version: %s
Author : %s
`
)

var broker = NewBroker()

// Hello handler.
func helloHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	w.Write([]byte(fmt.Sprintf(helloMessage, version, author)))
}

// Drivers handler.
func driversHandler(messages <-chan amqp.Delivery, done chan error) {
	for d := range messages {
		//log.Printf("got %dB delivery: [%v] %q", len(d.Body), d.DeliveryTag, d.Body)
		broker.Notifier <- []byte(d.Body)
	}
	log.Printf("driversHandler: messages channel closed")
	done <- nil
}

// exists returns whether the given file or directory exists or not
func exists(path string) (bool, error) {
    _, err := os.Stat(path)
    if err == nil { return true, nil }
    if os.IsNotExist(err) { return false, nil }
    return true, err
}

func getPublicDir() (string, error) {
	exe, err := os.Executable()
	if err != nil {
		return "", err
	}
	dir := filepath.Dir(exe)
	pub := filepath.Join(dir, "public")
	ok, err := exists(pub)
	if !ok || err != nil {
		return "", fmt.Errorf("Cannot find public directory: %s", pub)
	}
	return pub, nil
}


func main() {
	dir, err := getPublicDir()
	if err != nil {
		log.Fatal(err)
	}
	rootHandler := http.FileServer(http.Dir(dir))
	http.Handle("/", rootHandler)
	http.Handle("/hello", http.HandlerFunc(helloHandler))
	http.Handle("/events/", broker)

	go func() {
		err := StartConsumer(defaultURI, exchangeName, routingKey, driversHandler)
		if err != nil {
			log.Fatal(err)
		}
	}()

	log.Printf("Serve on %s\n", serverURL)
	log.Fatal("HTTP server error: ", http.ListenAndServe(serverURL, nil))
}
