package main

import (
	"net"
	"strings"
	"time"
)

const NumFields int = 6
const DefaultTimeout int = 5 // default TCP timeout in seconds

type Minestat struct {
	Address        string
	Port           string
	Online         bool
	Version        string
	Motd           string
	CurrentPlayers string
	MaxPlayers     string
	Latency        time.Duration
}


func GetMinestat(givenAddress string, givenPort string, optionalTimeout ...int) *Minestat {
	var m Minestat

	timeout := DefaultTimeout
	if len(optionalTimeout) > 0 {
		timeout = optionalTimeout[0]
	}
	m.Address = givenAddress
	m.Port = givenPort
	/* Latency may report a misleading value of >1s due to name resolution delay when using net.Dial().
	   A workaround for this issue is to use an IP address instead of a hostname or FQDN. */
	start_time := time.Now()
	conn, err := net.DialTimeout("tcp", m.Address + ":" + m.Port, time.Duration(timeout) * time.Second)
	m.Latency = time.Since(start_time)
	m.Latency = m.Latency.Round(time.Millisecond)
	if err != nil {
		m.Online = false
		return &m
	}

	_, err = conn.Write([]byte("\xFE\x01"))
	if err != nil {
		m.Online = false
		return &m
	}

	rawData := make([]byte, 512)
	_, err = conn.Read(rawData)
	if err != nil {
		m.Online = false
		return &m
	}
	_ = conn.Close()

	if len(rawData) == 0 {
		m.Online = false
		return &m
	}

	data := strings.Split(string(rawData[:]), "\x00\x00\x00")
	if data != nil && len(data) >= NumFields {
		m.Online = true
		m.Version = data[2]
		m.Motd = data[3]
		m.CurrentPlayers = data[4]
		m.MaxPlayers = data[5]
	} else {
		m.Online = false
	}

	return &m
}
