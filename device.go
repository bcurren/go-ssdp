package ssdp

import (
	"encoding/xml"
	"io"
	"net/http"
	"net/url"
	"time"
)

type Device struct {
	SpecVersion      SpecVersion `xml:"specVersion"`
	URLBase          string      `xml:"URLBase"`
	DeviceType       string      `xml:"device>deviceType"`
	FriendlyName     string      `xml:"device>friendlyName"`
	Manufacturer     string      `xml:"device>manufacturer"`
	ManufacturerURL  string      `xml:"device>manufacturerURL"`
	ModelDescription string      `xml:"device>modelDescription"`
	ModelName        string      `xml:"device>modelName"`
	ModelNumber      string      `xml:"device>modelNumber"`
	ModelURL         string      `xml:"device>modelURL"`
	SerialNumber     string      `xml:"device>serialNumber"`
	UDN              string      `xml:"device>UDN"`
	UPC              string      `xml:"device>UPC"`
	PresentationURL  string      `xml:"device>presentationURL"`
	Icons            []Icon      `xml:"device>iconList>icon"`
}

type SpecVersion struct {
	Major int `xml:"major"`
	Minor int `xml:"minor"`
}

type Icon struct {
	MIMEType string `xml:"mimetype"`
	Width    int    `xml:"width"`
	Height   int    `xml:"height"`
	Depth    int    `xml:"depth"`
	URL      string `xml:"url"`
}

func SearchForDevices(st string, mx time.Duration) ([]Device, error) {
	responses, err := Search(st, mx)
	if err != nil {
		return nil, err
	}

	locations := reduceOnLocation(responses)

	return collectDevices(locations)
}

func collectDevices(locations []url.URL) ([]Device, error) {
	devices := make([]Device, 0, len(locations))
	for _, location := range locations {
		device, err := getDescriptionXml(location)
		if err != nil {
			return nil, err
		}
		devices = append(devices, *device)
	}

	return devices, nil
}

func getDescriptionXml(url url.URL) (*Device, error) {
	response, err := http.Get(url.String())
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	return decodeDescription(response.Body)
}

func reduceOnLocation(responses []SearchResponse) []url.URL {
	uniqueLocations := make(map[url.URL]bool)

	for _, response := range responses {
		uniqueLocations[*response.Location] = true
	}

	locations := make([]url.URL, 0, len(uniqueLocations))
	for location, _ := range uniqueLocations {
		locations = append(locations, location)
	}

	return locations
}

func decodeDescription(reader io.Reader) (*Device, error) {
	decoder := xml.NewDecoder(reader)

	device := &Device{}
	err := decoder.Decode(device)
	if err != nil {
		return nil, err
	}

	return device, nil
}
