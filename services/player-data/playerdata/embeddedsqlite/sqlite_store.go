//go:build !noembeddedsqlite

package embeddedsqlite

import (
	"database/sql"
	"errors"
	"os"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite"

	"github.com/Lokee86/space-rocks/player-data/playerdata"
	"github.com/Lokee86/space-rocks/player-data/protocol"
)

type Config struct {
	Path string
}

type Store struct {
	db *sql.DB
}

func New(config Config) (*Store, error) {
	if config.Path == "" {
		return nil, errors.New("path is required")
	}
	if config.Path != ":memory:" {
		parentDir := filepath.Dir(config.Path)
		if parentDir != "." {
			if err := os.MkdirAll(parentDir, 0755); err != nil {
				return nil, err
			}
		}
	}

	db, err := sql.Open("sqlite", config.Path)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, err
	}

	return &Store{db: db}, nil
}

func (s *Store) Close() error {
	if s == nil || s.db == nil {
		return nil
	}
	return s.db.Close()
}

func (s *Store) InitSchema() error {
	if s == nil || s.db == nil {
		return errors.New("sqlite store is not open")
	}

	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer func() {
		_ = tx.Rollback()
	}()

	statements := []string{
		`CREATE TABLE IF NOT EXISTS local_profiles (
			local_profile_id TEXT PRIMARY KEY,
			display_name TEXT NOT NULL DEFAULT '',
			created_at TEXT NOT NULL,
			updated_at TEXT NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS local_profile_default (
			id INTEGER PRIMARY KEY CHECK (id = 1),
			identity_kind TEXT NOT NULL,
			local_profile_id TEXT NOT NULL DEFAULT '',
			updated_at TEXT NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS local_player_stats (
			local_profile_id TEXT PRIMARY KEY,
			total_score INTEGER NOT NULL DEFAULT 0,
			high_score INTEGER NOT NULL DEFAULT 0,
			ship_deaths INTEGER NOT NULL DEFAULT 0,
			games_played INTEGER NOT NULL DEFAULT 0,
			created_at TEXT NOT NULL,
			updated_at TEXT NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS local_player_match_results (
			result_id TEXT PRIMARY KEY,
			match_id TEXT NOT NULL,
			local_profile_id TEXT NOT NULL,
			score INTEGER NOT NULL DEFAULT 0,
			ship_deaths INTEGER NOT NULL DEFAULT 0,
			created_at TEXT NOT NULL
		)`,
	}

	for _, statement := range statements {
		if _, err := tx.Exec(statement); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (s *Store) ensureLocalProfile(localProfileID string) error {
	if s == nil || s.db == nil {
		return errors.New("sqlite store is not open")
	}
	if localProfileID == "" {
		return errors.New("local_profile_id is required")
	}

	now := time.Now().UTC().Format(time.RFC3339)

	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer func() {
		_ = tx.Rollback()
	}()

	if _, err := tx.Exec(
		`INSERT INTO local_profiles (local_profile_id, created_at, updated_at)
		 VALUES (?, ?, ?)
		 ON CONFLICT(local_profile_id) DO NOTHING`,
		localProfileID, now, now,
	); err != nil {
		return err
	}

	if _, err := tx.Exec(
		`INSERT INTO local_player_stats (local_profile_id, total_score, high_score, ship_deaths, games_played, created_at, updated_at)
		 VALUES (?, 0, 0, 0, 0, ?, ?)
		 ON CONFLICT(local_profile_id) DO NOTHING`,
		localProfileID, now, now,
	); err != nil {
		return err
	}

	return tx.Commit()
}

func (s *Store) ListLocalProfiles() ([]playerdata.LocalProfileSummary, error) {
	if s == nil || s.db == nil {
		return nil, errors.New("sqlite store is not open")
	}

	rows, err := s.db.Query(
		`SELECT local_profile_id, display_name
		 FROM local_profiles
		 ORDER BY created_at ASC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	profiles := make([]playerdata.LocalProfileSummary, 0)
	for rows.Next() {
		var profile playerdata.LocalProfileSummary
		if err := rows.Scan(&profile.LocalProfileID, &profile.DisplayName); err != nil {
			return nil, err
		}
		profiles = append(profiles, profile)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return profiles, nil
}

func (s *Store) GetDefaultLocalProfile() (playerdata.LocalProfileDefault, error) {
	if s == nil || s.db == nil {
		return playerdata.LocalProfileDefault{}, errors.New("sqlite store is not open")
	}

	var defaultProfile playerdata.LocalProfileDefault
	err := s.db.QueryRow(
		`SELECT identity_kind, local_profile_id
		 FROM local_profile_default
		 WHERE id = 1`,
	).Scan(&defaultProfile.IdentityKind, &defaultProfile.LocalProfileID)
	if errors.Is(err, sql.ErrNoRows) {
		return guestLocalProfileDefault(), nil
	}
	if err != nil {
		return playerdata.LocalProfileDefault{}, err
	}

	switch defaultProfile.IdentityKind {
	case playerdata.IdentityKindGuest:
		return guestLocalProfileDefault(), nil
	case playerdata.IdentityKindLocalProfile:
		if defaultProfile.LocalProfileID == "" {
			return guestLocalProfileDefault(), nil
		}

		err = s.db.QueryRow(
			`SELECT display_name
			 FROM local_profiles
			 WHERE local_profile_id = ?`,
			defaultProfile.LocalProfileID,
		).Scan(&defaultProfile.DisplayName)
		if errors.Is(err, sql.ErrNoRows) {
			return guestLocalProfileDefault(), nil
		}
		if err != nil {
			return playerdata.LocalProfileDefault{}, err
		}

		return defaultProfile, nil
	default:
		return guestLocalProfileDefault(), nil
	}
}

func (s *Store) SetDefaultLocalProfile(identityKind string, localProfileID string) (playerdata.LocalProfileDefault, error) {
	if s == nil || s.db == nil {
		return playerdata.LocalProfileDefault{}, errors.New("sqlite store is not open")
	}

	now := time.Now().UTC().Format(time.RFC3339)

	switch identityKind {
	case playerdata.IdentityKindGuest:
		tx, err := s.db.Begin()
		if err != nil {
			return playerdata.LocalProfileDefault{}, err
		}
		defer func() {
			_ = tx.Rollback()
		}()

		if _, err := tx.Exec(
			`INSERT INTO local_profile_default (id, identity_kind, local_profile_id, updated_at)
			 VALUES (1, ?, ?, ?)
			 ON CONFLICT(id) DO UPDATE SET
				identity_kind = excluded.identity_kind,
				local_profile_id = excluded.local_profile_id,
				updated_at = excluded.updated_at`,
			playerdata.IdentityKindGuest, "", now,
		); err != nil {
			return playerdata.LocalProfileDefault{}, err
		}

		if err := tx.Commit(); err != nil {
			return playerdata.LocalProfileDefault{}, err
		}

		return guestLocalProfileDefault(), nil
	case playerdata.IdentityKindLocalProfile:
		if localProfileID == "" {
			return playerdata.LocalProfileDefault{}, errors.New("local_profile_id is required")
		}

		var displayName string
		if err := s.db.QueryRow(
			`SELECT display_name
			 FROM local_profiles
			 WHERE local_profile_id = ?`,
			localProfileID,
		).Scan(&displayName); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return playerdata.LocalProfileDefault{}, playerdata.ErrLocalProfileNotFound
			}
			return playerdata.LocalProfileDefault{}, err
		}

		tx, err := s.db.Begin()
		if err != nil {
			return playerdata.LocalProfileDefault{}, err
		}
		defer func() {
			_ = tx.Rollback()
		}()

		if _, err := tx.Exec(
			`INSERT INTO local_profile_default (id, identity_kind, local_profile_id, updated_at)
			 VALUES (1, ?, ?, ?)
			 ON CONFLICT(id) DO UPDATE SET
				identity_kind = excluded.identity_kind,
				local_profile_id = excluded.local_profile_id,
				updated_at = excluded.updated_at`,
			playerdata.IdentityKindLocalProfile, localProfileID, now,
		); err != nil {
			return playerdata.LocalProfileDefault{}, err
		}

		if err := tx.Commit(); err != nil {
			return playerdata.LocalProfileDefault{}, err
		}

		return playerdata.LocalProfileDefault{
			IdentityKind:   playerdata.IdentityKindLocalProfile,
			LocalProfileID: localProfileID,
			DisplayName:    displayName,
		}, nil
	default:
		return playerdata.LocalProfileDefault{}, errors.New("identity_kind must be guest or local_profile")
	}
}

func (s *Store) CreateLocalProfile(localProfileID string, displayName string, stats protocol.PlayerDataStats) (playerdata.LocalProfileSummary, error) {
	if s == nil || s.db == nil {
		return playerdata.LocalProfileSummary{}, errors.New("sqlite store is not open")
	}
	if localProfileID == "" {
		return playerdata.LocalProfileSummary{}, errors.New("local_profile_id is required")
	}
	if displayName == "" {
		return playerdata.LocalProfileSummary{}, errors.New("display_name is required")
	}

	now := time.Now().UTC().Format(time.RFC3339)

	tx, err := s.db.Begin()
	if err != nil {
		return playerdata.LocalProfileSummary{}, err
	}
	defer func() {
		_ = tx.Rollback()
	}()

	if _, err := tx.Exec(
		`INSERT INTO local_profiles (local_profile_id, display_name, created_at, updated_at)
		 VALUES (?, ?, ?, ?)`,
		localProfileID, displayName, now, now,
	); err != nil {
		return playerdata.LocalProfileSummary{}, err
	}
	if _, err := tx.Exec(
		`INSERT INTO local_player_stats (local_profile_id, total_score, high_score, ship_deaths, games_played, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		localProfileID,
		stats.TotalScore,
		stats.HighScore,
		stats.ShipDeaths,
		stats.GamesPlayed,
		now,
		now,
	); err != nil {
		return playerdata.LocalProfileSummary{}, err
	}

	if err := tx.Commit(); err != nil {
		return playerdata.LocalProfileSummary{}, err
	}

	return playerdata.LocalProfileSummary{
		LocalProfileID: localProfileID,
		DisplayName:    displayName,
	}, nil
}

func (s *Store) UpdateLocalProfileDisplayName(localProfileID string, displayName string) (playerdata.LocalProfileSummary, error) {
	if s == nil || s.db == nil {
		return playerdata.LocalProfileSummary{}, errors.New("sqlite store is not open")
	}
	if localProfileID == "" {
		return playerdata.LocalProfileSummary{}, errors.New("local_profile_id is required")
	}
	if displayName == "" {
		return playerdata.LocalProfileSummary{}, errors.New("display_name is required")
	}

	tx, err := s.db.Begin()
	if err != nil {
		return playerdata.LocalProfileSummary{}, err
	}
	defer func() {
		_ = tx.Rollback()
	}()

	var found string
	err = tx.QueryRow(
		`SELECT local_profile_id
		 FROM local_profiles
		 WHERE local_profile_id = ?`,
		localProfileID,
	).Scan(&found)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return playerdata.LocalProfileSummary{}, playerdata.ErrLocalProfileNotFound
		}
		return playerdata.LocalProfileSummary{}, err
	}

	now := time.Now().UTC().Format(time.RFC3339)
	if _, err := tx.Exec(
		`UPDATE local_profiles
		 SET display_name = ?, updated_at = ?
		 WHERE local_profile_id = ?`,
		displayName, now, localProfileID,
	); err != nil {
		return playerdata.LocalProfileSummary{}, err
	}

	if err := tx.Commit(); err != nil {
		return playerdata.LocalProfileSummary{}, err
	}

	return playerdata.LocalProfileSummary{
		LocalProfileID: localProfileID,
		DisplayName:    displayName,
	}, nil
}

func (s *Store) DeleteLocalProfile(localProfileID string) error {
	if s == nil || s.db == nil {
		return errors.New("sqlite store is not open")
	}
	if localProfileID == "" {
		return errors.New("local_profile_id is required")
	}

	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer func() {
		_ = tx.Rollback()
	}()

	var found string
	err = tx.QueryRow(
		`SELECT local_profile_id
		 FROM local_profiles
		 WHERE local_profile_id = ?`,
		localProfileID,
	).Scan(&found)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return playerdata.ErrLocalProfileNotFound
		}
		return err
	}

	if _, err := tx.Exec(
		`DELETE FROM local_player_match_results
		 WHERE local_profile_id = ?`,
		localProfileID,
	); err != nil {
		return err
	}
	if _, err := tx.Exec(
		`DELETE FROM local_player_stats
		 WHERE local_profile_id = ?`,
		localProfileID,
	); err != nil {
		return err
	}
	if _, err := tx.Exec(
		`DELETE FROM local_profiles
		 WHERE local_profile_id = ?`,
		localProfileID,
	); err != nil {
		return err
	}

	now := time.Now().UTC().Format(time.RFC3339)
	if _, err := tx.Exec(
		`INSERT INTO local_profile_default (id, identity_kind, local_profile_id, updated_at)
		 VALUES (1, ?, ?, ?)
		 ON CONFLICT(id) DO UPDATE SET
			identity_kind = excluded.identity_kind,
			local_profile_id = excluded.local_profile_id,
			updated_at = excluded.updated_at
		 WHERE local_profile_default.identity_kind = ? AND local_profile_default.local_profile_id = ?`,
		playerdata.IdentityKindGuest, "", now, playerdata.IdentityKindLocalProfile, localProfileID,
	); err != nil {
		return err
	}

	return tx.Commit()
}

