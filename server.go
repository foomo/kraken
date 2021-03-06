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
	URL      string   `json:"url"`
	Priority int      `json:"priority"`
	Method   string   `json:"method,omitempty"`
	Body     []byte   `json:"body,omitempty"`
	Tags     []string `json:"tags,omitempty"`
	Locks    []string `json:"locks,omitempty"`
}

// ServerStatus - status of the whole server
type ServerStatus struct {
	Tentacles map[string]*TentacleStatus `json:"tentacles"`
}

type ServerStatistics struct {
	Tentacles map[string]*TentacleStatistics `json:"tentacles"`
}

func newServerStatistics() *ServerStatistics {
	s := &ServerStatistics{
		Tentacles: make(map[string]*TentacleStatistics),
	}
	return s
}

func newServerStatus() *ServerStatus {
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

func (s *Server) jsonResponse(code int, w http.ResponseWriter, response interface{}) {
	jsonBytes, err := json.MarshalIndent(response, "", "\t")
	if err != nil {
		panic(err)
	}
	w.WriteHeader(code)
	w.Write(jsonBytes)
}

func (s *Server) getServerStatus() *ServerStatus {
	status := newServerStatus()
	for name := range s.kraken.tentacles {
		status.Tentacles[name] = s.getTentacleStatus(name)
	}
	return status
}

func (s *Server) getServerStatistics() *ServerStatistics {
	stats := newServerStatistics()
	for name := range s.kraken.tentacles {
		stats.Tentacles[name] = s.getTentacleStatistics(name)
	}
	return stats
}

func (s *Server) getTentacleStatus(name string) *TentacleStatus {
	tentacle, ok := s.kraken.tentacles[name]
	if ok {
		return tentacle.getStatus()
	}
	return nil
}

func (s *Server) getTentacleStatistics(name string) *TentacleStatistics {
	tentacle, ok := s.kraken.tentacles[name]
	if ok {
		return tentacle.GetStatistics()
	}
	return nil
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

/statistics

	GET: get the statistics of this kraken

/statistics/<name>

	GET: get the statistics of the given tentacle

/tentacle/<name>

	PUT / POST : create or overwrite a new tentacle with body {"bandwidth": <int>, "retry": <int>}
	PATCH      : patch the tentacle change it bandwidth and number of retries with body  {"bandwidth": <int>, "retry": <int>}
	GET        : get the status of an existing tentacle
	DELETE     : get rid of the tentacle


/tentacle/<name>/<preyId>

	PUT/POST   : let me catch some prey with body { "url" : <string>, "priority" : <int>, ["method" : <string>, "body" : <string>, "tags" : [<string>, ...] }

`
)

func (s *Server) help(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "text/plain;utf-8;")
	w.Write([]byte(help))
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	p := r.URL.Path
	//log.Println(r.Method, r.URL.Path)
	switch p {
	case "/status":
		s.jsonResponse(http.StatusOK, w, s.getServerStatus())
		return
	case "/statistics":
		s.jsonResponse(http.StatusOK, w, s.getServerStatistics())
		return
	default:
		if len(p) == 0 {
			s.help(w)
			return
		}
		parts := strings.Split(p[1:], "/")
		if strings.HasPrefix(p, "/statistics") {
			if len(parts) == 2 {
				s.jsonResponse(http.StatusOK, w, s.getTentacleStatistics(parts[1]))
				return
			}
		} else if strings.HasPrefix(p, "/tentacle") {

			if len(parts) == 2 {
				switch r.Method {
				case "PATCH":
					tentacleDef := &TentacleDefinition{}
					decodeBody(r, tentacleDef)
					err := s.kraken.SqueezeTentacle(parts[1], tentacleDef.Bandwidth, tentacleDef.Retry)
					if err != nil {
						s.jsonResponse(http.StatusNotFound, w, "unknown tentacle")
						return
					}
					s.jsonResponse(http.StatusOK, w, "tentacle was updated")
					return
				case "POST":
				case "PUT":
					tentacleDef := &TentacleDefinition{}
					decodeBody(r, tentacleDef)
					s.kraken.GrowTentacle(parts[1], tentacleDef.Bandwidth, tentacleDef.Retry)
					//log.Println("server tentacles", s.kraken.tentacles)
					s.jsonResponse(http.StatusCreated, w, "created a tentacle")
					return
				case "DELETE":
					s.kraken.CutOffTentacle(parts[1])
					s.jsonResponse(http.StatusOK, w, "tentacle deleted")
					return
				case "GET":
					s.jsonResponse(http.StatusOK, w, s.getTentacleStatus(parts[1]))
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
						Method:   preyDefinition.Method,
						Body:     preyDefinition.Body,
						Tags:     preyDefinition.Tags,
						Locks:    preyDefinition.Locks,
					}
					s.jsonResponse(http.StatusOK, w, s.kraken.Catch(parts[1], prey))
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
