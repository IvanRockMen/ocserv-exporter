package main

import (
	"net/http"
	"sync"
	"time"

	"github.com/criteo/ocserv-exporter/lib/occtl"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
)

type Exporter struct {
	interval    time.Duration
	listenAddr  string
	occtlCli    *occtl.Client
	promHandler http.Handler

	users []occtl.UsersMessage

	lock sync.Mutex
}

func NewExporter(occtlCli *occtl.Client, listenAddr string, interval time.Duration) *Exporter {
	return &Exporter{
		listenAddr:  listenAddr,
		interval:    interval,
		occtlCli:    occtlCli,
		promHandler: promhttp.Handler(),
	}
}

func (e *Exporter) Run() error {
	// run once to ensure we have data before starting the server
	e.update()

	go func() {
		for range time.Tick(e.interval) {
			e.update()
		}
	}()

	mux := http.NewServeMux()
	mux.HandleFunc("/metrics", e.metricsHandler)

	log.Infof("Listening on http://%s", e.listenAddr)
	return http.ListenAndServe(e.listenAddr, mux)
}

func (e *Exporter) update() {
	e.lock.Lock()
	defer e.lock.Unlock()

	e.updateStatus()
	e.updateUsers()
}

func (e *Exporter) updateStatus() {
	status, err := e.occtlCli.ShowStatus()
	if err != nil {
		log.Errorf("Failed to get server status: %v", err)
		occtlStatusScrapeError.WithLabelValues().Inc()
		ocservActiveSessions.Reset()
		ocservHandledSessions.Reset()
		ocservIPsBanned.Reset()
		return
	}
	ocservStartTime.WithLabelValues().Set(float64(status.RawUpSince))
	ocservActiveSessions.WithLabelValues().Set(float64(status.ActiveSessions))
	ocservHandledSessions.WithLabelValues().Set(float64(status.HandledSessions))
	ocservIPsBanned.WithLabelValues().Set(float64(status.IPsBanned))
	ocservTotalAuthenticationFailures.WithLabelValues().Set(float64(status.TotalAuthenticationFailures))
	ocservSessionsHandled.WithLabelValues().Set(float64(status.SessionsHandled))
	ocservTimedOutSessions.WithLabelValues().Set(float64(status.TimedOutSessions))
	ocservTimedOutIdleSessions.WithLabelValues().Set(float64(status.TimedOutIdleSessions))
	ocservClosedDueToErrorSessions.WithLabelValues().Set(float64(status.ClosedDueToErrorSessions))
	ocservAuthenticationFailures.WithLabelValues().Set(float64(status.AuthenticationFailures))
	ocservAverageAuthTime.WithLabelValues().Set(float64(status.RawAverageAuthTime))
	ocservMaxAuthTime.WithLabelValues().Set(float64(status.RawMaxAuthTime))
	ocservAverageSessionTime.WithLabelValues().Set(float64(status.RawAverageSessionTime))
	ocservMaxSessionTime.WithLabelValues().Set(float64(status.RawMaxSessionTime))
	ocservTX.WithLabelValues().Set(float64(status.RawTX))
	ocservRX.WithLabelValues().Set(float64(status.RawRX))
}

func (e *Exporter) updateUsers() {
	e.users = nil

	ocservUserTX.Reset()
	ocservUserRX.Reset()
	ocservUserStartTime.Reset()
	users, err := e.occtlCli.ShowUsers()
	if err != nil {
		log.Errorf("Failed to get users details: %v", err)
		occtlUsersScrapeError.WithLabelValues().Inc()
		return
	}

	for _, user := range users {
		ocservUserTX.WithLabelValues(user.Username, user.RemoteIP, user.MTU, user.VPNIPv4, user.VPNIPv6, user.Device, user.UserAgent).Set(float64(user.RawTX))
		ocservUserRX.WithLabelValues(user.Username, user.RemoteIP, user.MTU, user.VPNIPv4, user.VPNIPv6, user.Device, user.UserAgent).Set(float64(user.RawRX))
		ocservUserStartTime.WithLabelValues(user.Username, user.RemoteIP, user.MTU, user.VPNIPv4, user.VPNIPv6, user.Device, user.UserAgent).Set(float64(user.RawConnectedAt))
	}
}

func (e *Exporter) metricsHandler(rw http.ResponseWriter, r *http.Request) {
	e.lock.Lock()
	defer e.lock.Unlock()
	e.promHandler.ServeHTTP(rw, r)
}
