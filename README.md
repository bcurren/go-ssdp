# go-ssdp

A library that implements the client side of SSDP (Simple Service Discovery Protocol).

## Usage

* [Setup your go environment](http://golang.org/doc/code.html)
* ```go get http://github.com/bcurren/go-ssdp```
* Write code using the library.

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