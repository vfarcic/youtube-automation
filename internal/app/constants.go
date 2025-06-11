package app

// Phase message constants for consistent messaging
const (
	// Edit cancelled messages
	MessageInitialDetailsEditCancelled = "Initial details edit cancelled."
	MessageWorkProgressEditCancelled   = "Work progress edit cancelled."
	MessagePostProductionEditCancelled = "Post-production edit cancelled."
	MessageDefinitionPhaseAborted      = "Definition phase aborted."

	// Error messages
	ErrorRunInitialDetailsForm     = "failed to run initial details edit form"
	ErrorRunWorkProgressForm       = "failed to run work progress edit form"
	ErrorRunPostProductionForm     = "failed to run post-production edit form"
	ErrorSaveInitialDetails        = "failed to save initial details"
	ErrorSaveWorkProgress          = "failed to save work progress"
	ErrorSavePostProductionDetails = "failed to save post-production details"
	ErrorDefinitionPhase           = "error during definition phase"

	// Success messages
	MessageInitialDetailsUpdated = "initial details updated"
	MessageWorkProgressUpdated   = "work progress updated"
	MessagePostProductionUpdated = "post-production details updated"

	// Changes not saved messages
	MessageChangesNotSavedInitialDetails = "Changes not saved for initial details."
	MessageChangesNotSavedWorkProgress   = "Changes not saved for work progress."
	MessageChangesNotSavedPostProduction = "Changes not saved for post-production."

	// Other messages
	MessageDefinitionPhaseComplete = "--- Definition Phase Complete ---"
)
