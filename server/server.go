package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/julienschmidt/httprouter"
)

type controller struct {
	db *gorm.DB
}

func (c *controller) receive(w http.ResponseWriter, r *http.Request, _ httprouter.Params) error {
	defer r.Body.Close()
	var notification Notification
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&notification); err != nil {
		return err
	}
	errors, err := notification.ToErrors()
	if err != nil {
		return err
	}
	for _, error := range errors {
		if err := c.db.Create(&error).Error; err != nil {
			return err
		}
	}
	return nil
}

func (c *controller) report(w http.ResponseWriter, r *http.Request, _ httprouter.Params) error {
	w.Header().Add("Content-Type", "application/json")
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	var errors []Error
	if err := c.db.Preload("Events").Find(&errors).Error; err != nil {
		return err
	}
	return encoder.Encode(&errors)
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
	db := dbOrDie()
	defer db.Close()
	ctrl := &controller{db: db.Debug()}
	router.POST("/", handleFallible(ctrl.receive))
	router.GET("/", handleFallible(ctrl.report))
	log.Print("About to start serving on port 1984")
	http.ListenAndServe(":1984", router)
}

func dbOrDie() *gorm.DB {
	db, err := gorm.Open("sqlite3", "/data/sqlite3")
	if err != nil {
		log.Fatalf("error opening DB: %v", err)
	}
	if err := setupSchema(db); err != nil {
		log.Fatalf("error setting up DB: %v", err)
	}
	return db
}
