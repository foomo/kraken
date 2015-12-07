package kraken

import (
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
)

func panicOnErr(err error) {
	if err != nil {
		panic(err)
	}
}

func mockServer(mockHandler func(w http.ResponseWriter, r *http.Request)) *httptest.Server {
	ts := httptest.NewServer(http.HandlerFunc(mockHandler))
	defer ts.Close()
	return ts
}

func mockKraken() (ks *Server, ts *httptest.Server) {
	ks = NewServer(NewKraken())
	ts = httptest.NewServer(ks)
	return ks, ts
}

func TestTentacleCreation(t *testing.T) {
	ks, ts := mockKraken()
	defer ts.Close()
	c := NewClient(ts.URL)
	tentacleName := "sepp"
	bandwidth := 3
	retry := 3
	err := c.CreateTentacle(tentacleName, bandwidth, retry)
	panicOnErr(err)
	tentacle, ok := ks.kraken.tentacles[tentacleName]
	if ok != true {
		t.Fatal("tentacle is missing")
	}
	if retry != tentacle.Retry {
		t.Fatal("retry fail", retry, tentacle.Retry)
	}
	if bandwidth != tentacle.Bandwidth {
		t.Fatal("bandwidth fail")
	}
}

func TestTentacleDeletion(t *testing.T) {
	ks, ts := mockKraken()
	defer ts.Close()
	c := NewClient(ts.URL)
	tentacleName := "sepp"
	bandwidth := 3
	retry := 3
	err := c.CreateTentacle(tentacleName, bandwidth, retry)
	panicOnErr(err)
	tentacle, ok := ks.kraken.tentacles[tentacleName]
	if ok != true {
		t.Fatal("tentacle is missing")
	}
	err = c.DeleteTentacle(tentacleName)
	panicOnErr(err)
	tentacle, ok = ks.kraken.tentacles[tentacleName]
	if ok {
		t.Fatal("tentacle should be missing", tentacle)
	}
	err = c.DeleteTentacle(tentacleName)
	if err != nil {
		t.Fatal("deleting a non existent tentacle should result in an error")
	}
}

func TestTentaclePatch(t *testing.T) {
	ks, ts := mockKraken()
	defer ts.Close()
	c := NewClient(ts.URL)
	tentacleName := "sepp"
	bandwidth := 3
	retry := 3
	panicOnErr(c.CreateTentacle(tentacleName, bandwidth, retry))
	panicOnErr(c.PatchTentacle(tentacleName, bandwidth+1, retry+1))

	tentacle, ok := ks.kraken.tentacles[tentacleName]
	if ok != true {
		t.Fatal("tentacle is missing")
	}
	if retry+1 != tentacle.Retry {
		t.Fatal("retry fail", retry)
	}
	if bandwidth+1 != tentacle.Bandwidth {
		t.Fatal("bandwidth fail")
	}
}

func createTentacle(tentacleName string, ts *httptest.Server) *Client {
	c := NewClient(ts.URL)
	bandwidth := 3
	retry := 3
	panicOnErr(c.CreateTentacle(tentacleName, bandwidth, retry))
	return c
}

func TestPrey(t *testing.T) {
	methods := []string{}
	waitChan := make(chan bool)
	ms := mockServer(func(w http.ResponseWriter, r *http.Request) {
		methods = append(methods, r.Method)
		waitChan <- true
	})
	defer ms.Close()
	ks, ts := mockKraken()
	defer ts.Close()
	tentacleName := "sepp"
	c := createTentacle(tentacleName, ts)
	preyID := "a"
	preyMethod := "TEST"
	go func() {
		<-waitChan
		if methods[1] != preyMethod {
			log.Println("method did not hit the server")
		}
	}()
	tags := []string{"test"}
	panicOnErr(c.AddPrey(tentacleName, preyID, ms.URL+"/foo", preyMethod, []byte{}, tags))
	tentacleStatus, err := c.GetTentacle(tentacleName)
	panicOnErr(err)
	if tentacleStatus == nil {
		t.Fatal("nil tentacle status")
	}
	if tentacleStatus.Name != tentacleName {
		t.Fatal("name mismatch")
	}
	preyA, ok := ks.kraken.tentacles[tentacleName].Prey[preyID]
	if ok == false {
		log.Fatal("no that is not ok i want my prey", preyID)
	}
	if preyA.Method != preyMethod {
		t.Fatal("that is not the method i was looking for")
	}
	if preyA.Tags[0] != tags[0] {
		t.Fatal("tag mismatch")
	}
}

func TestBody(t *testing.T) {
	bodyChan := make(chan string)
	ms := mockServer(func(w http.ResponseWriter, r *http.Request) {
		body, _ := ioutil.ReadAll(r.Body)
		bodyChan <- string(body)
	})
	defer ms.Close()
	ks, ts := mockKraken()
	defer ts.Close()
	tentacleName := "sepp"
	c := createTentacle(tentacleName, ts)
	preyID := "a"
	preyMethod := "TEST"
	testBody := "a test body"
	go func() {
		if testBody != <-bodyChan {
			t.Fatal("body fail")
		}
	}()
	panicOnErr(c.AddPrey(tentacleName, preyID, ms.URL+"/foo", preyMethod, []byte(testBody), []string{}))
	tentacleStatus, err := c.GetTentacle(tentacleName)
	panicOnErr(err)
	if tentacleStatus == nil {
		t.Fatal("nil tentacles tatus")
	}
	if tentacleStatus.Name != tentacleName {
		t.Fatal("name mismatch")
	}
	preyA, ok := ks.kraken.tentacles[tentacleName].Prey[preyID]
	if ok == false {
		log.Fatal("no that is not ok i want my prey", preyID)
	}
	if string(preyA.Body) != testBody {
		t.Fatal("body missmatch", testBody, preyA.Body)
	}

}

func TestGetServerStatus(t *testing.T) {
	bodyChan := make(chan string)
	ms := mockServer(func(w http.ResponseWriter, r *http.Request) {
		body, _ := ioutil.ReadAll(r.Body)
		bodyChan <- string(body)
	})
	defer ms.Close()
	_, ts := mockKraken()
	defer ts.Close()
	tentacleName := "sepp"
	c := createTentacle(tentacleName, ts)
	preyID := "a"
	preyMethod := "TEST"
	testBody := "a test body"
	go func() {
		if testBody != <-bodyChan {
			t.Fatal("body fail")
		}
	}()
	panicOnErr(c.AddPrey(tentacleName, preyID, ms.URL+"/foo", preyMethod, []byte(testBody), []string{}))
	serverStatus, err := c.GetServerStatus()
	panicOnErr(err)
	tentacleStatus, ok := serverStatus.Tentacles[tentacleName]
	if !ok {
		t.Fatal("missing tentacle in status")
	}
	_, ok = tentacleStatus.Prey[preyID]
	if !ok {
		t.Fatal("missing test prey in server status")
	}

}
