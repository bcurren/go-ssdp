package udpn

import (
	"net"
	"net/url"
	"time"
	"bufio"
	"strings"
	"net/http"
)

type Response struct {
	Control  string
	Server   string
	ST       string
	Ext      string
	USN      string
	Location *url.URL
	Date     time.Time
	Addr     *net.UDPAddr
}

func Search(st string, mx time.Duration) (res []Response, err error) {
	conn, err := listenForResponse()
	if conn != nil {
		defer conn.Close()
	}
	if err != nil {
		return
	}
	
	err = makeSearchRequest(conn, st, mx)
	if err != nil {
		return
	}
	
	res, err = readResponses(conn, mx)
	return
}

func ParseResponse(httpResponse string, responseAddr *net.UDPAddr) (res Response, err error) {
	request, err := http.NewRequest("M-SEARCH", "239.255.255.250:1900", strings.NewReader(""))
	if err != nil {
		return
	}
	
	reader := bufio.NewReader(strings.NewReader(httpResponse))
	response, err := http.ReadResponse(reader, request)
	if err != nil {
		return
	}
	headers := response.Header
	
	res = Response{}
	
	res.Control = headers.Get("cache-control")
	res.Server = headers.Get("server")
	res.ST = headers.Get("st")
	res.Ext = headers.Get("ext")
	res.USN = headers.Get("usn")
	res.Addr = responseAddr
	
	location, err := response.Location()
	if err != nil {
		return
	}
	res.Location = location
	
	date := headers.Get("date")
	if date != "" {
		res.Date, err = http.ParseTime(date)
		if err != nil {
			return
		}
	}
	
	return
}

func readResponses(conn *net.UDPConn, duration time.Duration) (responses []Response, err error) {
	responses = make([]Response, 0, 10)
	conn.SetReadDeadline(time.Now().Add(duration))
	
	buf := make([]byte, 1024)
	for {
		rlen, addr, err := conn.ReadFromUDP(buf)
		if nerr, ok := err.(net.Error); ok && nerr.Timeout() {
			break // timeout reached, return what we've found
		}
		if err != nil {
			return nil, err
		}
		
		response, err := ParseResponse(string(buf[:rlen]), addr)
		if err != nil {
			return nil, err
		}
		responses = append(responses, response)
	}
	
	return
}

func listenForResponse() (conn *net.UDPConn, err error) {
	serverAddr, _ := net.ResolveUDPAddr("udp", "0.0.0.0:1900")
	conn, err = net.ListenUDP("udp", serverAddr)
	return
}

func makeSearchRequest(conn *net.UDPConn, st string, mx time.Duration) (err error) {	
	discoveryAddr, _ := net.ResolveUDPAddr("udp", "239.255.255.250:1900")
	requestBody := "M-SEARCH * HTTP/1.1\r\f" +
		"HOST:239.255.255.250:1900\r\f" +
		"ST:" + st + "\r\f" +
		"Man:\"ssdp:discover\"\r\f" +
		"MX:" + string(mx / time.Second) + "\r\f" +
		"\r\f"
	_, err = conn.WriteTo([]byte(requestBody), discoveryAddr)
	
	return
}
