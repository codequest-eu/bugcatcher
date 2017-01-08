package main

import (
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
)

type Notification struct {
	APIKey string              `json:"apiKey"`
	Events []NotificationEvent `json:"events,omitempty"`
} // type Notification

func (n *Notification) ToErrors() ([]Error, error) {
	var ret []Error
	for _, event := range n.Events {
		errors, err := event.ToErrors()
		if err != nil {
			return nil, err
		}
		ret = append(ret, errors...)
	}
	return ret, nil
}

type NotificationEvent struct {
	Exceptions   []Exception `json:"exceptions,omitempty"`
	Severity     string      `json:"severity,omitempty"`
	GroupingHash string      `json:"groupingHash,omitempty"`
	Device       struct {
		Hostname string `json:"hostname,omitempty"`
	} `json:"device,omitempty"`
	Metadata json.RawMessage `json:"metaData,omitempty"`
} // type NotificationEvent

func (ne *NotificationEvent) ToErrors() ([]Error, error) {
	errors := make([]Error, len(ne.Exceptions))
	for pos, exception := range ne.Exceptions {
		maybeError, err := exception.ToError(ne)
		if err != nil {
			return nil, err
		}
		errors[pos] = *maybeError
	}
	return errors, nil
}

type Exception struct {
	ErrorClass string                   `json:"errorClass,omitempty"`
	Message    string                   `json:"message,omitempty"`
	StackTrace []NotificationStackFrame `json:"stacktrace,omitempty"`
} // type Exception

func (ex *Exception) ToError(ne *NotificationEvent) (*Error, error) {
	stackFrames := make([]StackFrame, len(ex.StackTrace))
	for pos, frame := range ex.StackTrace {
		stackFrame, err := frame.StackFrame()
		if err != nil {
			return nil, err
		}
		stackFrames[pos] = *stackFrame
	}
	event := &Event{
		Hostname: ne.Device.Hostname,
		Message:  ex.Message,
	}
	event.SetEventData(EventData{
		StackTrace: stackFrames,
		Metadata:   ne.Metadata,
	})
	ret := &Error{
		GroupingHash: ex.GroupingHash(ne),
		ErrorClass:   ex.ErrorClass,
		Severity:     ne.Severity,
		Location:     ex.Location(),
		Events:       []Event{*event},
	}
	return ret, nil
}

func (e *Exception) Location() string {
	if len(e.StackTrace) == 0 {
		return "unknown location"
	}
	topFrame := e.StackTrace[0]
	return fmt.Sprintf("%s:%d", topFrame.File, topFrame.LineNumber)
}

func (e *Exception) GroupingHash(ne *NotificationEvent) string {
	source := ne.GroupingHash
	if source == "" {
		source = fmt.Sprintf(
			"%s|%s|%s",
			e.ErrorClass,
			e.Location(),
			ne.Severity,
		)
	}
	return fmt.Sprintf("%x", sha1.Sum([]byte(source)))
}

type StackFrameData struct {
	File       string `json:"file,omitempty"`
	LineNumber int    `json:"lineNumber,omitempty"`
	Method     string `json:"method,omitempty"`
} // type StackFrameData

type NotificationStackFrame struct {
	StackFrameData
	Code map[string]string `json:"code,omitempty"`
} // type NotificationStackFrame

func (nsf *NotificationStackFrame) StackFrame() (*StackFrame, error) {
	loc, err := nsf.linesOfCode()
	if err != nil {
		return nil, err
	}
	return &StackFrame{
		StackFrameData: nsf.StackFrameData,
		Code:           loc,
	}, nil
}

func (sf *NotificationStackFrame) linesOfCode() ([]LineOfCode, error) {
	ret := make([]LineOfCode, len(sf.Code))
	var lineNumbers []int
	for stringNumber := range sf.Code {
		lineNumber, err := strconv.Atoi(stringNumber)
		if err != nil {
			return nil, err
		}
		lineNumbers = append(lineNumbers, lineNumber)
	}
	sort.Ints(lineNumbers)
	for pos, lineNumber := range lineNumbers {
		ret[pos] = LineOfCode{
			LineNumber: lineNumber,
			Content:    sf.Code[fmt.Sprintf("%d", lineNumber)],
		}
	}
	return ret, nil
}

type LineOfCode struct {
	LineNumber int    `json:"lineNumber"`
	Content    string `json:"content"`
} // type LineOfCode
