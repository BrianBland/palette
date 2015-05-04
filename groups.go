package palette

import (
	"github.com/BrianBland/go-hue"
)

func (p *Palette) SetGroup(lights []hue.Light, states []hue.LightState) error {
	var ret error
	for i, light := range lights {
		err := p.user.SetLightState(light.Id, &states[i%len(states)])
		if err != nil {
			ret = err
		}
	}
	return ret
}

func (p *Palette) SetComplementary(lights []hue.Light, primary hue.LightState) error {
	secondary := RotateDegrees(primary, 180)
	states := []hue.LightState{primary, secondary}
	return p.SetGroup(lights, states)
}

func (p *Palette) SetTriad(lights []hue.Light, primary hue.LightState) error {
	secondary := RotateDegrees(primary, 120)
	tertiary := RotateDegrees(primary, 240)
	states := []hue.LightState{primary, secondary, tertiary}
	return p.SetGroup(lights, states)
}

func (p *Palette) SetAnalogous(lights []hue.Light, primary hue.LightState) error {
	secondary := RotateDegrees(primary, 30)
	tertiary := RotateDegrees(primary, -30)
	states := []hue.LightState{primary, secondary, tertiary}
	return p.SetGroup(lights, states)
}

func (p *Palette) SetSplitComplementary(lights []hue.Light, primary hue.LightState) error {
	secondary := RotateDegrees(primary, 150)
	tertiary := RotateDegrees(primary, 210)
	states := []hue.LightState{primary, secondary, tertiary}
	return p.SetGroup(lights, states)
}

func (p *Palette) SetRectangle(lights []hue.Light, primary hue.LightState) error {
	accent := RotateDegrees(primary, 60)
	complementary := RotateDegrees(primary, 180)
	complementaryAccent := RotateDegrees(primary, 240)
	states := []hue.LightState{primary, accent, complementary, complementaryAccent}
	return p.SetGroup(lights, states)
}

func (p *Palette) SetSquare(lights []hue.Light, primary hue.LightState) error {
	accent := RotateDegrees(primary, 90)
	complementary := RotateDegrees(primary, 180)
	complementaryAccent := RotateDegrees(primary, 270)
	states := []hue.LightState{primary, accent, complementary, complementaryAccent}
	return p.SetGroup(lights, states)
}
