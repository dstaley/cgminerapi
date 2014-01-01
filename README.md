# cgminerapi [![GoDoc](https://godoc.org/github.com/dstaley/cgminerapi?status.png)](http://godoc.org/github.com/dstaley/cgminerapi)

cgminerapi is a simple Go library for accessing cgminer's API.

## Example
```go
package main

import (
	"fmt"
	"github.com/dstaley/cgminerapi"
)

func main() {
	client := cgminerapi.NewCgminerAPI("localhost", "4028")
	c := cgminerapi.APICommand{Method: "summary"}
	r, err := client.Send(&c)
	if err != nil {
		fmt.Printf("%s\n", err)
	}
	fmt.Printf("%+v", r)
}
```

## License
MIT