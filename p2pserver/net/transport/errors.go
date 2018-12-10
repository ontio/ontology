package transport

import "fmt"

type DialError struct {
	TransportName string
	IPAddr        string
	Err           string
}

func (this* DialError) Error() string {
	errStr := fmt.Sprintf("Dial to %s err by transport %s, the detailed err info is:%s", this.IPAddr, this.TransportName, this.Err)

	return errStr
}


