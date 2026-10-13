package telemetry

// buildRules returns every rule, grouped by subject. The order matters: a
// recognized span is erased before the next rule runs, so the more specific
// rules must come before the generic ones they could collide with. Each group
// puts its own rules in that order. Between groups, one dependency: the link
// rules go before the speed rules, because the MAX glyph that stands for an
// RSSI of 100% also starts the maximum speeds.
//
// Some elements cannot be told apart from another one by glyphs, because the
// firmware draws them in the same way; the first one on the screen keeps the
// column (see rule.apply). README.md lists the pairs under Limitations.
func buildRules() []rule {
	var rules []rule
	for _, group := range [][]rule{
		linkRules(),
		timeRules(),
		batteryRules(),
		flightRules(),
		positionRules(),
		speedRules(),
		performanceRules(),
		sensorRules(),
		systemRules(),
	} {
		rules = append(rules, group...)
	}
	return rules
}
