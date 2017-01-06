package main

import "github.com/jinzhu/gorm"

func setupSchema(db *gorm.DB) error {
	return db.Debug().AutoMigrate(&Error{}, &Event{}).Error
}
