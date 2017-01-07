package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"

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
	reportedErrors, err := notification.ToErrors()
	if err != nil {
		return err
	}
	for _, reportedError := range reportedErrors {
		if err := c.processError(&reportedError); err != nil {
			return err
		}
	}
	return nil
}

func (c *controller) listErrors(w http.ResponseWriter, r *http.Request, _ httprouter.Params) error {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	var errors []Error
	if err := c.db.Preload("Events").Find(&errors).Error; err != nil {
		return err
	}
	return encoder.Encode(&struct {
		APIKey string  `json:"apiKey"`
		Errors []Error `json:"errors"`
	}{
		APIKey: os.Getenv("API_KEY"),
		Errors: errors,
	})
}

func (c *controller) processError(reportedError *Error) error {
	var matches []Error
	query := c.db.Where(
		"error_class = ? AND location = ? AND severity = ?",
		reportedError.ErrorClass,
		reportedError.Location,
		reportedError.Severity,
	)
	if err := query.Find(&matches).Error; err != nil {
		return err
	}
	if len(matches) == 0 {
		return c.db.Save(reportedError).Error
	}
	event := reportedError.Events[0]
	event.ErrorID = matches[0].ID
	return c.db.Save(&event).Error
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

func indexHTML(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	http.ServeFile(w, r, "/frontend/index.html")
}

func main() {
	router := httprouter.New()
	db := dbOrDie()
	defer db.Close()
	ctrl := &controller{db: db.Debug()}
	router.GET("/", indexHTML)
	router.POST("/", handleFallible(ctrl.receive))
	router.GET("/errors", handleFallible(ctrl.listErrors))
	router.ServeFiles("/assets/*filepath", http.Dir("/frontend/assets"))
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
