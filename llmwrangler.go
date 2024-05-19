package main

import (
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"
)

type LlmWrangler struct {
	ListenPort        int
	hosts             map[string]time.Duration
	hostsLock         sync.RWMutex
	mapClientHost     map[string]string
	mapClientHostLock sync.RWMutex
}

func (lw *LlmWrangler) Init() {
	lw.hosts = make(map[string]time.Duration)
	lw.mapClientHost = make(map[string]string)
}

func (lw *LlmWrangler) Start() {
	http.Handle("/", http.FileServer(http.Dir("./static")))
	http.HandleFunc("/api/register", lw.handleRegister)
	http.HandleFunc("/api/unregister", lw.handleUnregister)
	http.HandleFunc("/api/hosts", lw.handleHosts)
	http.HandleFunc("/api/latency", lw.handleLatency)
	http.HandleFunc("/completion", lw.handleCompletion)

	err := http.ListenAndServe(":"+strconv.Itoa(lw.ListenPort), nil)
	if err != nil {
		log.Println(err)
	}
}

func (lw *LlmWrangler) RegisterHost(host string) {
	lw.hostsLock.Lock()
	lw.hosts[host] = 0
	lw.hostsLock.Unlock()
	log.Println("Host Registered:", host)
}

func (lw *LlmWrangler) UnregisterHost(host string) {
	lw.hostsLock.Lock()
	delete(lw.hosts, host)
	lw.hostsLock.Unlock()
	log.Println("Host Unregistered:", host)
}

func (lw *LlmWrangler) getLeastBusyHost() string {
	leastTime := time.Hour
	leastBusy := ""

	lw.hostsLock.RLock()
	for host, time := range lw.hosts {
		if time <= leastTime {
			leastTime = time
			leastBusy = host
		}
	}
	lw.hostsLock.RUnlock()

	return leastBusy
}

func (lw *LlmWrangler) AssignLeastBusyHost(clientIP string) string {
	return lw.getLeastBusyHost()
}

func (lw *LlmWrangler) AssignHost(clientIP string) string {
	lw.mapClientHostLock.Lock()
	host, ok := lw.mapClientHost[clientIP]
	if !ok {
		host = lw.getLeastBusyHost()
		if host != "" {
			lw.mapClientHost[clientIP] = host
		}
	}

	lw.hostsLock.RLock()
	_, hostOk := lw.hosts[host]
	lw.hostsLock.RUnlock()
	if !hostOk {
		host = lw.getLeastBusyHost()
		if host != "" {
			lw.mapClientHost[clientIP] = host
		}
	}

	lw.mapClientHostLock.Unlock()
	return host
}
