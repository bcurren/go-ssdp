package udpn

import (
	"testing"
	"net"
	"net/url"
	"time"
)

func Test_listenForResponse(t *testing.T) {
	conn, err := listenForResponse()
	if err != nil {
		t.Fatal("Error listening for response.", err)
	}
	if conn == nil {
		t.Error("Connection is nil.")
	}
	defer conn.Close()
}

func Test_ParseResponse(t *testing.T) {
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
	response, err := ParseResponse(responseBody, responseAddr)
	if err != nil {
		t.Fatal("Error while parsing the response.", err)
	}
	
	assertEqual(t, "max-age=100", response.Control, "response.Control")
	assertEqual(t, "FreeRTOS/6.0.5, UPnP/1.0, IpBridge/0.1", response.Server, "response.Server")
	assertEqual(t, "upnp:rootdevice", response.ST, "response.ST")
	assertEqual(t, "", response.Ext, "response.Ext")
	assertEqual(t, "uuid:2f402f80-da50-11e1-9b23-0017880a4c69::upnp:rootdevice", response.USN, "response.USN")
	assertEqual(t, responseAddr, response.Addr, "response.Addr")
	
	url, _ := url.Parse("http://10.1.2.3:80/description.xml")
	if url.String() != response.Location.String() {
		t.Errorf("%q is not equal to %q. %q", url.String(), response.Location.String(), "response.Location")
	}
	
	gmt, _ := time.LoadLocation("UTC")
	date := time.Date(2013, time.August, 18, 8, 49, 37, 0, gmt)
	assertEqual(t, date, response.Date, "response.Date")
}

func Test_ParseResponse_NoDateOrLocation(t *testing.T) {
	responseBody := "HTTP/1.1 200 OK\r\n" +
		"\r\n"
		
	responseAddr, _ := net.ResolveUDPAddr("udp", "10.1.2.3:1900")
	response, err := ParseResponse(responseBody, responseAddr)
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
