package base_swagger_service

type StrategyError interface {
	setError(checks []CheckerResult) bool
}

type StrategyErrorOne struct{}

// setError if one of services has problem
func (s StrategyErrorOne) setError(checks []CheckerResult) bool {
	for _, check := range checks {
		if check.Status == false {
			return true
		}
	}

	return false
}

type StrategyErrorAll struct{}

// setError if all services has problems
func (s StrategyErrorAll) setError(checks []CheckerResult) bool {
	for _, check := range checks {
		if check.Status == true {
			return false
		}
	}

	return true
}

type StrategyErrorIgnore struct{}

// setError ignore all problems
func (s StrategyErrorIgnore) setError(checks []CheckerResult) bool {
	return false
}
