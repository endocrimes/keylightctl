package keylight

import "context"

type KeyLight struct {
	Name    string
	DNSAddr string
	Port    int
}

type KeyLightSettings struct {
	PowerOnBehavior       int `json:"powerOnBehavior"`
	PowerOnBrightness     int `json:"powerOnBrightness"`
	PowerOnTemperature    int `json:"powerOnTemperature"`
	SwitchOnDurationMs    int `json:"switchOnDurationMs"`
	SwitchOffDurationMs   int `json:"switchOffDurationMs"`
	ColorChangeDurationMs int `json:"colorChangeDurationMs"`
}

type AccessoryInfo struct {
	ProductName         string   `json:"productName"`
	HardwareBoardType   int      `json:"hardwareBoardType"`
	FirmwareBuildNumber int      `json:"firmwareBuildNumber"`
	FirmwareVersion     string   `json:"firmwareVersion"`
	SerialNumber        string   `json:"serialNumber"`
	DisplayName         string   `json:"displayName"`
	Features            []string `json:"features"`
}

type KeyLightLight struct {
	On          int `json:"on"`
	Brightness  int `json:"brightness"`
	Temperature int `json:"temperature"`
}

func (l *KeyLightLight) Copy() *KeyLightLight {
	nl := new(KeyLightLight)
	*nl = *l
	return nl
}

type KeyLightOptions struct {
	Count  int              `json:"numberOfLights"`
	Lights []*KeyLightLight `json:"lights"`
}

func (o *KeyLightOptions) Copy() *KeyLightOptions {
	no := new(KeyLightOptions)
	*no = *o

	lights := make([]*KeyLightLight, len(o.Lights))
	for idx, light := range o.Lights {
		lights[idx] = light.Copy()
	}

	no.Lights = lights

	return no
}

type Discovery interface {
	// Run will start the given discovery client and run syncronously until the
	// provided context is shutdown.
	Run(ctx context.Context) error

	// ResultsCh returns a channel of discovered Key Lights.
	// NOTE: It currently does not filter the results so may return non key light
	//       entities if they expose the `_elg._tcp` service over mdns.
	//       Use KeyLight.FetchAccessoryInfo(...) to determine the accessory info
	//       if you run into problems.
	ResultsCh() <-chan *KeyLight
}

func NewDiscovery() (Discovery, error) {
	return newBonjourDiscovery()
}
