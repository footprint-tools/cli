package store

// UpdateCache holds cached update check info.
type UpdateCache struct {
	LastCheck     string
	LatestVersion string
}

// GetUpdateCache retrieves the cached update check info.
func (s *Store) GetUpdateCache() (UpdateCache, error) {
	var cache UpdateCache
	err := s.db.QueryRow(`
		SELECT COALESCE(update_last_check, ''), COALESCE(update_latest_version, '')
		FROM state WHERE id = 1
	`).Scan(&cache.LastCheck, &cache.LatestVersion)
	if err != nil {
		return UpdateCache{}, err
	}
	return cache, nil
}

// SetUpdateCache stores the update check info.
func (s *Store) SetUpdateCache(lastCheck, latestVersion string) error {
	_, err := s.db.Exec(`
		UPDATE state SET update_last_check = ?, update_latest_version = ? WHERE id = 1
	`, lastCheck, latestVersion)
	return err
}
