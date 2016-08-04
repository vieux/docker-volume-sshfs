package volume

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/docker/go-connections/sockets"
)

func TestHandler(t *testing.T) {
	p := &testPlugin{}
	h := NewHandler(p)
	l := sockets.NewInmemSocket("test", 0)
	go h.Serve(l)
	defer l.Close()

	client := &http.Client{Transport: &http.Transport{
		Dial: l.Dial,
	}}

	// Create
	resp, err := pluginRequest(client, createPath, Request{Name: "foo"})
	if err != nil {
		t.Fatal(err)
	}
	if resp.Err != "" {
		t.Fatalf("error while creating volume: %v", err)
	}
	if p.create != 1 {
		t.Fatalf("expected create 1, got %d", p.create)
	}

	// Get
	resp, err = pluginRequest(client, getPath, Request{Name: "foo"})
	if err != nil {
		t.Fatal(err)
	}
	if resp.Err != "" {
		t.Fatalf("got error getting volume: %s", resp.Err)
	}
	if resp.Volume.Name != "foo" {
		t.Fatalf("expected volume `foo`, got %v", resp.Volume)
	}
	if p.get != 1 {
		t.Fatalf("expected get 1, got %d", p.get)
	}

	// List
	resp, err = pluginRequest(client, listPath, Request{Name: "foo"})
	if err != nil {
		t.Fatal(err)
	}
	if resp.Err != "" {
		t.Fatalf("expected no volume, got: %s", resp.Err)
	}
	if len(resp.Volumes) != 1 {
		t.Fatalf("expected 1 volume, got %v", resp.Volumes)
	}
	if resp.Volumes[0].Name != "foo" {
		t.Fatalf("expected volume `foo`, got %v", resp.Volumes[0])
	}
	if p.list != 1 {
		t.Fatalf("expected list 1, got %d", p.list)
	}

	// Path
	if _, err := pluginRequest(client, hostVirtualPath, Request{Name: "foo"}); err != nil {
		t.Fatal(err)
	}
	if p.path != 1 {
		t.Fatalf("expected path 1, got %d", p.path)
	}

	// Mount
	if _, err := pluginRequest(client, mountPath, Request{Name: "foo"}); err != nil {
		t.Fatal(err)
	}
	if p.mount != 1 {
		t.Fatalf("expected mount 1, got %d", p.mount)
	}

	// Unmount
	if _, err := pluginRequest(client, unmountPath, Request{Name: "foo"}); err != nil {
		t.Fatal(err)
	}
	if p.unmount != 1 {
		t.Fatalf("expected unmount 1, got %d", p.unmount)
	}

	// Remove
	resp, err = pluginRequest(client, removePath, Request{Name: "foo"})
	if err != nil {
		t.Fatal(err)
	}
	if resp.Err != "" {
		t.Fatalf("got error removing volume: %s", resp.Err)
	}
	if p.remove != 1 {
		t.Fatalf("expected remove 1, got %d", p.remove)
	}

	// Capabilities
	resp, err = pluginRequest(client, capabilitiesPath, Request{})
	if err != nil {
		t.Fatal(err)
	}
	if resp.Err != "" {
		t.Fatalf("got error removing volume: %s", resp.Err)
	}
	if p.capabilities != 1 {
		t.Fatalf("expected remove 1, got %d", p.capabilities)
	}
}

func pluginRequest(client *http.Client, method string, req Request) (*Response, error) {
	b, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}
	resp, err := client.Post("http://localhost"+method, "application/json", bytes.NewReader(b))
	if err != nil {
		return nil, err
	}
	var vResp Response
	err = json.NewDecoder(resp.Body).Decode(&vResp)
	if err != nil {
		return nil, err
	}

	return &vResp, nil
}

type testPlugin struct {
	volumes      []string
	create       int
	get          int
	list         int
	path         int
	mount        int
	unmount      int
	remove       int
	capabilities int
}

func (p *testPlugin) Create(req Request) Response {
	p.create++
	p.volumes = append(p.volumes, req.Name)
	return Response{}
}

func (p *testPlugin) Get(req Request) Response {
	p.get++
	for _, v := range p.volumes {
		if v == req.Name {
			return Response{Volume: &Volume{Name: v}}
		}
	}
	return Response{Err: "no such volume"}
}

func (p *testPlugin) List(req Request) Response {
	p.list++
	var vols []*Volume
	for _, v := range p.volumes {
		vols = append(vols, &Volume{Name: v})
	}
	return Response{Volumes: vols}
}

func (p *testPlugin) Remove(req Request) Response {
	p.remove++
	for i, v := range p.volumes {
		if v == req.Name {
			p.volumes = append(p.volumes[:i], p.volumes[i+1:]...)
			return Response{}
		}
	}
	return Response{Err: "no such volume"}
}

func (p *testPlugin) Path(req Request) Response {
	p.path++
	for _, v := range p.volumes {
		if v == req.Name {
			return Response{}
		}
	}
	return Response{Err: "no such volume"}
}

func (p *testPlugin) Mount(req MountRequest) Response {
	p.mount++
	for _, v := range p.volumes {
		if v == req.Name {
			return Response{}
		}
	}
	return Response{Err: "no such volume"}
}

func (p *testPlugin) Unmount(req UnmountRequest) Response {
	p.unmount++
	for _, v := range p.volumes {
		if v == req.Name {
			return Response{}
		}
	}
	return Response{Err: "no such volume"}
}

func (p *testPlugin) Capabilities(req Request) Response {
	p.capabilities++
	return Response{Capabilities: Capability{Scope: "local"}}
}
