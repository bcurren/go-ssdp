package udpn

import (
	"bufio"
	"bytes"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const (
	Port        = 1900
	BroadcastIP = "239.255.255.250"
)

type SearchResponse struct {
	Control  string
	Server   string
	ST       string
	Ext      string
	USN      string
	Location *url.URL
	Date     time.Time
	Addr     *net.UDPAddr
}

type searchReader interface {
	ReadFromUDP(b []byte) (n int, addr *net.UDPAddr, err error)
	SetReadDeadline(t time.Time) error
}

func Search(st string, mx time.Duration) (responses []SearchResponse, err error) {
	conn, err := listenForSearchResponses()
	if conn != nil {
		defer conn.Close()
	}
	if err != nil {
		return
	}

	searchBytes, broadcastAddr := buildSearchRequest(st, mx)
	// Write search bytes on the wire so all devices can respond
	_, err = conn.WriteTo(searchBytes, broadcastAddr)
	if err != nil {
		return
	}

	responses, err = readSearchResponses(conn, mx)
	return
}

func listenForSearchResponses() (conn *net.UDPConn, err error) {
	serverAddr, _ := net.ResolveUDPAddr("udp", "0.0.0.0:"+strconv.Itoa(Port))
	conn, err = net.ListenUDP("udp", serverAddr)
	return
}

func buildSearchRequest(st string, mx time.Duration) (searchBytes []byte, broadcastAddr *net.UDPAddr) {
	// Placeholder to replace with * later on
	replaceMePlaceHolder := "/replacemewithstar"

	broadcastAddr, _ = net.ResolveUDPAddr("udp", BroadcastIP+":"+strconv.Itoa(Port))
	request, _ := http.NewRequest("M-SEARCH",
		"http://"+broadcastAddr.String()+replaceMePlaceHolder, strings.NewReader(""))

	headers := request.Header
	headers.Set("User-Agent", "")
	headers.Set("st", st)
	headers.Set("man", `"ssdp:discover"`)
	headers.Set("mx", strconv.Itoa(int(mx/time.Second)))

	searchBytes = make([]byte, 0, 1024)
	buffer := bytes.NewBuffer(searchBytes)
	err := request.Write(buffer)
	if err != nil {
		panic("Fatal error writing to buffer. This should never happen (in theory).")
	}
	searchBytes = buffer.Bytes()

	// Replace placeholder with *. Needed because request always escapes * when it shouldn't
	searchBytes = bytes.Replace(searchBytes, []byte(replaceMePlaceHolder), []byte("*"), 1)

	return
}

func readSearchResponses(reader searchReader, duration time.Duration) (responses []SearchResponse, err error) {
	responses = make([]SearchResponse, 0, 10)
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

		response, err := parseSearchResponse(string(buf[:rlen]), addr)
		if err != nil {
			return nil, err
		}
		responses = append(responses, response)
	}

	return
}

func parseSearchResponse(httpResponse string, responseAddr *net.UDPAddr) (res SearchResponse, err error) {
	reader := bufio.NewReader(strings.NewReader(httpResponse))
	request := &http.Request{} // Needed for ReadResponse but doesn't have to be real
	response, err := http.ReadResponse(reader, request)
	if err != nil {
		return
	}
	headers := response.Header

	res = SearchResponse{}

	res.Control = headers.Get("cache-control")
	res.Server = headers.Get("server")
	res.ST = headers.Get("st")
	res.Ext = headers.Get("ext")
	res.USN = headers.Get("usn")
	res.Addr = responseAddr

	if headers.Get("location") != "" {
		res.Location, err = response.Location()
		if err != nil {
			return
		}
	}

	date := headers.Get("date")
	if date != "" {
		res.Date, err = http.ParseTime(date)
		if err != nil {
			return
		}
	}

	return
}


