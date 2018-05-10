package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/gorilla/mux"
	"nightwatch"
)

func newRequest(method, path string, body io.Reader) *http.Request {
	url := fmt.Sprintf("http://%s%s", *listenAddr, path)
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		panic(err)
	}
	req.Header.Set(nightwatch.VersionHeader, nightwatch.Version)
	return req
}

func readResponse(resp *http.Response) ([]byte, error) {
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		fmt.Fprintln(os.Stderr, "Server error:", resp.Status)
		return nil, errors.New(string(data))
	}

	return data, nil
}

func cmdList(r *mux.Router, args []string) error {
	client := &http.Client{}
	url, err := r.Get("list").URL()
	if err != nil {
		return err
	}

	resp, err := client.Do(newRequest(http.MethodGet, url.Path, nil))
	if err != nil {
		return err
	}
	data, err := readResponse(resp)
	if err != nil {
		return err
	}

	var l nightwatch.List
	if err := json.Unmarshal(data, &l); err != nil {
		return err
	}

	//fmt.Printf("%-8s  %-32s  Running  Failing\n", "ID", "Name")
	fmt.Printf("%-8s  %-20s  %-9s  %-6s  %-19s\n", "ID", "Name", "Times", "Status", "FailedAt")
	for _, i := range l {
		fmt.Printf("%-8d  %-20s  %-9d  %-6s  %-19s\n",
			i.ID, i.Name, i.Times, i.Status, i.FailedAt)
	}
	return nil
}

func cmdRegister(r *mux.Router, args []string) error {
	if len(args) != 1 {
		return errors.New("wrong number of arguments")
	}

	defs, err := loadYAML(args[0])
	if err != nil {
		return err
	}

	client := &http.Client{}
	url, err := r.Get("register").URL()
	if err != nil {
		return err
	}

	for _, md := range defs {
		data, err := json.Marshal(md)
		if err != nil {
			return err
		}
		req := newRequest(http.MethodPost, url.Path, bytes.NewBuffer(data))
		req.Header.Set("Content-Type", "application/json; charset=utf-8")
		resp, err := client.Do(req)
		if err != nil {
			return err
		}
		data2, err := readResponse(resp)
		if err != nil {
			return err
		}
		id := string(data2)
		fmt.Printf("%s is registered and started as monitor id=%s\n",
			md.Name, id)
	}

	return nil
}

func cmdShow(r *mux.Router, args []string) error {
	if len(args) != 1 {
		return errors.New("wrong number of arguments")
	}
	client := &http.Client{}
	url, err := r.Get("monitor").URL("id", args[0])
	if err != nil {
		return err
	}

	resp, err := client.Do(newRequest(http.MethodGet, url.Path, nil))
	if err != nil {
		return err
	}

	data, err := readResponse(resp)
	if err != nil {
		return err
	}

	var info nightwatch.MonitorInfo
	if err := json.Unmarshal(data, &info); err != nil {
		return err
	}
	fmt.Println("Name:", info.Name)
	fmt.Printf("Times: %v\n", info.Times)
	fmt.Printf("Status: %v\n", info.Status)
	fmt.Printf("FailedAt: %v\n", info.FailedAt)

	return nil
}

func cmdStart(r *mux.Router, args []string) error {
	if len(args) != 1 {
		return errors.New("wrong number of arguments")
	}
	client := &http.Client{}
	url, err := r.Get("monitor").URL("id", args[0])
	if err != nil {
		return err
	}
	req := newRequest(http.MethodPost, url.Path, strings.NewReader("start"))
	req.Header.Set("Content-Type", "text/plain; charset=utf-8")
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	_, err = readResponse(resp)
	if err != nil {
		return err
	}
	fmt.Println("Started.")
	return nil
}

func cmdStop(r *mux.Router, args []string) error {
	if len(args) != 1 {
		return errors.New("wrong number of arguments")
	}
	client := &http.Client{}
	url, err := r.Get("monitor").URL("id", args[0])
	if err != nil {
		return err
	}
	req := newRequest(http.MethodPost, url.Path, strings.NewReader("stop"))
	req.Header.Set("Content-Type", "text/plain; charset=utf-8")
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	_, err = readResponse(resp)
	if err != nil {
		return err
	}
	fmt.Println("Stopped.")
	return nil
}

func cmdUnregister(r *mux.Router, args []string) error {
	if len(args) != 1 {
		return errors.New("wrong number of arguments")
	}
	client := &http.Client{}
	url, err := r.Get("monitor").URL("id", args[0])
	if err != nil {
		return err
	}
	resp, err := client.Do(newRequest(http.MethodDelete, url.Path, nil))
	if err != nil {
		return err
	}
	_, err = readResponse(resp)
	if err != nil {
		return err
	}
	fmt.Println("Unregistered.")
	return nil
}

func cmdVerbosity(r *mux.Router, args []string) error {
	client := &http.Client{}
	url, err := r.Get("verbosity").URL()
	if err != nil {
		return err
	}
	if len(args) == 0 {
		req := newRequest(http.MethodGet, url.Path, nil)
		resp, err := client.Do(req)
		if err != nil {
			return err
		}
		data, err := readResponse(resp)
		if err != nil {
			return err
		}

		fmt.Println(string(data))
		return nil
	}

	req := newRequest(http.MethodPut, url.Path, strings.NewReader(args[0]))
	req.Header.Set("Content-Type", "text/plain; charset=utf-8")
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	_, err = readResponse(resp)
	if err == nil {
		fmt.Println("success.")
	}
	return err
}

func runCommand(cmd string, args []string) error {
	router := nightwatch.NewRouter()

	commands := map[string]func(r *mux.Router, args []string) error{
		"list":       cmdList,
		"register":   cmdRegister,
		"show":       cmdShow,
		"start":      cmdStart,
		"stop":       cmdStop,
		"unregister": cmdUnregister,
		"verbosity":  cmdVerbosity,
	}
	if f, ok := commands[cmd]; ok {
		return f(router, args)
	}
	return fmt.Errorf("no such command: %s", cmd)
}
