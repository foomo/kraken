package kraken

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"
)

type TentacleDefinition struct {
	Retry     int `json:"retry"`
	Bandwidth int `json:"bandwidth"`
}

type PreyDefinition struct {
	URL      string `json:"url"`
	Priority int    `json:"priority"`
}

type TentacleStatus struct {
	Name      string           `json:"name"`
	Retry     int              `json:"retry"`
	Bandwidth int              `json:"bandwidth"`
	Prey      map[string]*Prey `json:"prey"`
}

type ServerStatus struct {
	Tentacles map[string]*TentacleStatus `json:"tentacles"`
}

func NewServerStatus() *ServerStatus {
	s := &ServerStatus{
		Tentacles: make(map[string]*TentacleStatus),
	}
	return s
}

type Server struct {
	kraken *Kraken
}

func NewServer(k *Kraken) *Server {
	s := new(Server)
	s.kraken = k
	return s
}

func (s *Server) jsonResponse(w http.ResponseWriter, response interface{}) {
	jsonBytes, err := json.MarshalIndent(response, "", "\t")
	if err != nil {
		panic(err)
	}
	w.Write(jsonBytes)
}

func (s *Server) getServerStatus() *ServerStatus {
	status := NewServerStatus()
	for name, _ := range s.kraken.tentacles {
		status.Tentacles[name] = s.getTentacleStatus(name)
	}
	return status
}

func (s *Server) getTentacleStatus(name string) *TentacleStatus {
	tentacle, ok := s.kraken.tentacles[name]
	if ok {
		return &TentacleStatus{
			Name:      name,
			Retry:     tentacle.Retry,
			Prey:      tentacle.Prey,
			Bandwidth: tentacle.Bandwidth,
		}
	} else {
		return nil
	}
}

func decodeBody(r *http.Request, data interface{}) {
	jsonBytes, err := ioutil.ReadAll(r.Body)
	r.Body.Close()
	if err != nil {
		panic(err)
	}
	err = json.Unmarshal(jsonBytes, data)
	if err != nil {
		panic(err)
	}
}

const (
	help = `
Hello I am KRAKEN - URLs are my prey:

/status

	GET: get the status of this kraken


/tentacle/<name>

	PUT / POST : create or overwrite a new tentacle with body {"bandwidth": <int>, "retry": <int>}
	GET        : get the status of an existing tentacle
	DELETE     : get rid of the tentacle


/tentacle/<name>/<preyId>

	PUT/POST   : let me catch some prey with body { "url" : <string>, "priority" : <int> }

`
)

func (s *Server) help(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "text/plain;utf-8;")
	w.Write([]byte(help))
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	p := r.URL.Path
	switch p {
	case "/status":
		s.jsonResponse(w, s.getServerStatus())
		return
	default:
		if strings.HasPrefix(p, "/tentacle") {
			parts := strings.Split(p[1:], "/")
			if len(parts) == 2 {
				switch r.Method {
				case "POST":
				case "PUT":
					tentacleDef := &TentacleDefinition{}
					decodeBody(r, tentacleDef)
					s.kraken.GrowTentacle(parts[1], tentacleDef.Bandwidth, tentacleDef.Retry)
					s.jsonResponse(w, 1)
					return
				case "DELETE":
					s.kraken.CutOffTentacle(parts[1])
					s.jsonResponse(w, 1)
					return
				case "GET":
					s.jsonResponse(w, s.getTentacleStatus(parts[1]))
					return
				}
			} else if len(parts) == 3 {
				switch r.Method {
				case "POST":
				case "PUT":
					preyDefinition := &PreyDefinition{}
					decodeBody(r, preyDefinition)
					prey := &Prey{
						Id:       parts[2],
						URL:      preyDefinition.URL,
						Priority: preyDefinition.Priority,
					}
					s.jsonResponse(w, s.kraken.Catch(parts[1], prey))
					return
				}
			}
		}
	}
	s.help(w)
}

func (s *Server) ListenAndServe(address string) error {
	return http.ListenAndServe(address, s)
}
