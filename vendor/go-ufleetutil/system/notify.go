package system

import (
	"encoding/json"
	"errors"
	"go-ufleetutil/util"
	"io/ioutil"
	"strings"

	"log"
)

// Mail 邮件通知
type Mail struct {
	host      string
	Group     string   `json:"group"`
	Subject   string   `json:"subject"`
	Content   string   `json:"content"`
	EmailList []string `json:"emailList"`
}

// NewMailClient 新建Audit
func NewMailClient(host string) Mail {
	var mail Mail
	mail.host = host
	return mail
}

// Send 发送邮件
func (mail *Mail) Send(token string) error {
	var err error
	url := mail.host + "/v1/messagehub/email"
	b, err := json.Marshal(mail)
	if err != nil {
		return err
	}
	body := ioutil.NopCloser(strings.NewReader(string(b)))
	response, err := util.CreateRequest("POST", url, body, token)

	if err != nil {
		log.Println(string(response), err)
		return errors.New(string(response))
	}
	return err
}

// WebSocketNotice ws 通知
type WebSocketNotice struct {
	Module     string          `json:"module"`
	NoticeType string          `json:"type"`
	Action     string          `json:"action"`
	Content    json.RawMessage `json:"content"`
	Key        string          `json:"key"`
	host       string
}

// NewWsClient 新建Audit
func NewWsClient(host string) WebSocketNotice {
	var ws WebSocketNotice
	ws.host = host
	return ws
}

// Send 发送
func (ws *WebSocketNotice) Send() error {
	var err error
	ws.Module = "uflow"
	url := ws.host + "/v1/messagehub/ws/notice"
	b, err := json.Marshal(ws)
	if err != nil {
		return err
	}
	body := ioutil.NopCloser(strings.NewReader(string(b)))
	response, err := util.CreateRequest("POST", url, body, "")
	if err != nil {
		log.Println(string(response))
		return errors.New(string(response))
	}
	return err
}
