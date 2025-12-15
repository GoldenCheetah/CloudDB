package main

import (
	"testing"

	"github.com/emicklei/go-restful"
)

func TestRoutesRegistered(t *testing.T) {
	// init() is called automatically

	services := restful.RegisteredWebServices()
	if len(services) == 0 {
		t.Fatal("No services registered")
	}

	foundV1 := false
	for _, ws := range services {
		if ws.RootPath() == "/v1" {
			foundV1 = true
			break
		}
	}

	if !foundV1 {
		t.Error("Service /v1 not found")
	}
}
