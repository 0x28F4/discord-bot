package voiceactivation

import (
	"errors"
	"sync"
	"time"

	logr "github.com/GRVYDEV/S.A.T.U.R.D.A.Y/log"
	"github.com/GRVYDEV/S.A.T.U.R.D.A.Y/stt/engine"
	eng "github.com/GRVYDEV/S.A.T.U.R.D.A.Y/stt/engine"
)

const (
	SampleRate       = 16000
	silenceThresh    = 100 * time.Millisecond
	inferenceMinSize = 24000
	inferenceWindow  = 10 * time.Second
)

var Logger = logr.New()

type Params struct {
	OnDocumentUpdate func(eng.Document)
	Transcriber      eng.Transcriber
}

type Engine struct {
	writeMutex     sync.Mutex
	inferenceMutex sync.Mutex
	currentWindow  *window
	windows        []window
	timer          *time.Timer
	OnTranscribe   func(t engine.Transcription)
	transcriber    eng.Transcriber
}

func New(params Params) (*Engine, error) {
	if params.Transcriber == nil {
		return nil, errors.New("you must supply a Transciber to create an engine")
	}

	return &Engine{
		currentWindow: nil,
		windows:       make([]window, 0),
		transcriber:   params.Transcriber,
	}, nil
}

func (e *Engine) Write(pcm []float32) {
	e.writeMutex.Lock()
	defer e.writeMutex.Unlock()
	if e.timer != nil {
		e.timer.Stop()
	}
	if e.currentWindow == nil {
		e.currentWindow = newWindow(pcm)
	} else {
		e.currentWindow.grow(pcm)
	}

	if len(e.currentWindow.pkt) > SampleRate*10 {
		e.finalize()
	}
	e.timer = time.AfterFunc(silenceThresh, func() {
		e.finalize()
	})
}

func (e *Engine) ForceTranscribe() {
	e.writeMutex.Lock()
	defer e.writeMutex.Unlock()
	if e.timer != nil {
		e.timer.Stop()
	}

	e.finalize()
}

func (e *Engine) finalize() {
	Logger.Debugf("finalizing pcm window of length %d\n", len(e.currentWindow.pkt))
	e.windows = append(e.windows, *e.currentWindow)
	e.currentWindow = nil
	e.inference()
}

func (e *Engine) inference() {
	payload := make([]float32, 0)
	numWindows := 0
	for i, window := range e.windows {
		if !window.touch {
			payload = append(payload, window.pkt...)
			numWindows += 1
			e.windows[i].touch = true
		}
	}
	if len(payload) < inferenceMinSize {
		// add padding until payload has length of inferenceMinimumSize
		payload = append(payload, make([]float32, inferenceMinSize-len(payload))...)
	}
	Logger.Debugf("captured %d windows for inference with a total length of %d", numWindows, len(payload))
	e.transcribe(payload)
}

func (e *Engine) transcribe(payload []float32) {
	e.inferenceMutex.Lock()
	defer e.inferenceMutex.Unlock()
	transcript, err := e.transcriber.Transcribe(payload)
	if err != nil {
		Logger.Error(err, "error running inference")
		return
	}

	Logger.Debugf("got transcript %+v", transcript)
	if e.OnTranscribe != nil {
		e.OnTranscribe(transcript)
	}
}
