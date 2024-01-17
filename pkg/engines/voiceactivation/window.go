package voiceactivation

import "time"

type window struct {
	pkt       []float32
	timestamp time.Time
	touch     bool
}

func newWindow(pkt []float32) *window {
	return &window{
		pkt:       pkt,
		timestamp: time.Now(),
	}
}

func (w *window) grow(pkt []float32) {
	w.pkt = append(w.pkt, pkt...)
}
