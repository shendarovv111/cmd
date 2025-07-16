package dto

type IncomingMessage struct {
	UserID   string  `json:"userId"`
	UserName string  `json:"userName,omitempty"`
	Text     *string `json:"text,omitempty"`
	Action   *string `json:"action,omitempty"`
}

type OutgoingMessage struct {
	UserID  string   `json:"userId"`
	Text    string   `json:"text"`
	Buttons []Button `json:"buttons"`
}

type Button struct {
	Text   string `json:"text"`
	Action string `json:"action"`
}

type OutgoingMessages struct {
	Messages []OutgoingMessage `json:"messages"`
}

func NewOutgoingMessage(userID, text string, buttons []Button) *OutgoingMessage {
	return &OutgoingMessage{
		UserID:  userID,
		Text:    text,
		Buttons: buttons,
	}
}

func NewOutgoingMessages(messages ...OutgoingMessage) *OutgoingMessages {
	return &OutgoingMessages{
		Messages: messages,
	}
}
