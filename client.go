package kraken

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
)

// Client - a client to talk to kraken
type Client struct {
	serverAddress string
}

// NewClient makes a new client
func NewClient(serverAddress string) *Client {
	return &Client{
		serverAddress: serverAddress,
	}
}

// Get http get
func (c *Client) Get(uri string) (response *http.Response, err error) {
	return c.Do("GET", uri, nil)
}

// Do a request with data
func (c *Client) Do(method string, uri string, data interface{}) (response *http.Response, err error) {
	body := []byte{}
	if data != nil {
		body, err = json.Marshal(data)
		if err != nil {
			return nil, err
		}
	}
	request, err := http.NewRequest(method, c.serverAddress+uri, bytes.NewReader(body))
	return http.DefaultClient.Do(request)
}

// CreateTentacle - create a tentacle
func (c *Client) CreateTentacle(name string, bandwidth int, retry int) error {
	return c.tentacleAction("PUT", name, bandwidth, retry)
}

// PatchTentacle - create a tentacle
func (c *Client) PatchTentacle(name string, bandwidth int, retry int) error {
	return c.tentacleAction("PATCH", name, bandwidth, retry)
}

func (c *Client) AddPrey(tentacleName string, preyId string, urlString string, method string, body []byte) error {
	prey := &PreyDefinition{
		Method:   method,
		URL:      urlString,
		Priority: 1,
		Body:     body,
	}
	_, err := c.Do("PUT", "/tentacle/"+url.QueryEscape(tentacleName)+"/"+url.QueryEscape(preyId), prey)
	return err
}

func (c *Client) GetTentacle(tentacleName string) (tentacleStatus *TentacleStatus, err error) {
	response, err := c.Do("GET", "/tentacle/"+url.QueryEscape(tentacleName), nil)
	responseBytes, readErr := ioutil.ReadAll(response.Body)
	//log.Println("response bytes", string(responseBytes))
	if readErr != nil {
		response.Body.Close()
	}
	tentacleStatus = &TentacleStatus{}
	if err != nil {
		tentacleStatus = nil
	} else {
		err = json.Unmarshal(responseBytes, &tentacleStatus)
	}
	return tentacleStatus, err
}

func (c *Client) tentacleAction(action string, name string, bandwidth int, retry int) error {
	_, err := c.Do(action, "/tentacle/"+url.QueryEscape(name), &TentacleDefinition{
		Bandwidth: bandwidth,
		Retry:     retry,
	})
	return err
}
