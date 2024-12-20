package security

import (
	"math"
	"net/http"
	"net/url"
	"strconv"
	"sync"

	"github.com/hvuhsg/gatego/pkg/pathgraph"
	"github.com/hvuhsg/gatego/pkg/tracker"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

const (
	tracingCookieName = "sad-trc"
	cookieMaxAge      = 24 * 60 * 60 // 24 hours in seconds
	refererHeaderName = "Referer"
)

// RoutingAnomalyDetector handles path tracking logic and manages user sessions
type RoutingAnomalyDetector struct {
	graph                 *pathgraph.PathGraph
	numberOfJumps         int
	scoreSum              float64
	avgDiviation          float64
	lastPaths             sync.Map // Maps trace_id to last path
	trackerRoutingHistory sync.Map
	tracker               tracker.Tracker

	tresholdForRating int // The number of requests before starting to calculate anomaly score
	minScore          int // If the diviation form the avg diviation is lower then this then the session is not suspicuse
	maxScore          int // If the diviation form the avg diviation is larger then this then the session is fully suspicuse
	anomalyHeaderName string
}

func NewRoutingAnomalyDetector(headerName string, tresholdForRating, minScore, maxScore int) *RoutingAnomalyDetector {
	return &RoutingAnomalyDetector{
		graph:             pathgraph.NewPathGraph(),
		tracker:           tracker.NewCookieTracker(tracingCookieName, cookieMaxAge, false),
		anomalyHeaderName: headerName,
		minScore:          minScore,
		maxScore:          maxScore,
		tresholdForRating: tresholdForRating,
	}
}

// NewPathTracker creates a new PathTracker instance
func NewPathTracker(graph *pathgraph.PathGraph) *RoutingAnomalyDetector {
	return &RoutingAnomalyDetector{
		graph:     graph,
		lastPaths: sync.Map{},
	}
}

// Claculate anomaly score based on global avg routing and tracker routing
// This middleware uses a graph to represent every path called by users
// Eeach source, destination path has a vertex with the score of how many requests jumpt it,
// We save tracker (session) jumps history and calculate an anomaly score, and add it as header to the request.
func (pt *RoutingAnomalyDetector) AddAnomalyScore(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		span := trace.SpanFromContext(r.Context())

		// Get or create trace ID
		traceID := pt.tracker.GetTrackerID(r)

		if traceID != "" {
			// We do not want the tracker to be sent to the downstream server
			pt.tracker.RemoveTracker(r)
		} else { // Create new tracker if not found
			var err error
			traceID, err = pt.tracker.SetTracker(w)
			if err != nil {
				// Log error but continue serving
				next.ServeHTTP(w, r)
				return
			}

			// Create tracker history
			trackerH := &trackerHistory{jumpsCount: 0, jumpsScoreSum: 0}
			pt.trackerRoutingHistory.Store(traceID, trackerH)
		}

		currentPath := r.URL.Path

		// Get last path for this trace ID
		lastPath, exists := pt.getLastPath(traceID, r)
		if !exists {
			lastPath = "" // empty path means the user has entered the site for the first time
		}

		jumpScore := pt.graph.AddJump(lastPath, currentPath)
		value, ok := pt.trackerRoutingHistory.Load(traceID)

		var trackerH *trackerHistory
		if ok {
			trackerH = value.(*trackerHistory)
		}

		// update tracker history with jump score
		trackerH.jumpsCount++
		trackerH.jumpsScoreSum += jumpScore

		// update global stats
		pt.numberOfJumps++
		pt.scoreSum += jumpScore

		pt.lastPaths.Store(traceID, currentPath)

		anomalyScore := pt.calcAnomalyRating(trackerH)
		span.SetAttributes(attribute.Float64("RoutingAnomalyScore", anomalyScore))

		r.Header.Set(pt.anomalyHeaderName, strconv.FormatFloat(anomalyScore, 'f', 2, 64))

		// Call the next handler
		next.ServeHTTP(w, r)
	})
}

// GetLastPath retrieves the last path for a given trace ID from storage or referer header (in this order)
func (pt *RoutingAnomalyDetector) getLastPath(traceID string, r *http.Request) (string, bool) {
	path, exists := pt.lastPaths.Load(traceID)

	if !exists {
		u := r.Header.Get(refererHeaderName)
		url, err := url.Parse(u)
		if err == nil {
			path = url.Path
		}
	}

	return path.(string), exists
}

// 0 - is fully normal, 1 - fully suspicuse
func (pt *RoutingAnomalyDetector) calcAnomalyRating(trackerH *trackerHistory) float64 {
	avgGlobalScore := (pt.scoreSum / float64(pt.numberOfJumps)) * 2
	avgTrackerScore := trackerH.Avg()

	diviation := math.Abs(avgGlobalScore - avgTrackerScore)

	// If avg diviation is 0 it will return +Inf and get the correct result
	anomalyScore := (diviation / (pt.avgDiviation / 100))

	// Update avgDiviation with new diviation
	pt.avgDiviation = ((pt.avgDiviation * float64(pt.numberOfJumps)) + diviation) / float64(pt.numberOfJumps)

	// Only return 0 until useage data is collected
	if pt.numberOfJumps < pt.tresholdForRating {
		return 0
	}

	if anomalyScore < float64(pt.minScore) {
		return 0
	}

	if anomalyScore > float64(pt.maxScore) {
		return 1
	}

	return (anomalyScore - float64(pt.minScore)) / 100
}
