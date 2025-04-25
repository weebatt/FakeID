package models

type Task struct {
	ID      int64                  `json:"id"`
	Content map[string]interface{} `json:"content"`
}
