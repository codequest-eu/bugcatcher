package main

import (
	"encoding/json"
	"time"

	"github.com/jinzhu/gorm"
)

type Error struct {
	ID           uint    `gorm:"primary_key" json:"id"`
	GroupingHash string  `json:"-"`
	ErrorClass   string  `json:"errorClass"`
	Location     string  `json:"location"`
	Severity     string  `json:"severity"`
	CreatedAt    int64   `json:"createdAt"`
	UpdatedAt    int64   `json:"updatedAt"`
	Events       []Event `json:"events"`
} // type Error

func (e *Error) BeforeCreate(scope *gorm.Scope) error {
	if err := scope.SetColumn("CreatedAt", time.Now().Unix()); err != nil {
		return err
	}
	return scope.SetColumn("UpdatedAt", time.Now().Unix())
}

func (e *Error) BeforeUpdate(scope *gorm.Scope) error {
	return scope.SetColumn("UpdatedAt", time.Now().Unix())
}

type Event struct {
	ID             uint `gorm:"primary_key" json:"id"`
	ErrorID        uint `sql:"type:integer REFERENCES errors(id) ON DELETE CASCADE ON UPDATE CASCADE"`
	Hostname       string
	Message        string
	SerializedData []byte
	CreatedAt      int64
} // type Event

func (e *Event) BeforeCreate(scope *gorm.Scope) error {
	return scope.SetColumn("CreatedAt", time.Now().Unix())
}

func (e *Event) GetEventData() (*EventData, error) {
	var ret EventData
	if err := json.Unmarshal(e.SerializedData, &ret); err != nil {
		return nil, err
	}
	return &ret, nil
}

func (e *Event) SetEventData(data EventData) error {
	serializedData, err := json.Marshal(&data)
	if err != nil {
		return err
	}
	e.SerializedData = serializedData
	return nil
}

func (e *Event) MarshalJSON() ([]byte, error) {
	eventData, err := e.GetEventData()
	if err != nil {
		return nil, err
	}
	return json.Marshal(
		&struct {
			Hostname   string          `json:"hostname,omitempty"`
			Message    string          `json:"message,omitempty"`
			StackTrace []StackFrame    `json:"stackTrace,omitempty"`
			Metadata   json.RawMessage `json:"metaData,omitempty"`
			CreatedAt  int64           `json:"createdAt,omitempty"`
		}{
			Hostname:   e.Hostname,
			Message:    e.Message,
			StackTrace: eventData.StackTrace,
			Metadata:   eventData.Metadata,
			CreatedAt:  e.CreatedAt,
		},
	)
}

type EventData struct {
	StackTrace []StackFrame    `json:"stackTrace,omitempty"`
	Metadata   json.RawMessage `json:"metaData,omitempty"`
} // type EventData

type StackFrame struct {
	StackFrameData
	Code []LineOfCode `json:"code,omitempty"`
} // type NotificationStackFrame
