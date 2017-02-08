package main

type DeviceInformations struct {
	x *map[uint32]string
}

func (d *DeviceInformations)Initialize() {
	d.x = make(map[uint32]string)
}

func (d *DeviceInformations)RegisterDevice(deviceID uint32, deviceName string) {
	d.x[deviceID] = deviceName
}

func (d *DeviceInformations)GetDeviceName(deviceID uint32) (string, bool) {
	name, isExist := d.x[deviceID]
	if isExist == false {
		return "", false
	}

	return name, true
}