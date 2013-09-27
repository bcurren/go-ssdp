package ssdp

import (
	"bytes"
	"io/ioutil"
	"net/url"
	"path/filepath"
	"testing"
)

func Test_reduceOnLocation(t *testing.T) {
	responses := make([]SearchResponse, 3, 3)
	responses[0].Location, _ = url.Parse("http://192.168.0.10:80/description.xml")
	responses[1].Location, _ = url.Parse("http://192.168.0.11:80/description.xml")
	responses[2].Location, _ = url.Parse("http://192.168.0.10:80/description.xml")

	locations := reduceOnLocation(responses)

	assertEqual(t, 2, len(locations), "len(locations)")
	assertEqual(t, *responses[0].Location, locations[0], "locations[0]")
	assertEqual(t, *responses[1].Location, locations[1], "locations[1]")
}

func Test_parseDecriptionXml(t *testing.T) {
	descriptionFile := filepath.Join(".", "test_responses", "hue_description.xml")
	fileBytes, err := ioutil.ReadFile(descriptionFile)
	if err != nil {
		t.Fatal("Error reading in stub description.xml.", err)
	}

	device, err := decodeDescription(bytes.NewReader(fileBytes))
	if err != nil {
		t.Fatal("Could not decode description.", err)
	}

	assertEqual(t, 1, device.SpecVersion.Major, "SpecVersion.Major")
	assertEqual(t, 0, device.SpecVersion.Minor, "SpecVersion.Minor")
	assertEqual(t, "http://192.168.0.21:80/", device.URLBase, "URLBase")
	assertEqual(t, "urn:schemas-upnp-org:device:Basic:1", device.DeviceType, "DeviceType")
	assertEqual(t, "Philips hue (192.168.0.21)", device.FriendlyName, "FriendlyName")
	assertEqual(t, "Royal Philips Electronics", device.Manufacturer, "Manufacturer")
	assertEqual(t, "http://www.philips.com", device.ManufacturerURL, "ManufacturerURL")
	assertEqual(t, "Philips hue Personal Wireless Lighting", device.ModelDescription, "ModelDescription")
	assertEqual(t, "Philips hue bridge 2012", device.ModelName, "ModelName")
	assertEqual(t, "1000000000000", device.ModelNumber, "ModelNumber")
	assertEqual(t, "http://www.meethue.com", device.ModelURL, "ModelURL")
	assertEqual(t, "93eadbeef13", device.SerialNumber, "SerialNumber")
	assertEqual(t, "uuid:01234567-89ab-cdef-0123-456789abcdef", device.UDN, "UDN")
	assertEqual(t, "", device.UPC, "UPC")
	assertEqual(t, "index.html", device.PresentationURL, "PresentationURL")

	icons := device.Icons
	assertEqual(t, 2, len(icons), "len(icons)")
	assertEqual(t, "image/png", icons[0].MIMEType, "icons.MIMEType")
	assertEqual(t, 48, icons[0].Width, "icons.Width")
	assertEqual(t, 48, icons[0].Height, "icons.Height")
	assertEqual(t, 24, icons[0].Depth, "icons.Depth")
	assertEqual(t, "hue_logo_0.png", icons[0].URL, "icons.URL")
}
