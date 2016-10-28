package notice

type device struct {
	sid  int64
	host string
}

type devices map[string]device

func newDevices() devices {
	return make(map[string]device)
}

func (d devices) put(dev string, v device) {
	d[dev] = v
}

func (d devices) get(dev string) (device, bool) {
	if dev, ok := d[dev]; ok {
		return dev, true
	}
	return device{}, false
}

func (d devices) del(dev string, sid int64) {
	if v, ok := d[dev]; ok {
		if v.sid == sid {
			delete(d, dev)
		}
	}
}

func (d devices) empty() bool {
	return len(d) == 0
}
