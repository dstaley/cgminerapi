/*
Package cgminerapi provides a client for using the cgminer API.

Construct a new cgminer client, then use the various services on the client to
access different parts of the cgminer RPC API. For example:

	client := cgminerapi.NewCgminerAPI("localhost", "4028")

	command := cgminerapi.APICommand{Method: "summary"}
	resp, err := api.Send(&command)

Set optional parameters for an method using an APICommand's Parameter field.

	command := cgminerapi.APICommand{Method: "gpu", Parameter: "0"}
	resp, err := api.Send(&command)
*/
package cgminerapi

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"log"
	"net"
)

// APIClient stores connection details.
type APIClient struct {
	Host, Port string
}

type APIStatus struct {
	Code        int    `json:"Code,omitempty"`
	Description string `json:"Description,omitempty"`
	Msg         string `json:"Msg,omitempty"`
	STATUS      string `json:"STATUS,omitempty"`
	When        int    `json:"When,omitempty"`
}

// Response holds the various possible API response fields.
type Response struct {
	Status  []APIStatus `json:"STATUS"`
	Summary []Summary   `json:"SUMMARY,omitempty"`
	Config  []Config    `json:"CONFIG,omitempty"`
	Devs    []Devs      `json:"DEVS,omitempty"`
	Gpu     []Devs      `json:"GPU,omitempty"`
}

type Summary struct {
	Accepted           int     `json:"Accepted,omitempty"`
	BestShare          float64 `json:"Best Share,omitempty"`
	DeviceHardware     float64 `json:"Device Hardware%,omitempty"`
	DeviceRejected     float64 `json:"Device Rejected%,omitempty"`
	DifficultyAccepted float64 `json:"Difficulty Accepted,omitempty"`
	DifficultyRejected float64 `json:"Difficulty Rejected,omitempty"`
	DifficultyStale    float64 `json:"Difficulty Stale,omitempty"`
	Discarded          float64 `json:"Discarded,omitempty"`
	Elapsed            float64 `json:"Elapsed,omitempty"`
	FoundBlocks        float64 `json:"Found Blocks,omitempty"`
	GetFailures        float64 `json:"Get Failures,omitempty"`
	Getworks           float64 `json:"Getworks,omitempty"`
	HardwareErrors     float64 `json:"Hardware Errors,omitempty"`
	LocalWork          float64 `json:"Local Work,omitempty"`
	MHS5s              float64 `json:"MHS 5s,omitempty"`
	MHSav              float64 `json:"MHS av,omitempty"`
	NetworkBlocks      float64 `json:"Network Blocks,omitempty"`
	PoolRejected       float64 `json:"Pool Rejected%,omitempty"`
	PoolStale          float64 `json:"Pool Stale%,omitempty"`
	Rejected           float64 `json:"Rejected,omitempty"`
	RemoteFailures     float64 `json:"Remote Failures,omitempty"`
	Stale              float64 `json:"Stale,omitempty"`
	TotalMH            float64 `json:"Total MH,omitempty"`
	Utility            float64 `json:"Utility,omitempty"`
	WorkUtility        float64 `json:"Work Utility,omitempty"`
}

type Config struct {
	ADL          string  `json:"ADL,omitempty"`
	ADLinuse     string  `json:"ADL in use,omitempty"`
	ASCCount     float64 `json:"ASC Count,omitempty"`
	DeviceCode   string  `json:"Device Code,omitempty"`
	Expiry       float64 `json:"Expiry,omitempty"`
	FailoverOnly bool    `json:"Failover-Only,omitempty"`
	GPUCount     float64 `json:"GPU Count,omitempty"`
	Hotplug      float64 `json:"Hotplug,omitempty"`
	LogInterval  float64 `json:"Log Interval,omitempty"`
	OS           string  `json:"OS,omitempty"`
	PGACount     float64 `json:"PGA Count,omitempty"`
	PoolCount    float64 `json:"Pool Count,omitempty"`
	Queue        float64 `json:"Queue,omitempty"`
	ScanTime     float64 `json:"ScanTime,omitempty"`
	Strategy     string  `json:"Strategy,omitempty"`
}

