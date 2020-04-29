package keylight

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func (k *KeyLight) httpGet(ctx context.Context, path string, target interface{}) error {
	url := fmt.Sprintf("http://%s:%d/%s", k.DNSAddr, k.Port, path)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return json.NewDecoder(resp.Body).Decode(target)
}

func (k *KeyLight) httpPut(ctx context.Context, path string, body interface{}, target interface{}) error {
	buf := new(bytes.Buffer)
	err := json.NewEncoder(buf).Encode(body)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("http://%s:%d/%s", k.DNSAddr, k.Port, path)
	req, err := http.NewRequestWithContext(ctx, "PUT", url, buf)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return json.NewDecoder(resp.Body).Decode(target)
}

// FetchSettings allows you to retrieve general device settings.
func (k *KeyLight) FetchSettings(ctx context.Context) (*KeyLightSettings, error) {
	s := &KeyLightSettings{}
	err := k.httpGet(ctx, "elgato/lights/settings", s)
	return s, err
}

// FetchAccessoryInfo returns metadata for the accessory.
func (k *KeyLight) FetchAccessoryInfo(ctx context.Context) (*AccessoryInfo, error) {
	i := &AccessoryInfo{}
	err := k.httpGet(ctx, "elgato/accessory-info", i)
	return i, err
}

// FetchLightOptions returns all of the individual lights that are owned by an
// accessory. This in conjunction with UpdateLightOptions will allow you to
// control your lights.
func (k *KeyLight) FetchLightOptions(ctx context.Context) (*KeyLightOptions, error) {
	o := &KeyLightOptions{Lights: make([]*KeyLightLight, 0)}
	err := k.httpGet(ctx, "elgato/lights", o)
	return o, err
}

// UpdateLightOptions allows you to update the settings for individual lights
// in an accessory. It returns the updated options.
func (k *KeyLight) UpdateLightOptions(ctx context.Context, newOptions *KeyLightOptions) (*KeyLightOptions, error) {
	o := &KeyLightOptions{Lights: make([]*KeyLightLight, 0)}
	err := k.httpPut(ctx, "elgato/lights", newOptions, o)
	return o, err
}
