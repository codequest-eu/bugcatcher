package main

import "encoding/json"

type Notification struct {
	APIKey   string `json:"apiKey"`
	Notifier struct {
		Name    string `json:"name,omitempty"`
		Version string `json:"version,omitempty"`
		URL     string `json:"url,omitempty"`
	} `json:"notifier"`
	Events []struct {
		App struct {
			Version      string `json:"version,omitempty"`
			ReleaseStage string `json:"releaseStage,omitempty"`
			Type         string `json:"version,string"`
		} `json:"app,omitempty"`
		PayloadVersion string `json:"payloadVersion,omitempty"`
		Exceptions     []struct {
			ErrorClass string `json:"errorClass,omitempty"`
			Message    string `json:"message,omitempty"`
			StackTrace []struct {
				LineNumber int               `json:"lineNumber,omitempty"`
				Code       map[string]string `json:"code,omitempty"`
				File       string            `json:"file,omitempty"`
				Method     string            `json:"method,omitempty"`
			} `json:"stacktrace,omitempty"`
		} `json:"exceptions,omitempty"`
		Severity string `json:"severity,omitempty"`
		Device   struct {
			Hostname string `json:"hostname,omitempty"`
		} `json:"device,omitempty"`
		GroupingHash string          `json:"groupingHash,omitempty"`
		Metadata     json.RawMessage `json:"metaData,omitempty"`
	} `json:"events,omitempty"`
} // type Notification
