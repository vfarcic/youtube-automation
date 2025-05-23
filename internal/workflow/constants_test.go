package workflow

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPhaseConstants(t *testing.T) {
	assert.Equal(t, 0, PhasePublished, "PhasePublished should be 0")
	assert.Equal(t, 1, PhasePublishPending, "PhasePublishPending should be 1")
	assert.Equal(t, 2, PhaseEditRequested, "PhaseEditRequested should be 2")
	assert.Equal(t, 3, PhaseMaterialDone, "PhaseMaterialDone should be 3")
	assert.Equal(t, 4, PhaseStarted, "PhaseStarted should be 4")
	assert.Equal(t, 5, PhaseDelayed, "PhaseDelayed should be 5")
	assert.Equal(t, 6, PhaseSponsoredBlocked, "PhaseSponsoredBlocked should be 6")
	assert.Equal(t, 7, PhaseIdeas, "PhaseIdeas should be 7")
}

func TestPhaseNames(t *testing.T) {
	expectedPhaseNames := map[int]string{
		PhasePublished:        "Published",
		PhasePublishPending:   "Publish Pending",
		PhaseEditRequested:    "Edit Requested",
		PhaseMaterialDone:     "Material Done",
		PhaseStarted:          "Started",
		PhaseDelayed:          "Delayed",
		PhaseSponsoredBlocked: "Sponsored Blocked",
		PhaseIdeas:            "Ideas",
	}

	assert.Equal(t, expectedPhaseNames, PhaseNames, "PhaseNames map should match expected values")

	// Additionally, check if all defined constants have a corresponding name
	allConstants := []int{
		PhasePublished,
		PhasePublishPending,
		PhaseEditRequested,
		PhaseMaterialDone,
		PhaseStarted,
		PhaseDelayed,
		PhaseSponsoredBlocked,
		PhaseIdeas,
	}
	for _, phaseConst := range allConstants {
		_, ok := PhaseNames[phaseConst]
		assert.True(t, ok, "Phase constant %d should have a name in PhaseNames", phaseConst)
	}

	// And check if PhaseNames contains any extra keys not in constants
	assert.Equal(t, len(allConstants), len(PhaseNames), "PhaseNames should only contain keys corresponding to defined phase constants")
}
