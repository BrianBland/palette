package strand

import (
	"errors"
	"fmt"
	"github.com/bcurren/go-hue"
	"time"
)

// GetLights() is same as hue.User.GetLights() except all light ids are mapped to
// socket ids.
func (lg *LightStrand) GetLights() ([]hue.Light, error) {
	lights, err := lg.API.GetLights()
	return lg.mapLightIdsToSocketIds(lights), err
}

// GetNewLights() is same as hue.User.GetNewLights() except all light ids are mapped to
// socket ids.
func (lg *LightStrand) GetNewLights() ([]hue.Light, time.Time, error) {
	lights, lastUpdated, err := lg.API.GetNewLights()
	return lg.mapLightIdsToSocketIds(lights), lastUpdated, err
}

// SearchForNewLights() is same as hue.User.SearchForNewLights() except all light ids are mapped to
// socket ids.
func (lg *LightStrand) SearchForNewLights() error {
	return lg.API.SearchForNewLights()
}

// GetLightAttributes() is same as hue.User.GetLightAttributes() except all light ids are mapped to
// socket ids.
func (lg *LightStrand) GetLightAttributes(socketId string) (*hue.LightAttributes, error) {
	lightId, err := lg.getLightIdFromSocketId(socketId, fmt.Sprintf("/lights/%s", socketId))
	if err != nil {
		return nil, err
	}
	return lg.API.GetLightAttributes(lightId)
}

// SetLightName() is same as hue.User.SetLightName() except all light ids are mapped to
// socket ids.
func (lg *LightStrand) SetLightName(socketId string, name string) error {
	lightId, err := lg.getLightIdFromSocketId(socketId, fmt.Sprintf("/lights/%s", socketId))
	if err != nil {
		return err
	}
	return lg.API.SetLightName(lightId, name)
}

// SetLightState() is same as hue.User.SetLightState() except all light ids are mapped to
// socket ids.
func (lg *LightStrand) SetLightState(socketId string, state *hue.LightState) error {
	lightId, err := lg.getLightIdFromSocketId(socketId, fmt.Sprintf("/lights/%s/state", socketId))
	if err != nil {
		return err
	}
	return lg.API.SetLightState(lightId, state)
}

// SetGroupState() is same as hue.User.SetGroupState().
func (lg *LightStrand) SetGroupState(groupId string, state *hue.LightState) error {
	return lg.API.SetGroupState(groupId, state)
}

func (lg *LightStrand) getLightIdFromSocketId(socketId, address string) (string, error) {
	lightId := lg.Lights.GetValue(socketId)
	if lightId == "" {
		return "", createResourceNotAvailableAPIError(socketId, address)
	}

	return lightId, nil
}

func (lg *LightStrand) getSocketIdFromLightId(lightId string) (string, error) {
	socketId := lg.Lights.GetKey(lightId)
	if socketId == "" {
		return "", errors.New(fmt.Sprintf("Light Id %s is not mapped.", lightId))
	}

	return socketId, nil
}

func (lg *LightStrand) mapLightIdsToSocketIds(lights []hue.Light) []hue.Light {
	if lights == nil {
		return nil
	}

	lightsWithSocketId := make([]hue.Light, 0, len(lights))

	for _, light := range lights {
		socketId, err := lg.getSocketIdFromLightId(light.Id)
		// Skip any lights that haven't been registered with strand
		if err != nil {
			continue
		}

		light.Id = socketId
		lightsWithSocketId = append(lightsWithSocketId, light)
	}

	return lightsWithSocketId
}

func createResourceNotAvailableAPIError(resourceId, address string) error {
	apiError := &hue.APIError{}
	apiError.Errors = make([]hue.APIErrorDetail, 1, 1)
	apiError.Errors[0].Type = hue.ResourceNotAvailableErrorType
	apiError.Errors[0].Address = address
	apiError.Errors[0].Description = fmt.Sprintf("resource, %s, not available", resourceId)

	return apiError
}
