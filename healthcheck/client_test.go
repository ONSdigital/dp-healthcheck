package healthcheck

import (
	"context"
	"fmt"
	rchttp "github.com/ONSdigital/dp-rchttp"
	. "github.com/smartystreets/goconvey/convey"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
)

var client = &rchttp.Client{HTTPClient: &http.Client{}}
var ctx = context.Background()

type MockedHTTPResponse struct {
	StatusCode int
	Body       string
}

func getMockHealthCheckAPI(expectRequest http.Request, mockedHTTPResponse MockedHTTPResponse) *Client {
	httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != expectRequest.Method {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("unexpected HTTP method used"))
			return
		}
		w.WriteHeader(mockedHTTPResponse.StatusCode)
		fmt.Fprintln(w, mockedHTTPResponse.Body)
	}))
	sc := Checker(someChecker)
	return NewClient(nil, )
	}
}

func TestHealthCheck_Handler(t *testing.T) {
	Convey("Create a new Health Check", t, func() {

	})
}
