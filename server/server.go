package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/julienschmidt/httprouter"
)

type controller struct {
	lock          *sync.Mutex
	notifications []Notification
}

func (c *controller) receive(w http.ResponseWriter, r *http.Request, _ httprouter.Params) error {
	defer r.Body.Close()
	var notification Notification
	traceFile, err := os.Create(fmt.Sprintf("/traces/%d", time.Now().Unix()))
	if err != nil {
		return err
	}
	defer traceFile.Close()
	decoder := json.NewDecoder(io.TeeReader(r.Body, traceFile))
	if err := decoder.Decode(&notification); err != nil {
		return err
	}
	c.lock.Lock()
	defer c.lock.Unlock()
	c.notifications = append(c.notifications, notification)
	return nil
}

func (c *controller) report(w http.ResponseWriter, r *http.Request, _ httprouter.Params) error {
	w.Header().Add("Content-Type", "application/json")
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	return encoder.Encode(c.notifications)
}

type fallibleHandler func(http.ResponseWriter, *http.Request, httprouter.Params) error

func handleFallible(inner fallibleHandler) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
		log.Printf("%s %s %s", r.Method, r.URL.String(), r.RemoteAddr)
		if err := inner(w, r, params); err != nil {
			log.Printf("error: %v", err)
		}
	}
}

func main() {
	router := httprouter.New()
	ctrl := &controller{
		lock:          new(sync.Mutex),
		notifications: []Notification{},
	}
	router.POST("/", handleFallible(ctrl.receive))
	router.GET("/", handleFallible(ctrl.report))
	log.Print("About to start serving on port 1984")
	http.ListenAndServe(":1984", router)
}
