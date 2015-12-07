package kraken

import (
	"errors"
	"time"
)

type Kraken struct {
	tentacles map[string]*Tentacle
}

func NewKraken() *Kraken {
	k := &Kraken{
		tentacles: make(map[string]*Tentacle),
	}
	go func() {
		for {
			time.Sleep(time.Millisecond * 100)
			for _, tentacle := range k.tentacles {
				tentacle.Move()
			}
		}
	}()
	return k
}

func (k *Kraken) SqueezeTentacle(name string, bandwidth int, retry int) error {
	tentacle, ok := k.tentacles[name]
	if !ok {
		return errors.New("tentacle not found")
	}
	tentacle.Bandwidth = bandwidth
	tentacle.Retry = retry
	return nil

}

func (k *Kraken) GrowTentacle(name string, bandwidth int, retry int) {
	k.CutOffTentacle(name)
	k.tentacles[name] = NewTentacle(name, bandwidth, retry)
}

func (k *Kraken) CutOffTentacle(name string) {
	existingTentacle, exists := k.tentacles[name]
	if exists {
		existingTentacle.Die()
		delete(k.tentacles, name)
	}
}

func (k *Kraken) getTentacle(name string) (tentacle *Tentacle, ok bool) {
	tentacle, ok = k.tentacles[name]
	return tentacle, ok
}

func (k *Kraken) Catch(tentacleName string, prey *Prey) error {
	t, ok := k.getTentacle(tentacleName)
	if false == ok {
		return errors.New("I do not have a tentacle named " + tentacleName)
	}
	t.Entangle(prey)
	return nil
}
