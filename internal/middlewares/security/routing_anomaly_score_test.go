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
				numberOfJumps: tresholdForRating - 1,
				scoreSum:      100,
				avgDiviation:  10,
			},
			trackerHistory: &trackerHistory{jumpsScoreSum: 50, jumpsCount: 1},
			want:           0,
		},
		{
			name: "Score below minScore returns 0",
			detector: &RoutingAnomalyDetector{
				numberOfJumps: tresholdForRating + 1,
				scoreSum:      1000,
				avgDiviation:  100,
			},
			trackerHistory: &trackerHistory{jumpsScoreSum: 95, jumpsCount: 1},
			want:           0,
		},
		{
			name: "Score above maxScore returns 1",
			detector: &RoutingAnomalyDetector{
				numberOfJumps: tresholdForRating + 1,
				scoreSum:      1000,
				avgDiviation:  1,
			},
			trackerHistory: &trackerHistory{jumpsScoreSum: 50, jumpsCount: 1},
			want:           1,
		},
		{
			name: "Normal score calculation",
			detector: &RoutingAnomalyDetector{
				numberOfJumps: 100,
				scoreSum:      1000,
				avgDiviation:  50,
			},
			trackerHistory: &trackerHistory{jumpsScoreSum: 90, jumpsCount: 1},
			want:           0.6, // assuming avg score of 10 units
		},
		{
			name: "Zero avgDiviation handling",
			detector: &RoutingAnomalyDetector{
				numberOfJumps: tresholdForRating + 1,
				scoreSum:      100,
				avgDiviation:  0,
			},
			trackerHistory: &trackerHistory{jumpsScoreSum: 100, jumpsCount: 1},
			want:           1, // should return max score due to division by zero protection
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			initialAvgDiviation := tt.detector.avgDiviation
			got := tt.detector.calcAnomalyRating(tt.trackerHistory)

			if math.Abs(got-tt.want) > 0.0001 { // Using small epsilon for float comparison
				t.Errorf("calcAnomalyRating() = %v, want %v", got, tt.want)
			}

			// Test that avgDiviation is properly updated
			avgGlobalScore := tt.detector.scoreSum / float64(tt.detector.numberOfJumps)
			avgTrackerScore := tt.trackerHistory.jumpsScoreSum
			expectedDiviation := math.Abs(avgGlobalScore - avgTrackerScore)
			expectedAvgDiviation := ((initialAvgDiviation * float64(tt.detector.numberOfJumps-1)) + expectedDiviation) / float64(tt.detector.numberOfJumps)

			if math.Abs(tt.detector.avgDiviation-expectedAvgDiviation) > 0.0001 {
				t.Errorf("avgDiviation update incorrect: got %v, want %v", tt.detector.avgDiviation, expectedAvgDiviation)
			}
		})
	}
}

// Test helper functions
func TestAvgDiviationCalculation(t *testing.T) {
	detector := &RoutingAnomalyDetector{
		numberOfJumps: tresholdForRating + 1,
		scoreSum:      100,
		avgDiviation:  10,
	}

	history := &trackerHistory{jumpsScoreSum: 40, jumpsCount: 1}

	// First calculation
	initialAvgDiviation := detector.avgDiviation
	_ = detector.calcAnomalyRating(history)

	// Verify that the average diviation is updated correctly
	avgGlobalScore := detector.scoreSum / float64(detector.numberOfJumps)
	diviation := math.Abs(avgGlobalScore - history.jumpsScoreSum)
	expectedAvgDiviation := ((initialAvgDiviation * float64(detector.numberOfJumps-1)) + diviation) / float64(detector.numberOfJumps)

	if math.Abs(detector.avgDiviation-expectedAvgDiviation) > 0.0001 {
		t.Errorf("avgDiviation calculation incorrect: got %v, want %v", detector.avgDiviation, expectedAvgDiviation)
	}
}
