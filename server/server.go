package server

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"strings"

	"github.com/BrianBland/palette"

	"github.com/BrianBland/go-hue"
	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
)

func init() {
	log.SetLevel(log.DebugLevel)
}

type Server struct {
	palette *palette.Palette
}

func New(p *palette.Palette) *Server {
	return &Server{palette: p}
}

func (s *Server) ListenAndServe(addr string) error {
	log.Printf("Listening on %s...", addr)
	return (&http.Server{
		Addr:    addr,
		Handler: s.Handler(),
	}).ListenAndServe()
}

type request struct {
	Palette    string  `json:"palette"`
	Brightness *uint8  `json:"brightness"`
	Color      string  `json:"color"`
	Hue        *uint16 `json:"hue"`
	Saturation *uint8  `json:"saturation"`
	Alert      string  `json:"alert"`
	Effect     string  `json:"effect"`
}

func (r request) brightness() *uint8 {
	if r.Brightness == nil {
		return uint8Ptr(2<<7 - 1)
	}
	return r.Brightness
}

func (r request) hue() *uint16 {
	if r.Hue == nil {
		log.WithField("color", r.Color).Debug("No hue provided, using color instead")
		switch strings.ToLower(r.Color) {
		case "r", "red":
			return uint16Ptr(0)
		case "o", "orange":
			return uint16Ptr(hueFromDegrees(30))
		case "y", "yellow":
			return uint16Ptr(hueFromDegrees(60))
		case "g", "green":
			return uint16Ptr(hueFromDegrees(120))
		case "c", "cyan":
			return uint16Ptr(hueFromDegrees(180))
		case "b", "blue":
			return uint16Ptr(hueFromDegrees(240))
		case "i", "indigo":
			return uint16Ptr(hueFromDegrees(260))
		case "v", "violet":
			return uint16Ptr(hueFromDegrees(270))
		case "m", "magenta", "p", "purple":
			return uint16Ptr(hueFromDegrees(300))
		default:
			return uint16Ptr(randomHue())
		}
	}
	return r.Hue
}

func hueFromDegrees(d float64) uint16 {
	h := (uint16)((2 << 15) * (d / 360.0))
	log.WithFields(log.Fields{
		"degrees": d,
		"hue":     h,
	}).Debug("Hue from degrees")
	return h
}

func randomHue() uint16 {
	h := (uint16)(rand.Int31n(2 << 15))
	log.WithField("hue", h).Debug("Random hue")
	return h
}

func (r request) saturation() *uint8 {
	if r.Saturation == nil {
		log.Debug("No saturation provided, defaulting to max")
		return uint8Ptr(2<<7 - 1)
	}
	return r.Saturation
}

func (r request) effect() string {
	if r.Effect == "" {
		log.Debug("No effect provided, defaulting to none")
		return "none"
	}
	return r.Effect
}

func (s *Server) Handler() http.Handler {
	r := mux.NewRouter()
	r.StrictSlash(true)
	r.Handle("/", http.FileServer(http.Dir("static")))
	r.HandleFunc("/lights", s.getLights).Methods("GET")
	r.HandleFunc("/palette", s.setPalette).Methods("PUT", "POST")
	r.HandleFunc("/on", s.lightsOn).Methods("PUT", "POST")
	r.HandleFunc("/off", s.lightsOut).Methods("PUT", "POST")
	return r
}

func (s *Server) getLights(rw http.ResponseWriter, r *http.Request) {
	lights, err := s.palette.GetLights()
	if err != nil {
		http.Error(rw, http.StatusText(http.StatusBadGateway), http.StatusBadGateway)
		return
	}

	lightStates := make([]hue.LightState, 0)
	ch := s.palette.GetGroup(lights)
	for attrsOrErr := range ch {
		if attrsOrErr.Error != nil {
			err = attrsOrErr.Error
		} else {
			state := attrsOrErr.State
			if state != nil {
				lightStates = append(lightStates, *state)
			}
		}
	}

	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(rw, struct {
		Lights []hue.LightState `json:"lights"`
	}{
		Lights: lightStates,
	})
}

func (s *Server) setPalette(rw http.ResponseWriter, r *http.Request) {
	var req request
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(rw, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	lights, err := s.palette.GetLights()
	if err != nil {
		http.Error(rw, http.StatusText(http.StatusBadGateway), http.StatusBadGateway)
		return
	}
	state := hue.LightState{
		On:         boolPtr(true),
		Brightness: req.brightness(),
		Hue:        req.hue(),
		Saturation: req.saturation(),
		Alert:      req.Alert,
		Effect:     req.effect(),
	}
	log.WithFields(log.Fields{
		"palette":      req.Palette,
		"primaryState": state,
	}).Debug("Setting light state")
	var errChan <-chan error
	switch strings.ToLower(req.Palette) {
	case "complementary":
		errChan = s.palette.SetComplementary(lights, state)
	case "triad":
		errChan = s.palette.SetTriad(lights, state)
	case "analogous", "adjacent":
		errChan = s.palette.SetAnalogous(lights, state)
	case "split", "splitcomplementary":
		errChan = s.palette.SetSplitComplementary(lights, state)
	case "rectangle":
		errChan = s.palette.SetRectangle(lights, state)
	case "square":
		errChan = s.palette.SetSquare(lights, state)
	default:
		http.Error(rw, "Invalid palette", http.StatusBadRequest)
		return
	}
	err = handleErrChan(rw, errChan)
	if err == nil {
		s.getLights(rw, r)
	}
}

func (s *Server) lightsOn(rw http.ResponseWriter, r *http.Request) {
	log.Debug("Lights on!")
	lights, err := s.palette.GetLights()
	if err != nil {
		http.Error(rw, http.StatusText(http.StatusBadGateway), http.StatusBadGateway)
		return
	}
	state := hue.LightState{On: boolPtr(true)}
	errChan := s.palette.SetGroup(lights, []hue.LightState{state})
	handleErrChan(rw, errChan)
	if err == nil {
		s.getLights(rw, r)
	}
}

func (s *Server) lightsOut(rw http.ResponseWriter, r *http.Request) {
	log.Debug("Lights out!")
	lights, err := s.palette.GetLights()
	if err != nil {
		http.Error(rw, http.StatusText(http.StatusBadGateway), http.StatusBadGateway)
		return
	}
	state := hue.LightState{On: boolPtr(false)}
	errChan := s.palette.SetGroup(lights, []hue.LightState{state})
	handleErrChan(rw, errChan)
	if err == nil {
		s.getLights(rw, r)
	}
}

func handleErrChan(rw http.ResponseWriter, errChan <-chan error) error {
	var err error
	for errResponse := range errChan {
		if errResponse != nil {
			err = errResponse
		}
	}
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
	}
	return err
}

func writeJSON(rw http.ResponseWriter, payload interface{}) error {
	return writeJSONStatus(rw, payload, http.StatusOK)
}

func writeJSONStatus(rw http.ResponseWriter, payload interface{}, status int) error {
	marshalled, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return err
	}
	rw.Header().Set("Content-Type", "application/json; charset=utf-8")
	rw.WriteHeader(status)
	fmt.Fprint(rw, string(marshalled))
	return nil
}
