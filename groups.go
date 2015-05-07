package palette

import (
	"sync"

	"github.com/BrianBland/go-hue"
)

type LightAttributesOrError struct {
	*hue.LightAttributes
	Error error
}

func (p *Palette) GetGroup(lights []hue.Light) <-chan LightAttributesOrError {
	res := make(chan LightAttributesOrError, len(lights))
	var wg sync.WaitGroup
	wg.Add(len(lights))

	getLight := func(i int, res chan<- LightAttributesOrError) {
		attrs, err := p.GetLightAttributes(lights[i].Id)
		res <- LightAttributesOrError{LightAttributes: attrs, Error: err}
		wg.Done()
	}
	for i := range lights {
		go getLight(i, res)
	}

	go func() {
		wg.Wait()
		close(res)
	}()

	return res
}

func (p *Palette) SetGroup(lights []hue.Light, states []hue.LightState) <-chan error {
	res := make(chan error, len(lights))
	var wg sync.WaitGroup
	wg.Add(len(lights))

	setLight := func(i int, res chan<- error) {
		res <- p.SetLightState(lights[i].Id, &states[i%len(states)])
		wg.Done()
	}
	for i := range lights {
		go setLight(i, res)
	}

	go func() {
		wg.Wait()
		close(res)
	}()

	return res
}

func (p *Palette) SetComplementary(lights []hue.Light, primary hue.LightState) <-chan error {
	secondary := RotateDegrees(primary, 180)
	states := []hue.LightState{primary, secondary}
	return p.SetGroup(lights, states)
}

func (p *Palette) SetTriad(lights []hue.Light, primary hue.LightState) <-chan error {
	secondary := RotateDegrees(primary, 120)
	tertiary := RotateDegrees(primary, 240)
	states := []hue.LightState{primary, secondary, tertiary}
	return p.SetGroup(lights, states)
}

func (p *Palette) SetAnalogous(lights []hue.Light, primary hue.LightState) <-chan error {
	secondary := RotateDegrees(primary, 30)
	tertiary := RotateDegrees(primary, -30)
	states := []hue.LightState{primary, secondary, tertiary}
	return p.SetGroup(lights, states)
}

func (p *Palette) SetSplitComplementary(lights []hue.Light, primary hue.LightState) <-chan error {
	secondary := RotateDegrees(primary, 150)
	tertiary := RotateDegrees(primary, 210)
	states := []hue.LightState{primary, secondary, tertiary}
	return p.SetGroup(lights, states)
}

func (p *Palette) SetRectangle(lights []hue.Light, primary hue.LightState) <-chan error {
	accent := RotateDegrees(primary, 60)
	complementary := RotateDegrees(primary, 180)
	complementaryAccent := RotateDegrees(primary, 240)
	states := []hue.LightState{primary, accent, complementary, complementaryAccent}
	return p.SetGroup(lights, states)
}

func (p *Palette) SetSquare(lights []hue.Light, primary hue.LightState) <-chan error {
	accent := RotateDegrees(primary, 90)
	complementary := RotateDegrees(primary, 180)
	complementaryAccent := RotateDegrees(primary, 270)
	states := []hue.LightState{primary, accent, complementary, complementaryAccent}
	return p.SetGroup(lights, states)
}
