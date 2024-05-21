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
	hosts             map[string]HostStatus
	hostsLock         sync.RWMutex
	mapClientHost     map[string]string
	mapClientHostLock sync.RWMutex
	warmupPromptFile  string
}

func (lw *LlmWrangler) Init() {
	lw.hosts = make(map[string]HostStatus)
	lw.mapClientHost = make(map[string]string)
}

func (lw *LlmWrangler) Start() {
	http.Handle("/wrangler/", http.StripPrefix("/wrangler/", http.FileServer(http.Dir("./static"))))
	http.HandleFunc("/wrangler/api/register", lw.handleRegister)
	http.HandleFunc("/wrangler/api/unregister", lw.handleUnregister)
	http.HandleFunc("/wrangler/api/hosts", lw.handleHosts)
	http.HandleFunc("/wrangler/api/latency", lw.handleLatency)
	http.HandleFunc("GET /wrangler/api/test/{host}/{count}", lw.handleTest)

	http.HandleFunc("/", lw.handleLlamacpp)
	http.HandleFunc("/completion", lw.handleCompletion)
	http.HandleFunc("/v1/chat/completions", lw.handleCompletion)

	err := http.ListenAndServe(":"+strconv.Itoa(lw.ListenPort), nil)
	if err != nil {
		log.Println(err)
	}
}

func (lw *LlmWrangler) RegisterHost(host string) {
	lw.hostsLock.Lock()
	lw.hosts[host] = HostStatus{}
	lw.hostsLock.Unlock()
	log.Println("Host Registered:", host)

	lw.WarmupHost(host)
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
	for host, status := range lw.hosts {
		totalResponseTime := status.ResponseTime + status.ResponseTimeDebt
		if totalResponseTime <= leastTime && status.ResponseTime != 0 {
			leastTime = totalResponseTime
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
