package main

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

func (lw *LlmWrangler) WarmupHost(host string) {
	var tr = &http.Transport{
		//	TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{
		Transport: tr,
		Timeout:   time.Minute * 2,
	}

	// Get total slots
	type healthResponse struct {
		Status          string `json:"status"`
		SlotsIdle       int    `json:"slots_idle"`
		SlotsProcessing int    `json:"slots_processing"`
	}
	req, _ := http.NewRequest("GET", "http://"+host+"/health", nil)
	req.Header.Set("Content-Type", "application/json")
	res, err := client.Do(req)
	if err != nil {
		log.Println("Can't contact llama.cpp server:", err)
		return
	}
	data, _ := io.ReadAll(res.Body)
	health := &healthResponse{}
	json.Unmarshal(data, health)

	// warm up each slot
	log.Println("Warming up", health.SlotsIdle, "llama.cpp slots on", host)

	lw.hostsLock.Lock()
	status, hostOk := lw.hosts[host]
	if hostOk {
		status.OpenSlots = health.SlotsIdle
		lw.hosts[host] = status
	}
	lw.hostsLock.Unlock()

	responseTime := make(chan int64)
	for i := 0; i < health.SlotsIdle; i++ {
		go lw.warmupLlama(host, responseTime)
	}

	go func() {
		// wait for warmup
		for i := 0; i < health.SlotsIdle; i++ {
			<-responseTime
		}

		time.Sleep(time.Second * 5) // cooldown

		// and one more to get the post-warmup response time
		go lw.warmupLlama(host, responseTime)

		log.Println("Warmed up response time for", host, ":", <-responseTime, "ms")
	}()
}

func (lw *LlmWrangler) warmupLlama(host string, responseTime chan int64) error {
	lw.hostsLock.Lock()
	status, hostOk := lw.hosts[host]
	if hostOk {
		status.UseSlot()
		lw.hosts[host] = status
	}
	lw.hostsLock.Unlock()

	config := Completion{}
	config.Cache_prompt = true
	config.Mirostat_eta = 0.1
	config.Mirostat_tau = 5
	config.N_predict = 400
	config.Repeat_last_n = 96
	config.Repeat_penalty = 1.00
	config.Slot_id = -1
	config.Stream = true
	config.Temperature = 1
	config.Tfs_z = 1
	config.Top_k = 0
	config.Top_p = 1
	config.Typical_p = 1
	config.Min_p = 0.1

	prompt, err := os.ReadFile(lw.warmupPromptFile)
	if err != nil {
		log.Fatal("Could not read warmup-prompt-file:", lw.warmupPromptFile)
	}

	config.Prompt = string(prompt)

	b, _ := json.Marshal(config)

	var tr = &http.Transport{
		//	TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{
		Transport: tr,
		Timeout:   0,
	}

	startTime := time.Now()
	req, _ := http.NewRequest("POST", "http://"+host+"/completion", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	res, err := client.Do(req)
	if err != nil {
		log.Println(err)
	}
	io.ReadAll(res.Body)
	if err != nil {
		log.Println(err)
	}
	res.Body.Close()

	loadTime := time.Since(startTime)
	lw.hostsLock.Lock()
	status, hostOk = lw.hosts[host]
	if hostOk {
		status.FreeSlot()
		status.ResponseTime = loadTime
		lw.hosts[host] = status
	}
	lw.hostsLock.Unlock()

	responseTime <- loadTime.Milliseconds()
	return err
}
