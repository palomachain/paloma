package app

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"sync"
	"syscall"
	"time"
)

const (
	PigeonHealthCheckPort = "PIGEON_HEALTHCHECK_PORT"
)

type pigeonHealthCheck struct {
	Pid int `json:"pid"`
}

func onceFunc[V any](fn func() V) func() V {
	var once sync.Once
	var val V
	once.Do(func() {
		val = fn()
	})
	return func() V {
		return val
	}
}

var PigeonHTTPClient = onceFunc(func() *http.Client {
	return &http.Client{
		Timeout: time.Second * 5,
		Transport: &http.Transport{
			Dial: (&net.Dialer{
				Timeout: 5 * time.Second,
			}).Dial,
			TLSHandshakeTimeout: 5 * time.Second,
		},
	}
})

type httpClienter interface {
	Do(*http.Request) (*http.Response, error)
}

func GetPigonListenPort() int {
	listenPort, ok := os.LookupEnv(PigeonHealthCheckPort)

	if !ok || listenPort == "" {
		log.Fatal("pigeon listen port environment variable is not set")
	}

	port, err := strconv.ParseInt(listenPort, 10, 32)
	if err != nil {
		log.Fatalf("pigeon's port %s is invalid: %v", listenPort, err)
	}

	return int(port)
}

func CheckPigeonRunningLooper(ctx context.Context, client httpClienter) {
	PigeonMustRun(ctx, client)

	t := time.NewTicker(5 * time.Second)
	for range t.C {
		if err := ctx.Err(); err != nil {
			return
		}
		PigeonMustRun(ctx, client)
	}
}

func PigeonMustRun(ctx context.Context, client httpClienter) {
	port := GetPigonListenPort()
	// extract port from url and verify that it's open locally
	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("http://127.0.0.1:%d", port), nil)
	if err != nil {
		log.Fatalf("couldn't create an HTTP request: %v", err)
	}

	res, err := client.Do(req)
	if err != nil {
		log.Fatalf("error communicating with pigeon: %v. Is your pigeon running?", err)
	}

	defer res.Body.Close()

	var hc pigeonHealthCheck
	if err := json.NewDecoder(res.Body).Decode(&hc); err != nil {
		log.Fatalf("error encoding pigeon's response: %v", err)
	}

	if err := checkPid(hc.Pid); err != nil {
		log.Fatalf("error verifying the pid: %v", err)
	}
}

// checkPid checks if the pid is actually running on this machine.
// this is not the best check, but it's good enough.
func checkPid(pid int) error {
	process, err := os.FindProcess(pid)
	if err != nil {
		return err
	}

	if err := process.Signal(syscall.Signal(0)); err != nil {
		return err
	}
	return nil
}
