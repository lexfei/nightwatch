package alarm

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"

	"nightwatch"
	"nightwatch/actions"
)

const (
	defaultTimeout  = 30  // second
	defaultInterval = 240 // minute
)

var (
	client = &http.Client{}
)

type action struct {
	urlInit    *url.URL
	urlFail    *url.URL
	urlRecover *url.URL
	uuid       string
	module     string
	method     int
	interval   int
	message    string
	receiver   string
	timeout    time.Duration
}

type alarm struct {
	Uuid     string `json:"uuid"`
	Module   string `json:"module"`
	Title    string `json:"title"`
	Message  string `json:"message"`
	Method   int    `json:"method"`
	Receiver string `json:"receiver"`
	Interval int    `json:"interval"`
}

func processResponse(u *url.URL, resp *http.Response) error {
	defer func() {
		io.Copy(ioutil.Discard, resp.Body)
		resp.Body.Close()
	}()

	if 200 <= resp.StatusCode && resp.StatusCode < 300 {
		return nil
	}
	return fmt.Errorf("action:alarm:%s %s", u.String(), resp.Status)
}

func (a *action) request(u *url.URL, al *alarm) error {
	tu := *u
	header := make(http.Header)
	msg, err := json.Marshal(al)
	if err != nil {
		return err
	}

	var body io.ReadCloser
	var length int64
	header.Set("Content-Type", "application/json")
	length = int64(len(string(msg)))
	body = ioutil.NopCloser(strings.NewReader(string(msg)))
	req := &http.Request{
		Method:        "POST",
		URL:           &tu,
		Proto:         "HTTP/1.1",
		ProtoMajor:    1,
		ProtoMinor:    1,
		Header:        header,
		Body:          body,
		ContentLength: length,
		Host:          u.Host,
	}

	if a.timeout > 0 {
		ctx, cancel := context.WithTimeout(context.Background(), a.timeout)
		defer cancel()
		req = req.WithContext(ctx)
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	return processResponse(u, resp)
}

func (a *action) Init(name string) error {
	if a.urlInit == nil {
		return nil
	}

	al := alarm{
		Uuid:     a.uuid,
		Module:   a.module,
		Title:    "cr-monitor初始化通知",
		Message:  a.message,
		Method:   a.method,
		Receiver: a.receiver,
		Interval: a.interval,
	}

	return a.request(a.urlInit, &al)
}

func (a *action) Fail(name string, v float64) error {
	if a.urlFail == nil {
		return nil
	}

	al := alarm{
		Uuid:     a.uuid,
		Module:   a.module,
		Title:    "cr-monitor插件运行失败告警！",
		Message:  a.message,
		Method:   a.method,
		Receiver: a.receiver,
		Interval: a.interval,
	}
	return a.request(a.urlFail, &al)
}

func (a *action) Recover(name string, d time.Duration) error {
	if a.urlRecover == nil {
		return nil
	}

	al := alarm{
		Uuid:     a.uuid,
		Module:   a.module,
		Title:    "cr-monitor插件恢复正常通知",
		Message:  a.message,
		Method:   a.method,
		Receiver: a.receiver,
		Interval: a.interval,
	}

	return a.request(a.urlRecover, &al)
}

func (a *action) String() string {
	return fmt.Sprintf("action:alarm:%s:%s:%s",
		a.urlInit, a.urlFail, a.urlRecover)
}

func construct(params map[string]interface{}) (actions.Actor, error) {
	var uI, uF, uR *url.URL
	urlInit, err := nightwatch.GetString("url_init", params)
	switch err {
	case nil:
		uI, err = url.Parse(urlInit)
		if err != nil {
			return nil, err
		}
	case nightwatch.ErrNoKey:
	default:
		return nil, err
	}
	urlFail, err := nightwatch.GetString("url_fail", params)
	switch err {
	case nil:
		uF, err = url.Parse(urlFail)
		if err != nil {
			return nil, err
		}
	case nightwatch.ErrNoKey:
	default:
		return nil, err
	}
	urlRecover, err := nightwatch.GetString("url_recover", params)
	switch err {
	case nil:
		uR, err = url.Parse(urlRecover)
		if err != nil {
			return nil, err
		}
	case nightwatch.ErrNoKey:
	default:
		return nil, err
	}

	uuid, err := nightwatch.GetString("uuid", params)
	switch err {
	case nil:
	case nightwatch.ErrNoKey:
		uuid = "cr-monitor fail"
	default:
		return nil, err
	}

	module, err := nightwatch.GetString("module", params)
	switch err {
	case nil:
	case nightwatch.ErrNoKey:
		module = "ccs"
	default:
		return nil, err
	}

	method, err := nightwatch.GetInt("method", params)
	switch err {
	case nil:
	case nightwatch.ErrNoKey:
		method = 15
	default:
		return nil, err
	}

	interval, err := nightwatch.GetInt("interval", params)
	switch err {
	case nil:
	case nightwatch.ErrNoKey:
		interval = defaultInterval
	default:
		return nil, err
	}

	message, err := nightwatch.GetString("message", params)
	switch err {
	case nil:
	case nightwatch.ErrNoKey:
		message = "cr-monitor plugin fail"
	default:
		return nil, err
	}

	receiver, err := nightwatch.GetString("receiver", params)
	switch err {
	case nil:
	case nightwatch.ErrNoKey:
		receiver = "lkong"
	default:
		return nil, err
	}

	timeout, err := nightwatch.GetInt("timeout", params)
	switch err {
	case nil:
	case nightwatch.ErrNoKey:
		timeout = defaultTimeout
	default:
		return nil, err
	}

	return &action{
		urlInit:    uI,
		urlFail:    uF,
		urlRecover: uR,
		uuid:       uuid,
		module:     module,
		method:     method,
		interval:   interval,
		message:    message,
		receiver:   receiver,
		timeout:    time.Duration(timeout) * time.Second,
	}, nil
}

func init() {
	actions.Register("alarm", construct)
}
