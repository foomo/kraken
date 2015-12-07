package kraken

import (
	"bytes"
	"errors"
	"log"
	"net/http"
	"time"
)

type Prey struct {
	Id        string   `json:"id"`
	URL       string   `json:"url"`
	Priority  int      `json:"priority"`
	Errors    []string `json:"errors"`
	Status    string   `json:"status"`
	Time      int64    `json:"time"`
	Created   int64    `json:"created"`
	Completed int64    `json:"completed"`
	Method    string   `json:"verb"`
	Body      []byte   `json:"body"`
}

const (
	preyStatusWaiting    = "waiting"
	preyStatusProcessing = "processing"
	preyStatusRetry      = "retry"
	preyStatusFail       = "failed"
	preyStatusDone       = "done"
)

type PreyProcessingResult struct {
	Prey  *Prey
	Error error
	Time  int64
}

type Tentacle struct {
	Name            string
	Prey            map[string]*Prey
	Bandwidth       int
	UsedBandwidth   int
	Retry           int
	Queue           []string
	ChannelEntangle chan *Prey
	ChannelBurp     chan *PreyProcessingResult
	ChannelMove     chan int
	ChannelDie      chan int
	dead            bool
}

func NewTentacle(name string, bandwidth int, retry int) *Tentacle {
	t := &Tentacle{
		Name:            name,
		Bandwidth:       bandwidth,
		Retry:           retry,
		UsedBandwidth:   0,
		Prey:            make(map[string]*Prey),
		ChannelEntangle: make(chan *Prey),
		ChannelBurp:     make(chan *PreyProcessingResult),
		ChannelMove:     make(chan int),
		ChannelDie:      make(chan int),
		dead:            false,
	}
	log.Println("tentacle", t.Name, "is growing")
	go func() {
		for {
			select {
			case <-t.ChannelDie:
				log.Println("tentacle", t.Name, "is falling off")
				t.dead = true
				t.ChannelDie = nil
				t.ChannelEntangle = nil
				t.ChannelBurp = nil
				t.ChannelMove = nil
				return
			case <-t.ChannelMove:
				for t.UsedBandwidth < t.Bandwidth {
					nextPrey := t.nextPrey()
					if nextPrey != nil {
						nextPrey.Status = preyStatusProcessing
						t.UsedBandwidth++
						go t.kill(nextPrey)
					} else {
						break
					}
				}
			case newPrey := <-t.ChannelEntangle:
				newPrey.Status = preyStatusWaiting
				newPrey.Created = time.Now().UnixNano()
				t.Prey[newPrey.Id] = newPrey
				t.Queue = append(t.Queue, newPrey.Id)
				log.Println("tentacle", t.Name, "incoming prey", newPrey.Id)
			case processingResult := <-t.ChannelBurp:
				t.UsedBandwidth--
				processingResult.Prey.Time = processingResult.Time
				if processingResult.Error != nil {
					processingResult.Prey.Errors = append(processingResult.Prey.Errors, processingResult.Error.Error())
					if len(processingResult.Prey.Errors) < t.Retry {
						processingResult.Prey.Status = preyStatusRetry
					} else {
						t.markCompleteWithStatus(processingResult.Prey, preyStatusFail)
					}
				} else {
					t.markCompleteWithStatus(processingResult.Prey, preyStatusDone)
				}
			}
		}
	}()
	return t
}

func (t *Tentacle) markCompleteWithStatus(prey *Prey, status string) {
	prey.Status = status
	prey.Completed = time.Now().UnixNano()
	log.Println("tentacle", t.Name, "done with", prey.Id, "with status", status)
}

func (t *Tentacle) nextPrey() *Prey {
	for _, id := range t.Queue {
		p, _ := t.Prey[id]
		if p.Status == preyStatusRetry || p.Status == preyStatusWaiting {
			return p
		}
	}
	return nil
}
func (t *Tentacle) Die() {
	t.ChannelDie <- 1
}
func (t *Tentacle) Entangle(prey *Prey) {
	if !t.dead {
		t.ChannelEntangle <- prey
	}
}

// Move to kill
func (t *Tentacle) Move() {
	if !t.dead {
		t.ChannelMove <- 1
	}
}

// kill some prey
func (t *Tentacle) kill(prey *Prey) {
	if !t.dead {
		log.Println("tentacle", t.Name, "is about to kill", prey.Id, prey.URL)
		start := time.Now()
		method := prey.Method
		if len(method) == 0 {
			method = "GET"
		}
		req, err := http.NewRequest(method, prey.URL, bytes.NewReader([]byte(prey.Body)))
		if err == nil {
			resp, err := http.DefaultClient.Do(req)
			if err == nil {
				resp.Body.Close()
				if resp.StatusCode != http.StatusOK {
					err = errors.New("wrong response code " + resp.Status)
				}
			}
		}
		if t.ChannelBurp != nil {
			t.ChannelBurp <- &PreyProcessingResult{
				Prey:  prey,
				Error: err,
				Time:  time.Now().UnixNano() - start.UnixNano(),
			}
		}
	}
}