type Devs struct {
	Accepted            int     `json:"Accepted,omitempty"`
	DeviceElapsed       float64 `json:"Device Elapsed,omitempty"`
	DeviceHardware      float64 `json:"Device Hardware%,omitempty"`
	DeviceRejected      float64 `json:"Device Rejected%,omitempty"`
	Diff1Work           float64 `json:"Diff1 Work,omitempty"`
	DifficultyAccepted  float64 `json:"Difficulty Accepted,omitempty"`
	DifficultyRejected  float64 `json:"Difficulty Rejected,omitempty"`
	Enabled             string  `json:"Enabled,omitempty"`
	FanPercent          float64 `json:"Fan Percent,omitempty"`
	FanSpeed            int     `json:"Fan Speed,omitempty"`
	GPU                 float64 `json:"GPU,omitempty"`
	GPUActivity         int     `json:"GPU Activity,omitempty"`
	GPUClock            int     `json:"GPU Clock,omitempty"`
	GPUVoltage          float64 `json:"GPU Voltage,omitempty"`
	HardwareErrors      float64 `json:"Hardware Errors,omitempty"`
	Intensity           string  `json:"Intensity,omitempty"`
	LastShareDifficulty float64 `json:"Last Share Difficulty,omitempty"`
	LastSharePool       float64 `json:"Last Share Pool,omitempty"`
	LastShareTime       float64 `json:"Last Share Time,omitempty"`
	LastValidWork       float64 `json:"Last Valid Work,omitempty"`
	MHS5s               float64 `json:"MHS 5s,omitempty"`
	MHSav               float64 `json:"MHS av,omitempty"`
	MemoryClock         int     `json:"Memory Clock,omitempty"`
	Powertune           int     `json:"Powertune,omitempty"`
	Rejected            int     `json:"Rejected,omitempty"`
	Status              string  `json:"Status,omitempty"`
	Temperature         float64 `json:"Temperature,omitempty"`
	TotalMH             float64 `json:"Total MH,omitempty"`
	Utility             float64 `json:"Utility,omitempty"`
}

type APIError struct {
	Code    int    `json:"code,omitempty"`
	Message string `json:"message,omitempty"`
}

// APICommand holds the API method and any parameters specified.
type APICommand struct {
	Method    string `json:"command"`
	Parameter string `json:"parameter,omitempty"`
}

// NewCgminerAPI returns a pointer to an APIClient with the specified host and port.
func NewCgminerAPI(host string, port string) *APIClient {
	return &APIClient{host, port}
}

func ReadAll(r io.Reader) (string, error) {
	b, err := ioutil.ReadAll(r)
	return string(bytes.Trim(b, " \x00")), err
}

func Encode(e APIError) string {
	blob, _ := json.Marshal(e)
	return string(blob)
}

// Send sends the APICommand (and any specified parameters) and returns a Response containing
// the response from the API.
func (client *APIClient) Send(command *APICommand) (Response, error) {
	c, err := net.Dial("tcp", client.Host+":"+client.Port)
	if err != nil {
		log.Fatal(err)
		return Response{}, err
	}
	defer c.Close()

	blob, err := json.Marshal(command)
	if err != nil {
		log.Fatal(err)
		return Response{}, err
	}

	_, err = c.Write(blob)

	if err != nil {
		log.Fatal(err)
		return Response{}, err
	}

	jsonstring, err := ReadAll(c)
	if err != nil {
		log.Fatal(err)
		return Response{}, err
	}

	var resp Response
	err = json.Unmarshal([]byte(jsonstring), &resp)
	if err != nil {
		log.Fatal(err)
		return Response{}, err
	}

	switch resp.Status[0].STATUS {
	case "W", "I", "S":
		return resp, nil
	case "E", "F":
		return Response{}, errors.New(resp.Status[0].Msg)
	}
	return Response{}, errors.New("Unknown error")
}
