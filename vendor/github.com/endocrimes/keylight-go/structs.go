package keylight

import "context"

// Device represents an actual KeyLight. They can be found using a Discovery
// interface, or created with a given DNSAddr and Port when running in a static
// environment.
type Device struct {
	Name    string
	DNSAddr string
	Port    int
}

// DeviceSettings represents the data returned and accepted by a KeyLight for
// configuring its behaviour.
type DeviceSettings struct {
	PowerOnBehavior       int `json:"powerOnBehavior"`
	PowerOnBrightness     int `json:"powerOnBrightness"`
	PowerOnTemperature    int `json:"powerOnTemperature"`
	SwitchOnDurationMs    int `json:"switchOnDurationMs"`
	SwitchOffDurationMs   int `json:"switchOffDurationMs"`
	ColorChangeDurationMs int `json:"colorChangeDurationMs"`
}

// DeviceInfo returns read-only information about a given Elgato device.
// This is currently only supported by keylight devices but as other devices
// may expose _elg.tcp services, it may also be suppported by other products
// in the future.
type DeviceInfo struct {
	ProductName         string   `json:"productName"`
	HardwareBoardType   int      `json:"hardwareBoardType"`
	FirmwareBuildNumber int      `json:"firmwareBuildNumber"`
	FirmwareVersion     string   `json:"firmwareVersion"`
	SerialNumber        string   `json:"serialNumber"`
	DisplayName         string   `json:"displayName"`
	Features            []string `json:"features"`
}

// Light is the struct that encapsulates the state of a KeyLight's individual
// light. Most keylight devices currently only have one 'Light'.
type Light struct {
	On          int `json:"on"`
	Brightness  int `json:"brightness"`
	Temperature int `json:"temperature"`
}

// Copy returns a new copy of a light.
func (l *Light) Copy() *Light {
	nl := new(Light)
	*nl = *l
	return nl
}

// LightGroup represents a set of configurable lights within a given KeyLight
// device.
type LightGroup struct {
	Count  int      `json:"numberOfLights"`
	Lights []*Light `json:"lights"`
}

// Copy returns a new deep copy of a LightGroup.
func (o *LightGroup) Copy() *LightGroup {
	no := new(LightGroup)
	*no = *o

	lights := make([]*Light, len(o.Lights))
	for idx, light := range o.Lights {
		lights[idx] = light.Copy()
	}

	no.Lights = lights

	return no
}

// Discovery is the interface that is exposed to discover KeyLight devices.
// The default implementation discovers lights via Bonjour.
type Discovery interface {
	// Run will start the given discovery client and run syncronously until the
	// provided context is shutdown.
	Run(ctx context.Context) error

	// ResultsCh returns a channel of discovered Key Lights.
	// NOTE: It currently does not filter the results so may return non key light
	//       entities if they expose the `_elg._tcp` service over mdns.
	//       Use KeyLight.FetchDeviceInfo(...) to determine the accessory info
	//       if you run into problems.
	ResultsCh() <-chan *Device
}

// NewDiscovery returns a new default Disocvery implemetnation. This is currently
// backed by Bonjour.
func NewDiscovery() (Discovery, error) {
	return newBonjourDiscovery()
}
