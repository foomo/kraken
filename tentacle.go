package kraken

import (
	"bytes"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
	"time"
)

// Prey - this is what a kraken hunts
type Prey struct {
	Id         string   `json:"id"`
	URL        string   `json:"url"`
	Priority   int      `json:"priority"`
	Errors     []string `json:"errors"`
	Status     string   `json:"status"`
	Time       int64    `json:"time"`
	Created    int64    `json:"created"`
	Completed  int64    `json:"completed"`
	Method     string   `json:"method"`
	Body       []byte   `json:"body"`
	Tags       []string `json:"tags"`
	RetryAfter int64    `json:"retryAfter"`
	Locks      []string `json:"locks,omitempty"`
}

type preyStatus string

const (
	preyStatusWaiting    preyStatus = "waiting"
	preyStatusProcessing            = "processing"
	preyStatusRetry                 = "retry"
	preyStatusRetryAfter            = "retryAfter"
	preyStatusFail                  = "failed"
	preyStatusDone                  = "done"
)

type PreyProcessingResult struct {
	Prey       *Prey
	Error      error
	Time       int64
	RetryAfter int64
}

type getLock struct {
	locks        []string
	chanCallback chan bool
}

type Tentacle struct {
	Name                 string
	Prey                 map[string]*Prey
	Bandwidth            int
	UsedBandwidth        int
	Retry                int
	Queue                []string
	locks                map[string]int64
	ChannelEntangle      chan *Prey
	ChannelBurp          chan *PreyProcessingResult
	ChannelMove          chan int
	ChannelDie           chan int
	dead                 bool
	channelLockGet       chan *getLock
	channelLockRelease   chan []string
	channelGetStatus     chan *TentacleStatus
	channelGetLocks      chan map[string]int64
	channelGetStatistics chan *TentacleStatistics
}

// TentacleStatus - status of a tentacle
type TentacleStatus struct {
	Name      string           `json:"name"`
	Retry     int              `json:"retry"`
	Bandwidth int              `json:"bandwidth"`
	Locks     map[string]int64 `json:"locks"`
	Prey      map[string]*Prey `json:"prey"`
}

//
type TentacleStatistics struct {
	Name       string               `json:"name"`
	Retry      int                  `json:"retry"`
	Bandwidth  int                  `json:"bandwidth"`
	PreyStates map[preyStatus]int64 `json:"preyStates"`
}

func NewTentacle(name string, bandwidth int, retry int) *Tentacle {
	t := &Tentacle{
		Name:                 name,
		Bandwidth:            bandwidth,
		Retry:                retry,
		UsedBandwidth:        0,
		Prey:                 make(map[string]*Prey),
		ChannelEntangle:      make(chan *Prey),
		ChannelBurp:          make(chan *PreyProcessingResult),
		ChannelMove:          make(chan int),
		ChannelDie:           make(chan int),
		dead:                 false,
		channelLockRelease:   make(chan []string),
		channelLockGet:       make(chan *getLock),
		channelGetStatus:     make(chan *TentacleStatus),
		channelGetStatistics: make(chan *TentacleStatistics),
		channelGetLocks:      make(chan map[string]int64),
		locks:                make(map[string]int64),
	}
	log.Println("tentacle", t.Name, "is growing")
	go func() {
		for {
			select {
			case locks := <-t.channelGetLocks:
				for id, lock := range t.locks {
					locks[id] = lock
				}
				t.channelGetLocks <- locks
			case getLock := <-t.channelLockGet:
				lockable := true
				for _, l := range getLock.locks {
					_, locked := t.locks[l]
					if locked {
						lockable = false
					}
				}
				if lockable {
					for _, l := range getLock.locks {
						t.locks[l] = time.Now().UnixNano()
					}
					getLock.chanCallback <- true
				} else {
					getLock.chanCallback <- false
				}
			case releases := <-t.channelLockRelease:
				for _, release := range releases {
					delete(t.locks, release)
				}
			}
		}
	}()
	go func() {
		for {
			select {
			case <-t.channelGetStatistics:
				ps := map[preyStatus]int64{
					preyStatusWaiting:    0,
					preyStatusProcessing: 0,
					preyStatusRetry:      0,
					preyStatusRetryAfter: 0,
					preyStatusFail:       0,
					preyStatusDone:       0,
				}
				for _, prey := range t.Prey {
					ps[preyStatus(prey.Status)]++
				}
				stats := &TentacleStatistics{
					Name:       t.Name,
					Retry:      t.Retry,
					Bandwidth:  t.Bandwidth,
					PreyStates: ps,
				}
				t.channelGetStatistics <- stats
			case s := <-t.channelGetStatus:
				s.Name = t.Name
				s.Retry = t.Retry
				s.Prey = t.Prey
				s.Locks = t.getLocks()
				s.Bandwidth = t.Bandwidth
				jsonBytes, err := json.Marshal(s)
				s = nil
				if err == nil {
					s = &TentacleStatus{}
					err = json.Unmarshal(jsonBytes, &s)
					if err != nil {
						s = nil
					}
				}
				t.channelGetStatus <- s
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
				newPrey.Status = string(preyStatusWaiting)
				newPrey.Created = time.Now().UnixNano()
				t.Prey[newPrey.Id] = newPrey
				t.Queue = append(t.Queue, newPrey.Id)
				log.Println("tentacle", t.Name, "incoming prey", newPrey.Id)
			case processingResult := <-t.ChannelBurp:
				if len(processingResult.Prey.Locks) > 0 {
					t.channelLockRelease <- processingResult.Prey.Locks
				}
				t.UsedBandwidth--
				processingResult.Prey.Time = processingResult.Time
				if processingResult.Error != nil {
					// call prey failed with an error
					processingResult.Prey.Errors = append(processingResult.Prey.Errors, processingResult.Error.Error())
					if len(processingResult.Prey.Errors) < t.Retry {
						if processingResult.RetryAfter == 0 {
							// retry immediately
							processingResult.Prey.Status = preyStatusRetry
						} else {
							// retry after timeout
							processingResult.Prey.RetryAfter = processingResult.RetryAfter
							processingResult.Prey.Status = preyStatusRetryAfter
						}
					} else {
						// call prey failed
						t.markCompleteWithStatus(processingResult.Prey, preyStatusFail)
					}
				} else {
					// call prey succeeded
					t.markCompleteWithStatus(processingResult.Prey, preyStatusDone)
				}
			}
		}
	}()
	return t
}

