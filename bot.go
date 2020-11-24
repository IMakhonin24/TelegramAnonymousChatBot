package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"
)

// UpdateT ...
type UpdateT struct {
	Ok     bool            `json:"ok"`
	Result []UpdateResultT `json:"result"`
}

// UpdateResultT ...
type UpdateResultT struct {
	UpdateID int                  `json:"update_id"`
	Message  UpdateResultMessageT `json:"message"`
}

// UpdateResultMessageT ...
type UpdateResultMessageT struct {
	MessageID int               `json:"message_id"`
	From      UpdateResultFromT `json:"from"`
	Chat      UpdateResultChatT `json:"chat"`
	Date      int               `json:"date"`
	Text      string            `json:"text"`
}

// UpdateResultFromT ...
type UpdateResultFromT struct {
	ID        int    `json:"id"`
	IsBot     bool   `json:"is_bot"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Username  string `json:"username"`
	Language  string `json:"language_code"`
}

// UpdateResultChatT ...
type UpdateResultChatT struct {
	ID        int    `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Username  string `json:"username"`
	Type      string `json:"type"`
}

// SendMessageResponseT ...
type SendMessageResponseT struct {
	Ok     bool               `json:"ok"`
	Result SendMessageResultT `json:"result"`
}

// SendMessageResultT ...
type SendMessageResultT struct {
	MessageID int                    `json:"message_id"`
	From      SendMessageResultFromT `json:"from"`
	Chat      SendMessageResultChatT `json:"chat"`
	Date      int                    `json:"date"`
	Text      string                 `json:"text"`
}

// SendMessageResultFromT ...
type SendMessageResultFromT struct {
	ID        int    `json:"id"`
	IsBot     bool   `json:"is_bot"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

//SendMessageResultChatT ...
type SendMessageResultChatT struct {
	ID        int    `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Username  string `json:"username"`
	Type      string `json:"type"`
}

const baseTelegramURL = "https://api.telegram.org"
const telegramToken = "1317662114:AAG06lNpLwX-Sy42NAek6WhKFlgtFSzvgA4"
const getUpdatesURI = "getUpdates"
const sendMessageURI = "sendMessage"

const keywordStart = "/start"
const keywordHelp = "/help"
const keywordStop = "/stop"
const keywordJoinChat = "/join_chat"

var updatesOffset = 0
var chatPairs = make(map[int]int)

func main() {
	timeTicker := time.NewTicker(2 * time.Second)
	for {
		<-timeTicker.C
		update, err := getUpdates()
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		for _, item := range update.Result {
			switch item.Message.Text {
			case keywordStart:
				text := "Привет, " + item.Message.From.FirstName + " " + item.Message.From.LastName + ". Вот, что я умею:"
				sendMessage(item.Message.Chat.ID, text)
				getHelp(item.Message)
			case keywordHelp:
				getHelp(item.Message)
			case keywordStop:
				stopChat(item.Message)
			case keywordJoinChat:
				joinRandomChat(item.Message)
			}
			sendToChat(item.Message)
		}
	}
}

// getUpdates
func getUpdates() (UpdateT, error) {
	url := baseTelegramURL + "/bot" + telegramToken + "/" + getUpdatesURI + "?offset=" + strconv.Itoa(updatesOffset)
	response := getResponse(url)

	update := UpdateT{}
	err := json.Unmarshal(response, &update)
	if err != nil {
		return update, err
	}
	for _, item := range update.Result {
		if item.UpdateID >= updatesOffset {
			updatesOffset = item.UpdateID + 1
		}
	}
	return update, nil
}

// sendMessage Telegram. Отправка сообщения
func sendMessage(chatID int, text string) (SendMessageResponseT, error) {
	url := baseTelegramURL + "/bot" + telegramToken + "/" + sendMessageURI
	url = url + "?chat_id=" + strconv.Itoa(chatID) + "&text=" + text
	response := getResponse(url)

	sendMessage := SendMessageResponseT{}
	err := json.Unmarshal(response, &response)
	if err != nil {
		return sendMessage, err
	}

	return sendMessage, nil
}

// getResponse
func getResponse(url string) []byte {
	response := make([]byte, 0)
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println(err)

		return response
	}

	defer resp.Body.Close()

	for true {
		bs := make([]byte, 1024)
		n, err := resp.Body.Read(bs)
		response = append(response, bs[:n]...)

		if n == 0 || err != nil {
			break
		}
	}

	return response
}

// sendToChat Отправить сообщение анониму
func sendToChat(message UpdateResultMessageT) {
	sendMessage(chatPairs[message.From.ID], message.Text)
}

// stopChat Отключиться от анонимного чата
func stopChat(message UpdateResultMessageT) {
	fromID := message.From.ID
	toID := chatPairs[fromID]
	delete(chatPairs, fromID)
	delete(chatPairs, toID)

	sendMessage(fromID, "Вы отключились от диалога")
	sendMessage(toID, "Собеседник отключился от диалога")
}

// joinRandomChat Подключиться к анонимному чату
func joinRandomChat(message UpdateResultMessageT) {
	for fromID, toID := range chatPairs {
		if toID == -1 {
			toID = message.From.ID
			chatPairs[fromID] = toID
			chatPairs[toID] = fromID
			sendMessage(fromID, "К вам подключился аноним, можете общаться.")
			sendMessage(toID, "Вы подключились к анонимному диалогу, можете общаться.")
			return
		}
	}
	chatPairs[message.From.ID] = -1
	sendMessage(message.From.ID, "Ожидайте, к вам скороподключится анонимный собеседник.")
}

// getHelp Вывести помощь. Список доступных команд
func getHelp(message UpdateResultMessageT) {
	sendMessage(message.From.ID, keywordJoinChat+" - Подключиться к анонимному чату")
}
