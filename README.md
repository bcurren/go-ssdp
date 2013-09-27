# go-ssdp

A library that implements the client side of SSDP (Simple Service Discovery Protocol).

Please see [godoc.org](http://godoc.org/github.com/bcurren/go-ssdp) for a detailed API
description.

## Usage

* [Setup your go environment](http://golang.org/doc/code.html)
* ```go get github.com/bcurren/go-ssdp```
* Write code using the library.

### Get Device for devices on the network
```Go
package main

import (
	"github.com/bcurren/go-ssdp"
	"time"
	"fmt"
)

func main() {
	devices, err := ssdp.SearchForDevices("upnp:rootdevice", 3*time.Second)
	if err != nil {
		return
	}

	for _, device := range devices {
		fmt.Println(device.ModelName)
	}
}
```

### Get Responses for Search on the network
```Go
package main

import (
	"github.com/bcurren/go-ssdp"
	"time"
	"fmt"
)

func main() {
	responses, err := ssdp.Search("upnp:rootdevice", 3*time.Second)
	if err != nil {
		return
	}

	for _, response := range responses {
		// Do something with the response you discover
		fmt.Println(response)
	}
}
```
## How to contribute
* Fork
* Write tests and code
* Run go fmt
* Submit a pull request

