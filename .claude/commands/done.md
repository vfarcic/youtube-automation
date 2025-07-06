# Finish Session

Complete the current work session following the 19-step Complete Session Workflow.

## Usage
```
/finish-session
```

## Description
This command implements the complete session workflow across 4 phases:

**PHASE 0: PRE-SUBMISSION VALIDATION**
- Run full test suite to ensure all tests pass
- Run performance tests if API endpoints were modified  
- Verify build succeeds

**PHASE 1: CODE SUBMISSION**
- Check git status for uncommitted changes
- Create feature branch with descriptive name
- Stage all changes and commit with comprehensive message
- Push branch to origin
- Create PR with detailed description

**PHASE 2: REVIEW & MERGE**
- Check PR reviews (owner can self-approve)
- Merge PR using squash method
- Switch to main branch
- Delete local and remote feature branches
- Pull latest changes

**PHASE 3: POST-MERGE CLEANUP**
- Close GitHub PRD Issues with completion comment
- Clean up local task files (.taskmaster/tasks/*)
- Delete local PRD files (.taskmaster/docs/prd_*.txt)

**PHASE 4: API CHANGE HANDLING**
- Check for API endpoint modifications
- Suggest creating corresponding PRD in youtube-web repository if needed

## Steps Executed
This command automatically executes all 19 steps from the Complete Session Workflow without requiring memory lookup.

## Requirements
- Git repository with uncommitted changes
- GitHub CLI (gh) configured
- Go project that can build and test
- Push access to the repository