func (s *Store) LoadStats(identity protocol.PlayerDataIdentity) (protocol.PlayerDataStats, bool, error) {
	if s == nil || s.db == nil {
		return protocol.PlayerDataStats{}, false, errors.New("sqlite store is not open")
	}
	if identity.IdentityKind != playerdata.IdentityKindLocalProfile {
		return protocol.PlayerDataStats{}, false, errors.New("identity_kind must be local_profile")
	}
	if identity.LocalProfileID == "" {
		return protocol.PlayerDataStats{}, false, errors.New("local_profile_id is required")
	}

	if err := s.ensureLocalProfile(identity.LocalProfileID); err != nil {
		return protocol.PlayerDataStats{}, false, err
	}

	var stats protocol.PlayerDataStats
	err := s.db.QueryRow(
		`SELECT total_score, high_score, ship_deaths, games_played
		 FROM local_player_stats
		 WHERE local_profile_id = ?`,
		identity.LocalProfileID,
	).Scan(&stats.TotalScore, &stats.HighScore, &stats.ShipDeaths, &stats.GamesPlayed)
	if err != nil {
		return protocol.PlayerDataStats{}, false, err
	}

	stats.Wins = 0
	return stats, true, nil
}

func (s *Store) RecordMatchResult(command protocol.PlayerDataRecordMatchResult) (protocol.PlayerDataStats, bool, error) {
	if s == nil || s.db == nil {
		return protocol.PlayerDataStats{}, false, errors.New("sqlite store is not open")
	}
	if command.Identity.IdentityKind != playerdata.IdentityKindLocalProfile {
		return protocol.PlayerDataStats{}, false, errors.New("identity_kind must be local_profile")
	}
	if command.Identity.LocalProfileID == "" {
		return protocol.PlayerDataStats{}, false, errors.New("local_profile_id is required")
	}
	if command.ResultID == "" {
		return protocol.PlayerDataStats{}, false, errors.New("result_id is required")
	}
	if command.MatchID == "" {
		return protocol.PlayerDataStats{}, false, errors.New("match_id is required")
	}

	tx, err := s.db.Begin()
	if err != nil {
		return protocol.PlayerDataStats{}, false, err
	}
	defer func() {
		_ = tx.Rollback()
	}()

	now := time.Now().UTC().Format(time.RFC3339)
	if _, err := tx.Exec(
		`INSERT INTO local_profiles (local_profile_id, created_at, updated_at)
		 VALUES (?, ?, ?)
		 ON CONFLICT(local_profile_id) DO NOTHING`,
		command.Identity.LocalProfileID, now, now,
	); err != nil {
		return protocol.PlayerDataStats{}, false, err
	}
	if _, err := tx.Exec(
		`INSERT INTO local_player_stats (local_profile_id, total_score, high_score, ship_deaths, games_played, created_at, updated_at)
		 VALUES (?, 0, 0, 0, 0, ?, ?)
		 ON CONFLICT(local_profile_id) DO NOTHING`,
		command.Identity.LocalProfileID, now, now,
	); err != nil {
		return protocol.PlayerDataStats{}, false, err
	}

	err = tx.QueryRow(
		`SELECT 1
		 FROM local_player_match_results
		 WHERE result_id = ?`,
		command.ResultID,
	).Scan(new(int))
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return protocol.PlayerDataStats{}, false, err
	}
	if err == nil {
		stats, statsErr := s.loadLocalStatsTx(tx, command.Identity.LocalProfileID)
		if statsErr != nil {
			return protocol.PlayerDataStats{}, false, statsErr
		}
		if err := tx.Commit(); err != nil {
			return protocol.PlayerDataStats{}, false, err
		}
		return stats, true, nil
	}

	if _, err := tx.Exec(
		`INSERT INTO local_player_match_results (
			result_id, match_id, local_profile_id, score, ship_deaths, created_at
		) VALUES (?, ?, ?, ?, ?, ?)`,
		command.ResultID,
		command.MatchID,
		command.Identity.LocalProfileID,
		command.Score,
		command.ShipDeaths,
		now,
	); err != nil {
		return protocol.PlayerDataStats{}, false, err
	}
	if _, err := tx.Exec(
		`UPDATE local_player_stats
		 SET total_score = total_score + ?,
		     high_score = CASE WHEN high_score < ? THEN ? ELSE high_score END,
		     ship_deaths = ship_deaths + ?,
		     games_played = games_played + 1,
		     updated_at = ?
		 WHERE local_profile_id = ?`,
		command.Score,
		command.Score,
		command.Score,
		command.ShipDeaths,
		now,
		command.Identity.LocalProfileID,
	); err != nil {
		return protocol.PlayerDataStats{}, false, err
	}

	stats, err := s.loadLocalStatsTx(tx, command.Identity.LocalProfileID)
	if err != nil {
		return protocol.PlayerDataStats{}, false, err
	}
	if err := tx.Commit(); err != nil {
		return protocol.PlayerDataStats{}, false, err
	}

	return stats, false, nil
}

func (s *Store) loadLocalStatsTx(tx *sql.Tx, localProfileID string) (protocol.PlayerDataStats, error) {
	var stats protocol.PlayerDataStats
	err := tx.QueryRow(
		`SELECT total_score, high_score, ship_deaths, games_played
		 FROM local_player_stats
		 WHERE local_profile_id = ?`,
		localProfileID,
	).Scan(&stats.TotalScore, &stats.HighScore, &stats.ShipDeaths, &stats.GamesPlayed)
	if err != nil {
		return protocol.PlayerDataStats{}, err
	}
	stats.Wins = 0
	return stats, nil
}

func guestLocalProfileDefault() playerdata.LocalProfileDefault {
	return playerdata.LocalProfileDefault{
		IdentityKind:   playerdata.IdentityKindGuest,
		LocalProfileID: "",
		DisplayName:    "Guest",
	}
}
