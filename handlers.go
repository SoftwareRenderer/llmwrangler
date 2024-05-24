package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strconv"
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
	for _, status := range lw.hosts {
		if minTime > status.ResponseTime && status.ResponseTime != 0 {
			minTime = status.ResponseTime
		}

	}
	lw.hostsLock.RUnlock()

	out, _ := json.Marshal(minTime.Milliseconds())
	w.Write(out)
}

func (lw *LlmWrangler) handleHostAssignment(handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		clientIP := r.Header.Get("X-Real-IP")
		assignedHost := lw.AssignLeastBusyHost(clientIP)
		if assignedHost == "" {
			log.Println("Backend unavailable")
			http.Error(w, "Backend unavailable", http.StatusBadGateway)
			return
		}

		log.Println("Assigned", clientIP, assignedHost)
		r.Host = assignedHost
		r.URL.Host = assignedHost
		r.URL.Scheme = "http"
		if r.TLS != nil {
			r.URL.Scheme = "https"
		}

		handler(w, r)
	}
}

func (lw *LlmWrangler) handleLlamacpp(w http.ResponseWriter, r *http.Request) {
	b, err := io.ReadAll(r.Body)
	if err != nil {
		log.Println(err)
	}

	reader := io.NopCloser(bytes.NewBuffer(b))
	r.Body = reader

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
}

type peekResponseWriter struct {
	w          http.ResponseWriter
	statusCode int
}

func (p *peekResponseWriter) Header() http.Header {
	return p.w.Header()
}

func (p *peekResponseWriter) Write(b []byte) (int, error) {
	return p.w.Write(b)
}

func (p *peekResponseWriter) WriteHeader(code int) {
	p.statusCode = code
	p.w.WriteHeader(code)
}

func (lw *LlmWrangler) handleCompletion(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()

	lw.hostsLock.Lock()
	status, hostOk := lw.hosts[r.Host]
	if !hostOk {
		http.Error(w, "Host not available: "+r.Host, http.StatusBadGateway)
		return
	}
	status.UseSlot()
	lw.hosts[r.Host] = status
	lw.hostsLock.Unlock()

	wp := &peekResponseWriter{w, http.StatusOK}

	lw.handleLlamacpp(wp, r)

	lw.hostsLock.Lock()
	status, hostOk = lw.hosts[r.Host]
	if hostOk {
		status.FreeSlot()
		status.ResponseTime = time.Since(startTime)
	}
	lw.hosts[r.Host] = status
	lw.hostsLock.Unlock()
}

func (lw *LlmWrangler) handleTest(w http.ResponseWriter, r *http.Request) {
	host := r.PathValue("host")
	count, _ := strconv.Atoi(r.PathValue("count"))

	responseTime := make(chan int64)
	for i := 0; i < count; i++ {
		go lw.warmupLlama(host, responseTime)
	}

	var totalTime int64
	for i := 0; i < count; i++ {
		totalTime += <-responseTime
	}

	averageTime := int(totalTime) / count

	w.Write([]byte("Average: " + strconv.Itoa(averageTime) + "ms"))
}
