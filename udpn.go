package udpn

import (
	"net"
	"net/url"
	"time"
	"bufio"
	"strings"
	"net/http"
	"bytes"
	"strconv"
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
	
	broadcastAddr, bytes := makeSearchRequest(st, mx)
	_, err = conn.WriteTo(bytes, broadcastAddr)
	if err != nil {
		return
	}
	
	res, err = readResponses(conn, mx)
	return
}

func ParseResponse(httpResponse string, responseAddr *net.UDPAddr) (res Response, err error) {
	request, err := http.NewRequest("M-SEARCH", "239.255.255.250:1900/*", strings.NewReader(""))
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

func makeSearchRequest(st string, mx time.Duration) (broadcastAddr *net.UDPAddr, searchBytes []byte) {
	// Needed to specify * as PATh in HTTP
	replaceMePlaceHolder := "/replacemewithstar"
	
	broadcastAddr, _ = net.ResolveUDPAddr("udp", "239.255.255.250:1900")
	request, _ := http.NewRequest("M-SEARCH", 
		"http://" + broadcastAddr.String() + replaceMePlaceHolder, strings.NewReader(""))
	
	headers := request.Header
	headers.Set("User-Agent", "")
	headers.Set("st", st)
	headers.Set("man", `"ssdp:discover"`)
	headers.Set("mx", strconv.FormatInt(int64(mx / time.Second), 10))
	
	searchBytes = make([]byte, 0, 1024)
	buffer := bytes.NewBuffer(searchBytes)
	err := request.Write(buffer)
	if err != nil {
		panic("Fatal error writing to buffer. This should never happen (in theory).")
	}
	searchBytes = buffer.Bytes()
	
	// Path should be * unescape. This is a hack to accomplish that.
	searchBytes = bytes.Replace(searchBytes, []byte(replaceMePlaceHolder), []byte("*"), 1)
	
	return 
}
