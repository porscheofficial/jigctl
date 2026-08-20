package main

type colorDecisionInputs struct {
	IsTerminal  bool
	FlagNoColor bool
	FlagPlain   bool
	LookupEnv   func(string) (string, bool)
}

func shouldEnableColor(inputs colorDecisionInputs) bool {
	if inputs.FlagPlain {
		return false
	}
	if inputs.FlagNoColor {
		return false
	}
	if !inputs.IsTerminal {
		return false
	}

	if val, ok := inputs.LookupEnv("NO_COLOR"); ok && val != "" {
		return false
	}

	if val, ok := inputs.LookupEnv("TERM"); ok && val == "dumb" {
		return false
	}

	if _, ok := inputs.LookupEnv("JIGCTL_NO_COLOR"); ok {
		return false
	}

	return true
}
