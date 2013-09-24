package ssdp

import (
	"encoding/xml"
	"io"
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

func decodeDescription(reader io.Reader) (*DeviceDescription, error) {
	decoder := xml.NewDecoder(reader)

	deviceDescription := &DeviceDescription{}
	err := decoder.Decode(deviceDescription)
	if err != nil {
		return nil, err
	}

	return deviceDescription, nil
}
