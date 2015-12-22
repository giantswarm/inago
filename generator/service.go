package generator

import (
	"fmt"

	"github.com/giantswarm/infra-tmpl-go/parser"
)

type Service struct {
	Name      string
	After     string
	Scale     int
	Conflicts []string
	Units     []Unit `json:"units,omitempty"`
}

type Unit struct {
	// UnitType is the GS specific unit type like lb-register or ambassador.
	GSType string

	// SystemdType is the Type statement of a systemd unit file's Service
	// section. E.g. oneshot.
	SystemdType string

	Name      string
	Image     string
	Version   string
	Iptables  bool
	ExecStart []string

	TimeoutStartSec string
	TimeoutStopSec  string
	RemainAfterExit string

	MachineOf     string
	Global        bool
	FleetMetadata []string
}

func (s Service) NeighbourNames(u Unit) []string {
	names := []string{}
	prevName := s.PrevUnitName(u)
	nextName := s.NextUnitName(u)

	if prevName != "" {
		names = append(names, prevName)
	}
	if nextName != "" {
		names = append(names, nextName)
	}

	return names
}

func (s Service) PrevUnitName(u Unit) string {
	fmt.Printf(">>>> PrevUnitName u: %#v\n", u.Name)
	for i, unit := range s.Units {
		fmt.Printf(">>>> unit.Name: %#v\n", unit.Name)
		if unit.Name == u.Name {
			if i > 0 {
				return s.Units[i-1].Name
			}
		}
	}

	return ""
}

func (s Service) NextUnitName(u Unit) string {
	for i, unit := range s.Units {
		if unit.Name == u.Name {
			if len(s.Units) < i {
				return s.Units[i+1].Name
			}
		}
	}

	return ""
}

func mapParserServiceToGeneratorService(serviceTmpl parser.ServiceTmpl) Service {
	s := Service{
		Name:      serviceTmpl.Name,
		After:     serviceTmpl.After,
		Scale:     serviceTmpl.Scale,
		Conflicts: serviceTmpl.Conflicts,
	}

	for _, unitTmpl := range serviceTmpl.Units {
		u := Unit{}

		switch unitTmpl.GSType {
		case "user":
			u = mapParserUnitToUserUnit(unitTmpl)
		case "lb-register":
			u = mapParserUnitToLbRegisterUnit(unitTmpl)
		}

		s.Units = append(s.Units, u)
	}

	return s
}

func mapParserUnitToUserUnit(unitTmpl parser.UnitTmpl) Unit {
	u := Unit{
		Name:      unitTmpl.Name,
		Image:     unitTmpl.Image,
		Version:   unitTmpl.Version,
		ExecStart: unitTmpl.ExecStart,
	}

	return u
}

func mapParserUnitToLbRegisterUnit(unitTmpl parser.UnitTmpl) Unit {
	u := Unit{}

	return u
}
