package notifications

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type slackNotifier struct {
	BaseNotifier
}

type slackMessage struct {
	Text      string `json:"text"`
	IconEmoji string `json:"icon_emoji"`
}

func (sn *slackNotifier) Notify(text string) (string, error) {

	slackMessage := slackMessage{text, ":white_check_mark:"}
	jsonMessage, err := json.Marshal(slackMessage)
	if err != nil {
		return "", err
	}

	reader := bytes.NewReader(jsonMessage)

	url := sn.BaseNotifier.Config.SlackWebHook
	bodyType := "application/json"

	resp, err := http.Post(url, bodyType, reader)
	if err != nil {
		return "", err
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	response := fmt.Sprintf("%s", b)
	return response, nil
}
