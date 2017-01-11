package oauth

import (
	"fmt"
	"net"
	"net/http"
	"time"
)

func logRequest(w http.ResponseWriter, r *http.Request, info *jwtInfo) {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		host = r.RemoteAddr
	}

	timeStr := time.Now().Format("2/Jan/2006:15:04:05 -0700")

	str := fmt.Sprintf(`%v [%v] "%v %v %v" %v`, host, timeStr, r.Method, r.URL.Path, r.Proto, info.Username)

	fmt.Printf("%v\n", str)
}
