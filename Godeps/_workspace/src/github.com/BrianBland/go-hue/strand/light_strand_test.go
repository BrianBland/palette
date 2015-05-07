package strand

import (
	"github.com/bcurren/go-hue"
	"github.com/bcurren/go-hue/huetest"
	"strings"
	"testing"
)

func Test_NewLightStrand(t *testing.T) {
	stubHueAPI := &huetest.StubAPI{}
	lightStrand := NewLightStrand(3, stubHueAPI)

	if lightStrand.Length != 3 {
		t.Error("Light strand should have length given in constructor.")
	}
	if lightStrand.Lights.Length() != 0 {
		t.Error("Light strand should have Lights initialized to empty.")
	}
}

func Test_MapUnmappedLights(t *testing.T) {
	hueLights := make([]hue.Light, 1, 1)
	hueLights[0].Id = "3"

	stubHueAPI := &huetest.StubAPI{}
	stubHueAPI.GetLightsError = nil
	stubHueAPI.GetLightsReturn = hueLights

	lightStrand := NewLightStrand(1, stubHueAPI)
	normalState := &hue.LightState{}
	seletedState := &hue.LightState{}

	countTimesCalled := 0
	err := lightStrand.MapUnmappedLights(normalState, seletedState, func(lightId string) string {
		countTimesCalled += 1

		// Callback param is light id
		if lightId != "3" {
			t.Error("Callback parameter should have lightId 3.")
		}
		// Should have set hue light to red
		if stubHueAPI.SetLightStateParamLightId != "3" {
			t.Error("Should have set light 3.")
		}
		if stubHueAPI.SetLightStateParamLightState == normalState {
			t.Error("Should have set the state.")
		}

		return "1"
	})
	if err != nil {
		t.Fatal("Error returned when mapping")
	}

	if countTimesCalled != 1 {
		t.Error("Map function called more than 1 time.")
	}
	if lightStrand.Lights.GetValue("1") != "3" {
		t.Error("Didn't map to the correct id.")
	}

	// Should have set hue light to white
	if stubHueAPI.SetLightStateParamLightId != "3" {
		t.Error("Should have set light 3.")
	}
	if stubHueAPI.SetLightStateParamLightState == seletedState {
		t.Error("Should have set the state.")
	}
}

func Test_MapUnmappedLightsSkipXSocketIds(t *testing.T) {
	hueLights := make([]hue.Light, 1, 1)
	hueLights[0].Id = "3"

	stubHueAPI := &huetest.StubAPI{}
	stubHueAPI.GetLightsError = nil
	stubHueAPI.GetLightsReturn = hueLights

	lightStrand := NewLightStrand(3, stubHueAPI)
	state := &hue.LightState{}

	err := lightStrand.MapUnmappedLights(state, state, func(string) string {
		return "x"
	})
	if err != nil {
		t.Fatal("Error returned when mapping")
	}

	if lightStrand.Lights.Length() != 0 {
		t.Error("Should skip mapping when socket id is 'x'.")
	}
}

func Test_cleanInvalidMappedLightIds(t *testing.T) {
	hueLights := make([]hue.Light, 4, 4)
	hueLights[0].Id = "light3"

	lightStrand := NewLightStrand(3, &huetest.StubAPI{})
	lightStrand.Lights.Set("1", "light3")
	lightStrand.Lights.Set("2", "missinglightid")

	lightStrand.cleanInvalidMappedLightIds(hueLights)
	if lightStrand.Lights.Length() != 1 {
		t.Errorf("Should have deleted the bad key value")
	}
	if lightStrand.Lights.GetValue("2") != "" {
		t.Errorf("Should have deleted key 2")
	}
}

func Test_getUnmappedLightIds(t *testing.T) {
	hueLights := make([]hue.Light, 4, 4)
	hueLights[0].Id = "light3"
	hueLights[1].Id = "light1"
	hueLights[2].Id = "light5"
	hueLights[3].Id = "light2"

	lightStrand := NewLightStrand(3, &huetest.StubAPI{})
	lightStrand.Lights.Set("1", "light3")
	lightStrand.Lights.Set("2", "light2")
	lightStrand.Lights.Set("3", "light1")

	expected := []string{"light5"}
	actual := lightStrand.getUnmappedLightIds(hueLights)
	if !stringSlicesEqual(expected, actual) {
		t.Errorf("Expected a slice of all unmapped light ids. Expected %v but received %v.\n", expected, actual)
	}
}

// A function to test if two string slices are equal. If you know a better way, please
// update this function. Seems like there should be a beeter way but I couldn't find one.
func stringSlicesEqual(slice1 []string, slice2 []string) bool {
	// Check both nil or both not nil
	if slice1 == nil && slice2 == nil {
		return true
	} else if slice1 == nil && slice2 != nil {
		return false
	} else if slice1 != nil && slice2 == nil {
		return false
	}

	// Length must be the same
	if len(slice1) != len(slice2) {
		return false
	}

	// Contents must be the same
	sep := "|||"
	slices1String := strings.Join(slice1, sep)
	slices2String := strings.Join(slice2, sep)
	if slices1String != slices2String {
		return false
	}

	return true
}

func Test_validSocketId(t *testing.T) {
	lightStrand := NewLightStrand(3, nil)
	if lightStrand.validSocketId("0") {
		t.Error("Socket id 0 should be invalid.")
	}
	if !lightStrand.validSocketId("1") {
		t.Error("Socket id 1 should be valid.")
	}
	if !lightStrand.validSocketId("3") {
		t.Error("Socket id 3 should be valid.")
	}
	if lightStrand.validSocketId("4") {
		t.Error("Socket id 4 should be invalid.")
	}
	if lightStrand.validSocketId("notint") {
		t.Error("Socket id that is not an int should be invalid.")
	}
}

func Test_IsMappedSocketId(t *testing.T) {
	lightStrand := NewLightStrand(3, nil)
	if lightStrand.IsMappedSocketId("1") != false {
		t.Error("Should return false since it's not mapped.")
	}

	lightStrand.Lights.Set("1", "lightx")
	if lightStrand.IsMappedSocketId("1") != true {
		t.Error("Should return true since it's mapped.")
	}
}

func Test_ChangeLength(t *testing.T) {
	lightStrand := NewLightStrand(3, nil)
	lightStrand.Lights.Set("1", "light1")
	lightStrand.Lights.Set("2", "light2")
	lightStrand.Lights.Set("3", "light3")

	lightStrand.ChangeLength(2)
	if lightStrand.Length != 2 {
		t.Error("Should have changed the length to 2")
	}
	if lightStrand.Lights.Length() != 2 {
		t.Error("Should have changed the lights map length to 2.")
	}
	if lightStrand.Lights.GetValue("1") != "light1" {
		t.Error("Light 1 should still exist.")
	}
	if lightStrand.Lights.GetValue("2") != "light2" {
		t.Error("Light 2 should still exist.")
	}
	if lightStrand.Lights.GetValue("3") != "" {
		t.Error("Light 3 should be deleted.")
	}
}
