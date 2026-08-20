package runner

// ExitSummary captures the final evaluated state of a binding for exit code aggregation.
type ExitSummary struct {
	Projection Projection
	IsBlocking bool
}

type exitState struct {
	operational       bool
	blockedUnchecked  bool
	blockingViol      bool
	expectedUnchecked bool
	realResult        bool
}

// AggregateExitCode computes the final process exit code per ADR-0012.
func AggregateExitCode(verdicts []ExitSummary, strict bool) int {
	if len(verdicts) == 0 {
		return 77
	}

	var st exitState
	for _, v := range verdicts {
		updateState(&st, v)
	}

	if st.operational {
		return 2
	}
	if st.blockingViol || st.blockedUnchecked {
		return 1
	}
	if strict && st.expectedUnchecked {
		return 1
	}
	if st.realResult {
		return 0
	}
	return 77
}

func updateState(st *exitState, v ExitSummary) {
	switch v.Projection {
	case ProjectionOperational:
		st.operational = true
	case ProjectionBlockedUnchecked:
		st.blockedUnchecked = true
	case ProjectionViolation:
		if v.IsBlocking {
			st.blockingViol = true
		} else {
			st.realResult = true
		}
	case ProjectionExpectedUnchecked:
		st.expectedUnchecked = true
	case ProjectionPass:
		st.realResult = true
	case ProjectionInvalid:
	}
}
