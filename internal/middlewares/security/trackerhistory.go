package security

type trackerHistory struct {
	jumpsCount    int
	jumpsScoreSum float64
}

func (th trackerHistory) Avg() float64 {
	return th.jumpsScoreSum / float64(th.jumpsCount)
}
