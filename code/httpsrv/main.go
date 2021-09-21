package main

import (
	"fmt"
	"httpsrv/internal"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

// DefaultServerCloseSIG default close sig
var DefaultServerCloseSIG = []os.Signal{syscall.SIGINT, syscall.SIGTERM, syscall.SIGSEGV}

type ServiceImp struct {
}

func (handler *ServiceImp) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s, _ := ioutil.ReadAll(r.Body)
	fmt.Fprintf(w, "%s", s)
}

func main() {
	ch := make(chan os.Signal)
	s := internal.NewServer(internal.WithAddr(":8080"),
		internal.WithService("echo", &ServiceImp{}))
	go func() {
		if err := s.Serve(); err != nil {
			fmt.Printf("Serve error, err:%+v\n", err)
			ch <- syscall.SIGTERM
		}
	}()

	signal.Notify(ch, DefaultServerCloseSIG...)
	sig := <-ch

	fmt.Println("server exist, sig:", sig)

	return
}

/* Output1 test sig:
Serve error, err:test sig
httpsrv/internal.(*Server).Serve
        /root/source/github/go-dev/code/httpsrv/internal/server.go:58
main.main.func1
        /root/source/github/go-dev/code/httpsrv/main.go:29
runtime.goexit
        /usr/local/go/src/runtime/asm_amd64.s:1581
server exist, sig: terminated
   Output2 normal:
curl 127.0.0.1:8080/echo -d "hello"
hello
*/
