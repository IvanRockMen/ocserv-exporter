package main

import (
	"flag"
	"time"

	"github.com/criteo/ocserv-exporter/lib/occtl"
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
)

var (
	occtlStatusScrapeError = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "occtl_status_scrape_error_total",
		Help: "Total number of errors that occurred when calling occtl show status.",
	}, []string{})
	occtlUsersScrapeError = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "occtl_users_scrape_error_total",
		Help: "Total number of errors that occurred when calling occtl show users.",
	}, []string{})
	ocservStartTime = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "ocserv_start_time_seconds",
		Help: "Start time of ocserv since unix epoch in seconds.",
	}, []string{})
	ocservActiveSessions = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "ocserv_active_sessions",
		Help: "Current number of users connected.",
	}, []string{})
	ocservHandledSessions = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "ocserv_handled_sessions",
		Help: "Total number of sessions handled since server is up.",
	}, []string{})
	ocservIPsBanned = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "ocserv_ips_banned",
		Help: "Total number of IPs banned.",
	}, []string{})
	ocservTotalAuthenticationFailures = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "ocserv_total_authentication_failures",
		Help: "Total number of authentication failures since server is up.",
	}, []string{})
	ocservSessionsHandled = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "ocserv_sessions_handled",
		Help: "Total number of sessions handled since last stats reset.",
	}, []string{})
	ocservTimedOutSessions = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "ocserv_timed_out_sessions",
		Help: "Total number of timed out sessions since last stats reset.",
	}, []string{})
	ocservTimedOutIdleSessions = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "ocserv_timed_out_idle_sessions",
		Help: "Total number of sessions timed out (idle) since last stats reset.",
	}, []string{})
	ocservClosedDueToErrorSessions = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "ocserv_closed_error_sessions",
		Help: "Total number of sessions closed due to error since last stats reset.",
	}, []string{})
	ocservAuthenticationFailures = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "ocserv_authentication_failures",
		Help: "Total number of authentication failures since last stats reset.",
	}, []string{})
	ocservAverageAuthTime = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "ocserv_average_auth_time_seconds",
		Help: "Average time in seconds spent to authenticate users since last stats reset.",
	}, []string{})
	ocservMaxAuthTime = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "ocserv_max_auth_time_seconds",
		Help: "Maximum time in seconds spent to authenticate users since last stats reset.",
	}, []string{})
	ocservAverageSessionTime = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "ocserv_average_session_time_seconds",
		Help: "Average session time in seconds since last stats reset.",
	}, []string{})
	ocservMaxSessionTime = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "ocserv_max_session_time_seconds",
		Help: "Max session time in seconds since last stats reset.",
	}, []string{})
	ocservTX = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "ocserv_tx_bytes",
		Help: "Total TX usage in bytes since last stats reset.",
	}, []string{})
	ocservRX = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "ocserv_rx_bytes",
		Help: "Total RX usage in bytes since last stats reset.",
	}, []string{})
	ocservUserTX = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "ocserv_user_tx_bytes",
		Help: "Total TX usage in bytes of a user.",
	}, []string{"username", "remote_ip", "mtu", "ocserv_ipv4", "ocserv_ipv6", "device", "user_agent"})
	ocservUserRX = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "ocserv_user_rx_bytes",
		Help: "Total RX usage in bytes of a user.",
	}, []string{"username", "remote_ip", "mtu", "ocserv_ipv4", "ocserv_ipv6", "device", "user_agent"})
	ocservUserStartTime = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "ocserv_user_start_time_seconds",
		Help: "Start time of user session since unix epoch in seconds.",
	}, []string{"username", "remote_ip", "mtu", "ocserv_ipv4", "ocserv_ipv6", "device", "user_agent"})
)

func main() {
	var (
		interval   = flag.Duration("interval", 30*time.Second, "Delay between occtl scrape.")
		listen     = flag.String("listen", "127.0.0.1:8000", "Prometheus HTTP listen IP and port.")
		socketPath = flag.String("socket", "/var/run/occtl.socket", "Path to connect socket")
	)
	flag.Parse()

	prometheus.MustRegister(
		occtlStatusScrapeError,
		occtlUsersScrapeError,
		ocservStartTime,
		ocservActiveSessions,
		ocservHandledSessions,
		ocservIPsBanned,
		ocservTotalAuthenticationFailures,
		ocservSessionsHandled,
		ocservTimedOutSessions,
		ocservTimedOutIdleSessions,
		ocservClosedDueToErrorSessions,
		ocservAuthenticationFailures,
		ocservAverageAuthTime,
		ocservMaxAuthTime,
		ocservAverageSessionTime,
		ocservMaxSessionTime,
		ocservTX,
		ocservRX,
		ocservUserTX,
		ocservUserRX,
		ocservUserStartTime,
	)

	occtlCli, err := occtl.NewClient(&occtl.OcctlCommander{}, socketPath)
	if err != nil {
		log.Fatalf("Failed to initialize occtl client: %v", err)
	}

	exporter := NewExporter(occtlCli, *listen, *interval)
	err = exporter.Run()
	if err != nil {
		log.Fatal(err)
	}
}
