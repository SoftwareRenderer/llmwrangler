package main

import "time"

type HostStatus struct {
	ResponseTime     time.Duration `json:"responseTime"`
	ResponseTimeDebt time.Duration `json:"responseTimeDebt"`
	OpenSlots        int           `json:"openSlots"`
}

func (h *HostStatus) UseSlot() {
	h.OpenSlots--
	if h.OpenSlots < 0 {
		h.ResponseTimeDebt += h.ResponseTime
	}
}

func (h *HostStatus) FreeSlot() {
	h.OpenSlots++
	if h.OpenSlots > 0 {
		h.ResponseTimeDebt += 0
	}
}
