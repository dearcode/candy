package util

import (
	"testing"
)

/*
&{Begin:7480 End:9973 Host:127.0.0.1:3333}, size:2493
&{Begin:4987 End:7480 Host:127.0.0.1:2222}, size:2493
&{Begin:2494 End:4987 Host:127.0.0.1:4444}, size:2493
&{Begin:   0 End:2494 Host:127.0.0.1:1111}, size:2494

*/

func TestRegion(t *testing.T) {
	hosts := []string{
		"127.0.0.1:1111",
		"127.0.0.1:2222",
		"127.0.0.1:3333",
		"127.0.0.1:4444",
	}

	results := []struct {
		id   int64
		host string
	}{
		{0, hosts[0]},
		{2493, hosts[0]},
		{2494, hosts[3]},
		{4986, hosts[3]},
		{4987, hosts[1]},
		{7479, hosts[1]},
		{7489, hosts[2]},
		{9972, hosts[2]},
	}

	r, _ := NewRegions()

	for _, h := range hosts {
		o, err := r.Max()
		if err == ErrRegionNotFound {
			r.Split("", h)
		} else {
			r.Split(o.Host, h)
		}
	}

	for _, res := range results {
		region, ok := r.Get(res.id)
		if !ok {
			t.Fatalf("id:%d not found", res.id)
		}
		if res.host != region.Host {
			t.Fatalf("find id:%d, host:%v, expect host:%s\n", res.id, region.Host, res.host)
		}

		if _, err := r.GetByHost(res.host); err != nil {
			t.Fatalf("host:%s not found", res.host)
		}

	}
}

func TestMax(t *testing.T) {
	hosts := []string{
		"127.0.0.1:1111",
		"127.0.0.1:2222",
		"127.0.0.1:3333",
		"127.0.0.1:4444",
	}

	r, _ := NewRegions()
	_, err := r.Max()
	if err != ErrRegionNotFound {
		t.Fatalf("regions is empty, expect ErrRegionNotFound")
	}

	r.Split("", hosts[0])

	m, err := r.Max()
	if err != nil {
		t.Fatalf("regions max err:%s, expect nil", err.Error())
	}

	if m.Host != hosts[0] {
		t.Fatalf("region host:%s, expect:%s", m.Host, hosts[0])
	}

	r.Split(m.Host, hosts[1])

	m, err = r.Max()
	if err != nil {
		t.Fatalf("regions max err:%s, expect nil", err.Error())
	}
	if m.Host != hosts[1] {
		t.Fatalf("region host:%s, expect:%s", m.Host, hosts[1])
	}
	r.Split(m.Host, hosts[2])

	m, err = r.Max()
	if err != nil {
		t.Fatalf("regions max err:%s, expect nil", err.Error())
	}

	if m.Host != hosts[0] {
		t.Fatalf("region host:%s, expect:%s", m.Host, hosts[0])
	}
}

func TestRegionHostsInit(t *testing.T) {
	hosts := []string{
		"127.0.0.1:1111",
		"127.0.0.1:2222",
		"127.0.0.1:3333",
		"127.0.0.1:4444",
	}

	r, _ := NewRegions()
	for _, h := range hosts {
		o, err := r.Max()
		if err == ErrRegionNotFound {
			r.Split("", h)
		} else {
			r.Split(o.Host, h)
		}
	}

	_, err := r.Marshal()
	if err != nil {
		t.Fatalf("Marshal error")
	}

	result := `{"127.0.0.1:1111":{"Begin":0,"End":2494,"Host":"127.0.0.1:1111"},"127.0.0.1:2222":{"Begin":4987,"End":7480,"Host":"127.0.0.1:2222"},"127.0.0.1:3333":{"Begin":7480,"End":9973,"Host":"127.0.0.1:3333"},"127.0.0.1:4444":{"Begin":2494,"End":4987,"Host":"127.0.0.1:4444"}}`

	r, err = NewRegions(RegionsWithHosts(result))
	if err != nil {
		t.Fatalf("newRegions error:%s", err.Error())
	}

	for _, h := range hosts {
		if _, err := r.GetByHost(h); err != nil {
			t.Fatalf("can't find host:%s", h)
		}
	}

}
