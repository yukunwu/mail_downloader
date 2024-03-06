package main

import (
	"time"
)

type Filter struct {
	Before time.Time
	Since  time.Time
	MinUID uint32
	MaxUID uint32
	Folder string
}
type EmailInfo struct {
	ID         uint64    `json:"id"`
	UID        uint32    `json:"uid" validate:"required"`
	Date       time.Time `json:"date"`
	From       string    `json:"from"`
	To         string    `json:"to"`
	BusinessID string    `json:"business_id"`
	Subject    string    `json:"subject"`
	Content    string    `json:"content"`
	Label      uint8     `json:"label"`
	TaskID     int64     `json:"task_id"`
	Attachment string    `json:"attachment"`
	Folder     string    `json:"folder"`
}