func (t *Tentacle) getLocks() map[string]int64 {
	locks := make(map[string]int64)
	t.channelGetLocks <- locks
	return <-t.channelGetLocks
}

func (t *Tentacle) getStatus() *TentacleStatus {
	status := &TentacleStatus{}
	t.channelGetStatus <- status
	status = <-t.channelGetStatus
	return status
}

func (t *Tentacle) GetStatistics() *TentacleStatistics {
	t.channelGetStatistics <- nil
	return <-t.channelGetStatistics
}

func (t *Tentacle) markCompleteWithStatus(prey *Prey, status string) {
	prey.Status = status
	prey.Completed = time.Now().UnixNano()
	log.Println("tentacle", t.Name, "done with", prey.Id, "with status", status)
}

func (t *Tentacle) nextPrey() *Prey {
	now := time.Now().UnixNano()
	for _, id := range t.Queue {
		p, _ := t.Prey[id]
		if p.Status == string(preyStatusRetry) || p.Status == string(preyStatusWaiting) || (p.Status == string(preyStatusRetryAfter) && p.RetryAfter < now) {
			if len(p.Locks) > 0 {
				// that prey wants locking
				chanCallback := make(chan bool)
				t.channelLockGet <- &getLock{
					locks:        p.Locks,
					chanCallback: chanCallback,
				}
				if <-chanCallback {
					return p
				}
			} else {
				return p
			}
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
		retryAfter := int64(0)
		log.Println("tentacle", t.Name, "is about to kill", prey.Id, prey.URL)
		start := time.Now()
		method := prey.Method
		if len(method) == 0 {
			method = "GET"
		}
		req, err := http.NewRequest(method, prey.URL, bytes.NewReader([]byte(prey.Body)))
		if err == nil {
			resp, clientError := http.DefaultClient.Do(req)
			setWrongResponseErr := func() {
				err = errors.New("wrong response code " + resp.Status)
			}
			if clientError == nil {
				resp.Body.Close()
				// check status codes
				switch resp.StatusCode {
				case http.StatusServiceUnavailable:
					// service unavailable, check for retry-after header
					retryAfterHeader := resp.Header.Get("Retry-After")
					if len(retryAfterHeader) > 0 {
						retryAfterSec, strconvError := strconv.Atoi(retryAfterHeader)
						if strconvError != nil {
							err = errors.New("invalid retry-after header " + retryAfterHeader)
						} else {
							retryAfter = time.Now().UnixNano() + int64(retryAfterSec)*1000000000
							err = errors.New("prey requires retry after: " + retryAfterHeader)
						}
					} else {
						setWrongResponseErr()
					}
				case http.StatusOK:
					err = nil
				default:
					setWrongResponseErr()
				}
			}
		}
		if t.ChannelBurp != nil {
			t.ChannelBurp <- &PreyProcessingResult{
				Prey:       prey,
				Error:      err,
				Time:       time.Now().UnixNano() - start.UnixNano(),
				RetryAfter: retryAfter,
			}
		}
	}
}
