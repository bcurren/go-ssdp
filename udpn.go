// SSDP (Simple Service Discovery Protocol) package provides an implementation of the SSDP
// specification.
package ssdp

import (
	"bufio"
	"bytes"
	"io"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const (
	// The port for SSDP discovery
	Port = 1900

	// The IP for SSDP broadcast
	BroadcastIP = "239.255.255.250"
)

// The search response from a device implementing SSDP.
type SearchResponse struct {
	Control      string
	Server       string
	ST           string
	Ext          string
	USN          string
	Location     *url.URL
	Date         time.Time
	ResponseAddr *net.UDPAddr
}

// The search reader interface to read UDP packets on the wire with a timeout
// period specified.
type searchReader interface {
	ReadFromUDP(b []byte) (n int, addr *net.UDPAddr, err error)
	SetReadDeadline(t time.Time) error
}

// Search the network for SSDP devices using the given search string and duration
// to discover new devices. This function will return an array of SearchReponses
// discovered.
func Search(st string, mx time.Duration) ([]SearchResponse, error) {
	conn, err := listenForSearchResponses()
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	searchBytes, broadcastAddr := buildSearchRequest(st, mx)
	// Write search bytes on the wire so all devices can respond
	_, err = conn.WriteTo(searchBytes, broadcastAddr)
	if err != nil {
		return nil, err
	}

	return readSearchResponses(conn, mx)
}

func listenForSearchResponses() (*net.UDPConn, error) {
	serverAddr, _ := net.ResolveUDPAddr("udp", "0.0.0.0:"+strconv.Itoa(Port))
	return net.ListenUDP("udp", serverAddr)
}

func buildSearchRequest(st string, mx time.Duration) ([]byte, *net.UDPAddr) {
	// Placeholder to replace with * later on
	replaceMePlaceHolder := "/replacemewithstar"

	broadcastAddr, _ := net.ResolveUDPAddr("udp", BroadcastIP+":"+strconv.Itoa(Port))
	request, _ := http.NewRequest("M-SEARCH",
		"http://"+broadcastAddr.String()+replaceMePlaceHolder, strings.NewReader(""))

	headers := request.Header
	headers.Set("User-Agent", "")
	headers.Set("st", st)
	headers.Set("man", `"ssdp:discover"`)
	headers.Set("mx", strconv.Itoa(int(mx/time.Second)))

	searchBytes := make([]byte, 0, 1024)
	buffer := bytes.NewBuffer(searchBytes)
	err := request.Write(buffer)
	if err != nil {
		panic("Fatal error writing to buffer. This should never happen (in theory).")
	}
	searchBytes = buffer.Bytes()

	// Replace placeholder with *. Needed because request always escapes * when it shouldn't
	searchBytes = bytes.Replace(searchBytes, []byte(replaceMePlaceHolder), []byte("*"), 1)

	return searchBytes, broadcastAddr
}

func readSearchResponses(reader searchReader, duration time.Duration) ([]SearchResponse, error) {
	responses := make([]SearchResponse, 0, 10)
	// Only listen for responses for duration amount of time.
	reader.SetReadDeadline(time.Now().Add(duration))

	buf := make([]byte, 1024)
	for {
		rlen, addr, err := reader.ReadFromUDP(buf)
		if nerr, ok := err.(net.Error); ok && nerr.Timeout() {
			break // duration reached, return what we've found
		}
		if err != nil {
			return nil, err
		}

		response, err := parseSearchResponse(bytes.NewReader(buf[:rlen]), addr)
		if err != nil {
			return nil, err
		}
		responses = append(responses, *response)
	}

	return responses, nil
}

func parseSearchResponse(httpResponse io.Reader, responseAddr *net.UDPAddr) (*SearchResponse, error) {
	reader := bufio.NewReader(httpResponse)
	request := &http.Request{} // Needed for ReadResponse but doesn't have to be real
	response, err := http.ReadResponse(reader, request)
	if err != nil {
		return nil, err
	}
	headers := response.Header

	res := &SearchResponse{}

	res.Control = headers.Get("cache-control")
	res.Server = headers.Get("server")
	res.ST = headers.Get("st")
	res.Ext = headers.Get("ext")
	res.USN = headers.Get("usn")
	res.ResponseAddr = responseAddr

	if headers.Get("location") != "" {
		res.Location, err = response.Location()
		if err != nil {
			return nil, err
		}
	}

	date := headers.Get("date")
	if date != "" {
		res.Date, err = http.ParseTime(date)
		if err != nil {
			return nil, err
		}
	}

	return res, nil
}
