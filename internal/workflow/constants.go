package workflow

// Video phase constants that define the lifecycle stages of a video
const (
	PhasePublished        = 0
	PhasePublishPending   = 1
	PhaseEditRequested    = 2
	PhaseMaterialDone     = 3
	PhaseStarted          = 4
	PhaseDelayed          = 5
	PhaseSponsoredBlocked = 6
	PhaseIdeas            = 7
)

// PhaseNames maps phase constants to human-readable names
var PhaseNames = map[int]string{
	PhasePublished:        "Published",
	PhasePublishPending:   "Publish Pending", 
	PhaseEditRequested:    "Edit Requested",
	PhaseMaterialDone:     "Material Done",
	PhaseStarted:          "Started",
	PhaseDelayed:          "Delayed",
	PhaseSponsoredBlocked: "Sponsored Blocked",
	PhaseIdeas:            "Ideas",
}