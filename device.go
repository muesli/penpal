package main

import (
	"github.com/godbus/dbus"
)

// findDevices returns a list of all paired Bamboo devices
func findDevices() ([]dbus.ObjectPath, error) {
	obj := conn.Object("org.freedesktop.tuhi1", dbus.ObjectPath("/org/freedesktop/tuhi1"))
	v, err := obj.GetProperty("org.freedesktop.tuhi1.Manager.Devices")
	if err != nil {
		return nil, err
	}

	return v.Value().([]dbus.ObjectPath), nil
}

// findDrawings returns a list of all available drawings for a specific device
func findDrawings(dev dbus.ObjectPath) ([]uint64, error) {
	obj := conn.Object("org.freedesktop.tuhi1", dev)
	v, err := obj.GetProperty("org.freedesktop.tuhi1.Device.DrawingsAvailable")
	if err != nil {
		return nil, err
	}

	return v.Value().([]uint64), nil
}

// fetchDrawing retrieves the JSON data of a specific drawing
func fetchDrawing(dev dbus.ObjectPath, drawing uint64) ([]byte, error) {
	var data []byte

	obj := conn.Object("org.freedesktop.tuhi1", dev)
	err := obj.Call("org.freedesktop.tuhi1.Device.GetJSONData", 0, uint32(1), drawing).Store(&data)
	if err != nil {
		return nil, err
	}

	return data, nil
}

// startListening starts listening for new events from a Bamboo device
func startListening(dev dbus.ObjectPath) error {
	obj := conn.Object("org.freedesktop.tuhi1", dev)
	call := obj.Call("org.freedesktop.tuhi1.Device.StartListening", 0)
	if call.Err != nil {
		return call.Err
	}

	return nil
}
