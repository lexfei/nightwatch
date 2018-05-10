package nightwatch

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"nightwatch/monitor"
)

// MonitorInfo represents status of a monitor.
// This is used by show and list commands.
type MonitorInfo struct {
	ID       int    `json:"id,string"`
	Name     string `json:"name"`
	Running  bool   `json:"running"`
	Failing  bool   `json:"failing"`
	Status   string `json:"status"`
	Times    int64  `json:"times"`
	FailedAt string `json:"failedAt"`
}

func handleMonitor(w http.ResponseWriter, r *http.Request) {
	// guaranteed no error by mux.
	id, _ := strconv.Atoi(mux.Vars(r)["id"])

	m := monitor.FindMonitor(id)
	if m == nil {
		http.NotFound(w, r)
		return
	}

	if r.Method == http.MethodGet {
		mi := &MonitorInfo{
			ID:       m.ID(),
			Name:     m.Name(),
			Running:  m.Running(),
			Failing:  m.Failing(),
			Status:   m.Status(),
			Times:    m.Times(),
			FailedAt: m.FailedAt(),
		}
		data, err := json.Marshal(mi)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.Write(data)
		return
	}

	if r.Method == http.MethodDelete {
		m.Stop()
		monitor.Unregister(m)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "invalid method", http.StatusBadRequest)
		return
	}

	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	switch strings.TrimSpace(string(data)) {
	case "start":
		if err := m.Start(); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	case "stop":
		m.Stop()
	default:
		http.Error(w, "unknown action", http.StatusBadRequest)
	}
}
