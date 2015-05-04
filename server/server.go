package server

import (
	"encoding/json"
	"math/rand"
	"net/http"
	"strings"

	"github.com/BrianBland/palette"

	"github.com/BrianBland/go-hue"
	log "github.com/Sirupsen/logrus"
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
	h := http.NewServeMux()
	h.HandleFunc("/palette", s.setPalette)
	h.HandleFunc("/off", s.lightsOut)
	return h
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
	switch strings.ToLower(req.Palette) {
	case "complementary":
		err = s.palette.SetComplementary(lights, state)
	case "triad":
		err = s.palette.SetTriad(lights, state)
	case "analogous", "adjacent":
		err = s.palette.SetAnalogous(lights, state)
	case "split", "splitcomplementary":
		err = s.palette.SetSplitComplementary(lights, state)
	case "rectangle":
		err = s.palette.SetRectangle(lights, state)
	case "square":
		err = s.palette.SetSquare(lights, state)
	default:
		http.Error(rw, "Invalid palette", http.StatusBadRequest)
		return
	}
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
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
	err = s.palette.SetGroup(lights, []hue.LightState{state})
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}
}
