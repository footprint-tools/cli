package repo

import "github.com/Skryensya/footprint/internal/domain"

// Tracker wraps repository tracking operations and implements domain.RepoTracker.
type Tracker struct{}

// NewTracker creates a new repository tracker.
func NewTracker() *Tracker {
	return &Tracker{}
}

// DeriveID derives a repository ID from a remote URL or local path.
func (t *Tracker) DeriveID(remoteURL, localPath string) (domain.RepoID, error) {
	id, err := DeriveID(remoteURL, localPath)
	if err != nil {
		return "", err
	}
	return domain.RepoID(id), nil
}

// Track starts tracking a repository.
func (t *Tracker) Track(id domain.RepoID) (added bool, err error) {
	return Track(RepoID(id))
}

// Untrack stops tracking a repository.
func (t *Tracker) Untrack(id domain.RepoID) (removed bool, err error) {
	return Untrack(RepoID(id))
}

// IsTracked checks if a repository is being tracked.
func (t *Tracker) IsTracked(id domain.RepoID) (bool, error) {
	return IsTracked(RepoID(id))
}

// ListTracked returns all tracked repository IDs.
func (t *Tracker) ListTracked() ([]domain.RepoID, error) {
	ids, err := ListTracked()
	if err != nil {
		return nil, err
	}
	result := make([]domain.RepoID, len(ids))
	for i, id := range ids {
		result[i] = domain.RepoID(id)
	}
	return result, nil
}

// Verify Tracker implements domain.RepoTracker
var _ domain.RepoTracker = (*Tracker)(nil)
