package palette

import (
	"github.com/BrianBland/go-hue"
)

const hueResolution = 2 << 15

func RotateDegrees(state hue.LightState, degrees float64) hue.LightState {
	rotatedHue := (uint16)(
		(uint32)(
			(float64)(*state.Hue)+hueResolution*(degrees/360.0),
		) % hueResolution)
	return hue.LightState{
		On:         state.On,
		Brightness: state.Brightness,
		Hue:        &rotatedHue,
		Saturation: state.Saturation,
		Alert:      state.Alert,
		Effect:     state.Effect,
	}
}
