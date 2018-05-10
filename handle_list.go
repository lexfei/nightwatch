package nightwatch

import (
	"encoding/json"
	"net/http"

	"nightwatch/monitor"
)

// List represents JSON response for list command.
type List []*MonitorInfo

func handleList(w http.ResponseWriter, r *http.Request) {
	l := make(List, 0)
	for _, m := range monitor.ListMonitors() {
		l = append(l, &MonitorInfo{
			ID:       m.ID(),
			Name:     m.Name(),
			Running:  m.Running(),
			Failing:  m.Failing(),
			Status:   m.Status(),
			Times:    m.Times(),
			FailedAt: m.FailedAt(),
		})
	}

	data, err := json.Marshal(l)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Write(data)
}
