# go-udpn

A library that implements the client side of the UDPN (Universal Plug and Play) protocol. This library, currently, supports the  search / discovery method and nothing else. Please feel free to fork and add more implementations of the protocol.

## Usage

* [Setup your go environment](http://golang.org/doc/code.html)
* ```go get http://github.com/bcurren/go-udpn```
* Write code using the library.

```Go
import (
	"udpn"
	"time"
	"fmt"
)

responses, err = udpn.Search("upnp:rootdevice", time.Duration(3)*time.Second)
if err {
	return
}

for _, response := range responses {
	// Do something with the response you discover
	fmt.Println(response)
}

```
## How to contribute
* Fork
* Write tests and code
* Run go fmt
* Submit a pull request