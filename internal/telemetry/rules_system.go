package telemetry

import "github.com/ericogr/fpv-osd-decoder/internal/inav"

// gvarKeys is in the order of the variables.
var gvarKeys = [...]string{keyGVar0, keyGVar1, keyGVar2, keyGVar3}

// systemRules read the state of the flight controller: the configuration
// profile, the blackbox log and the global variables of the programming
// framework.
func systemRules() []rule {
	return []rule{
		newRule(rx(sym(inav.SymProfile), spaceRE, `([0-9]+)`), intField(keyProfile, 1)),
		newRule(rx(sym(inav.SymBlackbox), `([0-9]{5})`), intField(keyBlackboxLog, 1)),
		// "G0:" to "G3:" and a number
		newRule(rx(`\bG([0-3]):`, spaceRE, numRE), globalVariable),
	}
}

// globalVariable reads the number in group 2 into the variable named in group
// 1, which the expression limits to a digit from 0 to 3.
func globalVariable(g []string, s Sample) {
	if v, ok := parseNum(g[2]); ok {
		s.num(gvarKeys[g[1][0]-'0'], v)
	}
}
