package server

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/BrianBland/go-hue"
	"github.com/BrianBland/palette"
)

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
		return uint16Ptr(0)
	}
	return r.Hue
}

func (r request) saturation() *uint8 {
	if r.Saturation == nil {
		return uint8Ptr(2<<7 - 1)
	}
	return r.Saturation
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
		Effect:     req.Effect,
	}
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
