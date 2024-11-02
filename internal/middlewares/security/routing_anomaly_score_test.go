package security

import (
	"math"
	"testing"
)

func TestCalcAnomalyRating(t *testing.T) {
	tests := []struct {
		name           string
		detector       *RoutingAnomalyDetector
		trackerHistory *trackerHistory
		want           float64
	}{
		{
			name: "Below threshold returns 0",
			detector: &RoutingAnomalyDetector{
				numberOfJumps:     99,
				scoreSum:          100,
				avgDiviation:      10,
				tresholdForRating: 100,
				minScore:          100,
				maxScore:          200,
				anomalyHeaderName: "test",
			},
			trackerHistory: &trackerHistory{jumpsScoreSum: 50, jumpsCount: 1},
			want:           0,
		},
		{
			name: "Score below minScore returns 0",
			detector: &RoutingAnomalyDetector{
				numberOfJumps:     101,
				scoreSum:          500,
				avgDiviation:      100,
				tresholdForRating: 100,
				minScore:          100,
				maxScore:          200,
				anomalyHeaderName: "test",
			},
			trackerHistory: &trackerHistory{jumpsScoreSum: 95, jumpsCount: 1},
			want:           0,
		},
		{
			name: "Score above maxScore returns 1",
			detector: &RoutingAnomalyDetector{
				numberOfJumps:     101,
				scoreSum:          1000,
				avgDiviation:      1,
				tresholdForRating: 100,
				minScore:          100,
				maxScore:          200,
				anomalyHeaderName: "test",
			},
			trackerHistory: &trackerHistory{jumpsScoreSum: 50, jumpsCount: 1},
			want:           1,
		},
		{
			name: "Normal score calculation",
			detector: &RoutingAnomalyDetector{
				numberOfJumps:     100,
				scoreSum:          500,
				avgDiviation:      50,
				tresholdForRating: 100,
				minScore:          100,
				maxScore:          200,
				anomalyHeaderName: "test",
			},
			trackerHistory: &trackerHistory{jumpsScoreSum: 90, jumpsCount: 1},
			want:           0.6, // assuming avg score of 10 units
		},
		{
			name: "Zero avgDiviation handling",
			detector: &RoutingAnomalyDetector{
				numberOfJumps:     101,
				scoreSum:          100,
				avgDiviation:      0,
				tresholdForRating: 100,
				minScore:          100,
				maxScore:          200,
				anomalyHeaderName: "test",
			},
			trackerHistory: &trackerHistory{jumpsScoreSum: 100, jumpsCount: 1},
			want:           1, // should return max score due to division by zero protection
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.detector.calcAnomalyRating(tt.trackerHistory)

			if math.Abs(got-tt.want) > 0.0001 { // Using small epsilon for float comparison
				t.Errorf("calcAnomalyRating() = %v, want %v", got, tt.want)
			}
		})
	}
}
