package generator

import (
	"fmt"
	"strings"
)

type Statement struct {
	Key string
	Val interface{}
}

type Sections struct {
	Unit    []Statement
	Service []Statement
	Timer   []Statement
	Install []Statement
	XFleet  []Statement
}

func NewSections(s Service, u Unit) Sections {
	sections := Sections{
		Unit:    UnitSectionStatements(s, u),
		Service: ServiceSectionStatements(s, u),
		XFleet:  XFleetSectionStatements(s, u),
	}

	return sections
}

func UnitSectionStatements(s Service, u Unit) []Statement {
	statements := []Statement{}

	// Description
	statement := Statement{
		Key: "Description",
		Val: fmt.Sprintf("s:%s t:%s u:%s", s.Name, u.GSType, u.Name),
	}
	statements = append(statements, statement)

	// Wants
	if s.After != "" {
		statement := Statement{
			Key: "Wants",
			Val: s.After,
		}
		statements = append(statements, statement)
	}

	// Requires
	for _, n := range s.NeighbourNames(u) {
		statement := Statement{
			Key: "Requires",
			Val: n,
		}
		statements = append(statements, statement)
	}

	// After
	if s.After != "" {
		statement := Statement{
			Key: "Afer",
			Val: s.After,
		}
		statements = append(statements, statement)
	}
	for _, n := range s.NeighbourNames(u) {
		statement := Statement{
			Key: "After",
			Val: n,
		}
		statements = append(statements, statement)
	}

	return statements
}

func ServiceSectionStatements(s Service, u Unit) []Statement {
	statements := []Statement{}

	// Type
	if u.SystemdType != "" {
		statement := Statement{
			Key: "Type",
			Val: u.SystemdType,
		}
		statements = append(statements, statement)
	}

	// Restart policy
	if u.SystemdType != "oneshot" {
		statement := Statement{
			Key: "Restart",
			Val: "on-failure",
		}
		statements = append(statements, statement)

		statement = Statement{
			Key: "RestartSec",
			Val: "1",
		}
		statements = append(statements, statement)

		statement = Statement{
			Key: "StartLimitInterval",
			Val: "300s",
		}
		statements = append(statements, statement)

		statement = Statement{
			Key: "StartLimitBurst",
			Val: "3",
		}
		statements = append(statements, statement)
	}

	// TimeoutStartSec
	if u.TimeoutStartSec != "" {
		statement := Statement{
			Key: "TimeoutStartSec",
			Val: u.TimeoutStartSec,
		}
		statements = append(statements, statement)
	}

	// RemainAfterExit
	if u.RemainAfterExit != "" {
		statement := Statement{
			Key: "RemainAfterExit",
			Val: u.RemainAfterExit,
		}
		statements = append(statements, statement)
	}

	// EnvironmentFile
	statement := Statement{
		Key: "EnvironmentFile",
		Val: "/etc/environment",
	}
	statements = append(statements, statement)

	// Name
	if u.Name != "" {
		statement := Statement{
			Key: "Environment",
			Val: fmt.Sprintf("\"NAME=%s\"", u.Name),
		}
		statements = append(statements, statement)
	}

	// Image
	if u.Image != "" {
		statement := Statement{
			Key: "Environment",
			Val: fmt.Sprintf("\"IMAGE=%s\"", u.Image),
		}
		statements = append(statements, statement)

		statement = Statement{
			Key: "ExecStartPre",
			Val: "/usr/bin/docker pull $IMAGE",
		}
		statements = append(statements, statement)

		if u.TimeoutStopSec != "" {
			statement = Statement{
				Key: "ExecStartPre",
				Val: fmt.Sprintf("-/usr/bin/docker stop -t %s", u.TimeoutStopSec),
			}
			statements = append(statements, statement)

			statement = Statement{
				Key: "ExecStop",
				Val: fmt.Sprintf("-/usr/bin/docker stop -t %s", u.TimeoutStopSec),
			}
			statements = append(statements, statement)
		}

		statement = Statement{
			Key: "ExecStartPre",
			Val: "-/usr/bin/docker rm -f $NAME",
		}
		statements = append(statements, statement)

		statement = Statement{
			Key: "ExecStopPost",
			Val: "-/usr/bin/docker rm -f $NAME",
		}
		statements = append(statements, statement)
	}

	// ExecStart
	if len(u.ExecStart) > 0 {
		statement := Statement{
			Key: "ExecStart",
			Val: strings.Join(u.ExecStart, " "),
		}
		statements = append(statements, statement)
	}

	// Iptables
	if u.Iptables {
		statement := Statement{
			Key: "ExecStartPost",
			Val: "/home/core/setup_iptables_rules.sh $NAME",
		}
		statements = append(statements, statement)

		statement = Statement{
			Key: "ExecStop",
			Val: "/home/core/teardown_iptables_rules.sh $NAME",
		}
		statements = append(statements, statement)
	}

	return statements
}

func XFleetSectionStatements(s Service, u Unit) []Statement {
	statements := []Statement{}

	// Global
	if u.Global {
		statement := Statement{
			Key: "Global",
			Val: "true",
		}
		statements = append(statements, statement)
	} else {
		if u.MachineOf != "" {
			statement := Statement{
				Key: "MachineOf",
				Val: u.MachineOf,
			}
			statements = append(statements, statement)
		} else if s.PrevUnitName(u) != "" {
			statement := Statement{
				Key: "MachineOf",
				Val: s.PrevUnitName(u),
			}
			statements = append(statements, statement)
		}
	}

	// Conflicts
	if s.PrevUnitName(u) == "" {
		// only the first unit in the chain defines that
		for _, c := range s.Conflicts {
			statement := Statement{
				Key: "Conflicts",
				Val: c,
			}
			statements = append(statements, statement)
		}
	}

	for _, fm := range u.FleetMetadata {
		statement := Statement{
			Key: "MachineMetadata",
			Val: fm,
		}
		statements = append(statements, statement)
	}

	return statements
}
