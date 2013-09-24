package ssdp

import (
	"bytes"
	"net"
	"net/url"
	"strings"
	"testing"
	"time"
)

type timeoutError struct {
}

func (e *timeoutError) Error() string {
	return "i/o timeout"
}

func (e *timeoutError) Timeout() bool {
	return true
}

func (e *timeoutError) Temporary() bool {
	return true
}

type stubSearchReader struct {
	readIteration      int
	readBuffers        []string
	readAddr           *net.UDPAddr
	storedReadDeadline time.Time
}

func (s *stubSearchReader) ReadFromUDP(b []byte) (n int, addr *net.UDPAddr, err error) {
	// Return timeout error if no more buffers
	if s.readIteration >= len(s.readBuffers) {
		return 0, nil, &timeoutError{}
	}

	// Create and return the current buffer for the iteration
	buffer := bytes.NewBufferString(s.readBuffers[s.readIteration])
	n, err = buffer.Read(b)
	addr = s.readAddr

	s.readIteration += 1
	return
}

func (s *stubSearchReader) SetReadDeadline(t time.Time) error {
	s.storedReadDeadline = t
	return nil
}

func Test_listenForSearchResponses(t *testing.T) {
	conn, err := listenForSearchResponses()
	if err != nil {
		t.Fatal("Error listening for response.", err)
	}
	if conn == nil {
		t.Error("Connection is nil.")
	}
	defer conn.Close()
}

func Test_buildSearchRequest(t *testing.T) {
	expectedSearchRequest := "M-SEARCH * HTTP/1.1\r\n" +
		"Host: 239.255.255.250:1900\r\n" +
		"Man: \"ssdp:discover\"\r\n" +
		"Mx: 5\r\n" +
		"St: upnp:rootdevice\r\n" +
		"\r\n"
	expectedBroadcastAddr, _ := net.ResolveUDPAddr("udp", "239.255.255.250:1900")

	searchBytes, broadcastAddr := buildSearchRequest("upnp:rootdevice", 5*time.Second)

	actualSearchRequest := string(searchBytes)
	if expectedSearchRequest != actualSearchRequest {
		t.Errorf("Expected search request to be:\n\n%s\n but it was:\n\n%s\n",
			expectedSearchRequest, actualSearchRequest)
	}

	assertEqual(t, expectedBroadcastAddr.String(), broadcastAddr.String(), "broadcastAddr")
}

func Test_readSearchResponses(t *testing.T) {
	stub := &stubSearchReader{}
	stub.readAddr, _ = net.ResolveUDPAddr("udp", "192.168.0.12:1000")

	stub.readBuffers = make([]string, 0, 2)
	stub.readBuffers = append(stub.readBuffers, "HTTP/1.1 200 OK\r\n\r\n")
	stub.readBuffers = append(stub.readBuffers, "HTTP/1.1 200 OK\r\n\r\n")

	responses, err := readSearchResponses(stub, 1*time.Second)
	if err != nil {
		t.Fatal("Error while retrieving reading search responses.", err)
	}

	if len(stub.readBuffers) != len(responses) {
		t.Fatalf("Expected %d responses but received %d.", len(stub.readBuffers),
			len(responses))
	}
}

func Test_parseSearchResponse(t *testing.T) {
	responseBody := "HTTP/1.1 200 OK\r\n" +
		"CACHE-CONTROL: max-age=100\r\n" +
		"EXT:\r\n" +
		"LOCATION: http://10.1.2.3:80/description.xml\r\n" +
		"SERVER: FreeRTOS/6.0.5, UPnP/1.0, IpBridge/0.1\r\n" +
		"ST: upnp:rootdevice\r\n" +
		"USN: uuid:2f402f80-da50-11e1-9b23-0017880a4c69::upnp:rootdevice\r\n" +
		"Date: Sun, 18 Aug 2013 08:49:37 GMT\r\n" +
		"\r\n"

	responseAddr, _ := net.ResolveUDPAddr("udp", "10.1.2.3:1900")
	response, err := parseSearchResponse(strings.NewReader(responseBody), responseAddr)
	if err != nil {
		t.Fatal("Error while parsing the response.", err)
	}

	assertEqual(t, "max-age=100", response.Control, "response.Control")
	assertEqual(t, "FreeRTOS/6.0.5, UPnP/1.0, IpBridge/0.1", response.Server, "response.Server")
	assertEqual(t, "upnp:rootdevice", response.ST, "response.ST")
	assertEqual(t, "", response.Ext, "response.Ext")
	assertEqual(t, "uuid:2f402f80-da50-11e1-9b23-0017880a4c69::upnp:rootdevice", response.USN, "response.USN")
	assertEqual(t, responseAddr, response.ResponseAddr, "response.Addr")

	url, _ := url.Parse("http://10.1.2.3:80/description.xml")
	if url.String() != response.Location.String() {
		t.Errorf("%q is not equal to %q. %q", url.String(), response.Location.String(), "response.Location")
	}

	gmt, _ := time.LoadLocation("UTC")
	date := time.Date(2013, time.August, 18, 8, 49, 37, 0, gmt)
	assertEqual(t, date, response.Date, "response.Date")
}

func Test_parseSearchResponse_NoDateOrLocation(t *testing.T) {
	responseBody := "HTTP/1.1 200 OK\r\n\r\n"
	responseAddr, _ := net.ResolveUDPAddr("udp", "10.1.2.3:1900")
	response, err := parseSearchResponse(strings.NewReader(responseBody), responseAddr)
	if err != nil {
		t.Fatal("Error while parsing the response.", err)
	}

	emptyTime := time.Time{}
	if response.Date != emptyTime {
		t.Error("Date should be nil")
	}

	if response.Location != nil {
		t.Error("Location should be nil")
	}
}

func assertEqual(t *testing.T, expected interface{}, actual interface{}, errorMessage string) {
	if expected != actual {
		t.Errorf("%q is not equal to %q. %q", expected, actual, errorMessage)
	}
}
