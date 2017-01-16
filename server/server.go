package main

import (
	"crypto/subtle"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

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
	if notification.APIKey != os.Getenv("API_KEY") {
		log.Printf(
			"API key %q not recognized, expected %q",
			notification.APIKey,
			os.Getenv("API_KEY"),
		)
		return nil
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
	query := c.db.Order("updated_at DESC").Preload("Events", func(db *gorm.DB) *gorm.DB {
		return db.Order("created_at DESC")
	})
	if err := query.Find(&errors).Error; err != nil {
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

func (c *controller) deleteError(w http.ResponseWriter, r *http.Request, params httprouter.Params) error {
	stringID := params.ByName("id")
	intID, err := strconv.Atoi(stringID)
	if err != nil {
		return err
	}
	return c.db.Where("id = ?", intID).Delete(&Error{}).Error
}

func (c *controller) processError(reportedError *Error) error {
	var matches []Error
	query := c.db.Where("grouping_hash = ?", reportedError.GroupingHash)
	if err := query.Find(&matches).Error; err != nil {
		return err
	}
	if len(matches) == 0 {
		return c.db.Save(reportedError).Error
	}
	match := matches[0]
	event := reportedError.Events[0]
	event.ErrorID = match.ID
	match.UpdatedAt = time.Now().Unix()
	tx := c.db.Begin()
	if err := tx.Save(&event).Error; err != nil {
		tx.Rollback()
		return err
	}
	if err := tx.Save(&match).Error; err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit().Error
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

func withBasicAuth(inner httprouter.Handle) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
		user, pass, ok := r.BasicAuth()
		apiKey := []byte(os.Getenv("API_KEY"))
		if !ok || subtle.ConstantTimeCompare([]byte(user), apiKey) != 1 || subtle.ConstantTimeCompare([]byte(pass), apiKey) != 1 {
			w.Header().Set("WWW-Authenticate", `Basic realm="Who are you?"`)
			w.WriteHeader(401)
			w.Write([]byte("Unauthorised.\n"))
			return
		}
		inner(w, r, params)
	}
}

func indexHTML(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	http.ServeFile(w, r, "/frontend/index.html")
}

func main() {
	router := httprouter.New()
	db := dbOrDie()
	defer db.Close()
	ctrl := &controller{db: db}
	router.GET("/", withBasicAuth(indexHTML))
	router.POST("/", handleFallible(ctrl.receive))
	router.GET("/errors", withBasicAuth(handleFallible(ctrl.listErrors)))
	router.DELETE("/errors/:id", withBasicAuth(handleFallible(ctrl.deleteError)))
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
