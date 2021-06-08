package models

// Message is used to marshal/unmarshal Rabbit messages.
type Message struct {
	Text   string `json:"text"`
	Source string `json:"source"`
	Time   int64  `json:"time"`
}
