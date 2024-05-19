package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"time"
)

func (lw *LlmWrangler) handleRegister(w http.ResponseWriter, r *http.Request) {
	b, err := io.ReadAll(r.Body)
	if err != nil {
		log.Println(err)
	}

	res := struct {
		Host string `json:"host"`
	}{}
	json.Unmarshal(b, &res)

	lw.RegisterHost(res.Host)
}

func (lw *LlmWrangler) handleUnregister(w http.ResponseWriter, r *http.Request) {
	b, err := io.ReadAll(r.Body)
	if err != nil {
		log.Println(err)
	}

	res := struct {
		Host string `json:"host"`
	}{}
	json.Unmarshal(b, &res)

	lw.UnregisterHost(res.Host)
}

func (lw *LlmWrangler) handleHosts(w http.ResponseWriter, r *http.Request) {
	lw.hostsLock.RLock()
	out, _ := json.Marshal(lw.hosts)
	lw.hostsLock.RUnlock()

	w.Write(out)
}

func (lw *LlmWrangler) handleLatency(w http.ResponseWriter, r *http.Request) {
	minTime := time.Hour * 24

	lw.hostsLock.RLock()
	for _, time := range lw.hosts {
		if minTime > time && time != 0 {
			minTime = time
		}

	}
	lw.hostsLock.RUnlock()

	out, _ := json.Marshal(minTime.Milliseconds())
	w.Write(out)
}

func (lw *LlmWrangler) handleCompletion(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()

	b, err := io.ReadAll(r.Body)
	if err != nil {
		log.Println(err)
	}

	reader := io.NopCloser(bytes.NewBuffer(b))
	r.Body = reader

	clientIP := r.Header.Get("X-Real-IP")
	assignedHost := lw.AssignLeastBusyHost(clientIP)
	if assignedHost == "" {
		log.Println("Backend unavailable")
		http.Error(w, "Backend unavailable", http.StatusBadGateway)
		return
	}

	log.Println("Assigned", clientIP, assignedHost)
	r.Host = assignedHost

	r.URL.Scheme = "http"
	r.URL.Host = assignedHost

	r.RequestURI = ""
	var client = &http.Client{
		Timeout: 0,
	}

	res, err := client.Do(r)
	if err != nil {
		log.Println(err)
		http.Error(w, "Backend unavailable (refused)", http.StatusBadGateway)
		return
	}
	defer res.Body.Close()

	for k, vals := range res.Header {
		for _, v := range vals {
			w.Header().Add(k, v)
		}
	}

	buf := bufio.NewReader(res.Body)

	for {
		n, err := io.Copy(w, buf)
		if n == 0 {
			break
		}
		if err != nil {
			log.Println("Backend reading issue: ", err)
			break
		}

	}

	lw.hostsLock.Lock()
	lw.hosts[r.Host] = time.Since(startTime)
	lw.hostsLock.Unlock()
}
