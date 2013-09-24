package ssdp

import (
	"encoding/xml"
	"io"
	"net/http"
	"net/url"
	"time"
)

type DeviceDescription struct {
	SpecVersion      SpecVersion `xml:"specVersion"`
	UrlBase          string      `xml:"URLBase"`
	DeviceType       string      `xml:"device>deviceType"`
	FriendlyName     string      `xml:"device>friendlyName"`
	Manufacturer     string      `xml:"device>manufacturer"`
	ManufacturerUrl  string      `xml:"device>manufacturerURL"`
	ModelDescription string      `xml:"device>modelDescription"`
	ModelName        string      `xml:"device>modelName"`
	ModelNumber      string      `xml:"device>modelNumber"`
	ModelUrl         string      `xml:"device>modelURL"`
	SerialNumber     string      `xml:"device>serialNumber"`
	Udn              string      `xml:"device>UDN"`
	Upc              string      `xml:"device>UPC"`
	PresentationUrl  string      `xml:"device>presentationURL"`
	Icons            []Icon      `xml:"device>iconList>icon"`
}

type SpecVersion struct {
	Major int `xml:"major"`
	Minor int `xml:"minor"`
}

type Icon struct {
	MimeType string `xml:"mimetype"`
	Width    int    `xml:"width"`
	Height   int    `xml:"height"`
	Depth    int    `xml:"depth"`
	Url      string `xml:"url"`
}

func SearchForDevices(st string, mx time.Duration) ([]DeviceDescription, error) {
	// Search for devices
	responses, err := Search(st, mx)
	if err != nil {
		return nil, err
	}

	// Reduce to unique locations
	locations := reduceOnLocation(responses)

	// Collect device description for each location
	return collectDeviceDescriptions(locations)
}

func collectDeviceDescriptions(locations []url.URL) ([]DeviceDescription, error) {
	deviceDescriptions := make([]DeviceDescription, 0, len(locations))
	for _, location := range locations {
		deviceDescription, err := getDescriptionXml(location)
		if err != nil {
			return nil, err
		}
		deviceDescriptions = append(deviceDescriptions, *deviceDescription)
	}

	return deviceDescriptions, nil
}

func getDescriptionXml(url url.URL) (*DeviceDescription, error) {
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

func decodeDescription(reader io.Reader) (*DeviceDescription, error) {
	decoder := xml.NewDecoder(reader)

	deviceDescription := &DeviceDescription{}
	err := decoder.Decode(deviceDescription)
	if err != nil {
		return nil, err
	}

	return deviceDescription, nil
}
