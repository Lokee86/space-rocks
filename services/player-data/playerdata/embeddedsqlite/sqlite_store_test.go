//go:build !noembeddedsqlite

package embeddedsqlite

import (
	"database/sql"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Lokee86/space-rocks/player-data/playerdata"
	"github.com/Lokee86/space-rocks/player-data/protocol"
)

func TestNew(t *testing.T) {
	t.Run("empty path", func(t *testing.T) {
		store, err := New(Config{})
		if err == nil {
			t.Fatal("NewStore returned nil error for empty path")
		}
		if store != nil {
			t.Fatalf("NewStore returned store %+v for empty path", store)
		}
	})

	t.Run("memory path and close", func(t *testing.T) {
		store, err := New(Config{Path: ":memory:"})
		if err != nil {
			t.Fatalf("NewStore returned error: %v", err)
		}
		if store == nil {
			t.Fatal("NewStore returned nil store for :memory:")
		}
		if err := store.Close(); err != nil {
			t.Fatalf("Close returned error: %v", err)
		}
	})

	t.Run("creates parent directory for file path", func(t *testing.T) {
		tempDir := t.TempDir()
		dbPath := filepath.Join(tempDir, "nested", "player-data.sqlite3")

		store, err := New(Config{Path: dbPath})
		if err != nil {
			t.Fatalf("NewStore returned error: %v", err)
		}
		t.Cleanup(func() {
			_ = store.Close()
		})

		parentDir := filepath.Dir(dbPath)
		if info, err := os.Stat(parentDir); err != nil {
			t.Fatalf("parent directory stat failed: %v", err)
		} else if !info.IsDir() {
			t.Fatalf("parent path %q is not a directory", parentDir)
		}

		if err := store.InitSchema(); err != nil {
			t.Fatalf("InitSchema returned error: %v", err)
		}

		if _, err := os.Stat(dbPath); err != nil {
			t.Fatalf("database file stat failed: %v", err)
		}
	})
}

func TestStoreInitSchema(t *testing.T) {
	store, err := New(Config{Path: ":memory:"})
	if err != nil {
		t.Fatalf("NewStore returned error: %v", err)
	}
	t.Cleanup(func() {
		_ = store.Close()
	})

	if err := store.InitSchema(); err != nil {
		t.Fatalf("InitSchema returned error: %v", err)
	}
	if err := store.InitSchema(); err != nil {
		t.Fatalf("InitSchema returned error on second call: %v", err)
	}

	rows, err := store.db.Query(`PRAGMA table_info(local_player_stats)`)
	if err != nil {
		t.Fatalf("PRAGMA table_info query failed: %v", err)
	}
	defer rows.Close()

	var columns []string
	for rows.Next() {
		var (
			cid        int
			name       string
			columnType string
			notnull    int
			dfltValue  any
			pk         int
		)
		if err := rows.Scan(&cid, &name, &columnType, &notnull, &dfltValue, &pk); err != nil {
			t.Fatalf("Scan failed: %v", err)
		}
		columns = append(columns, name)
		if name == "wins" {
			t.Fatal("local_player_stats unexpectedly includes wins")
		}
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("rows iteration failed: %v", err)
	}

	got := strings.Join(columns, ",")
	wantColumns := []string{
		"local_profile_id",
		"total_score",
		"high_score",
		"ship_deaths",
		"games_played",
		"created_at",
		"updated_at",
	}
	for _, want := range wantColumns {
		if !strings.Contains(got, want) {
			t.Fatalf("local_player_stats columns = %q, missing %q", got, want)
		}
	}
}

func TestStoreEnsureLocalProfile(t *testing.T) {
	store, err := New(Config{Path: ":memory:"})
	if err != nil {
		t.Fatalf("NewStore returned error: %v", err)
	}
	t.Cleanup(func() {
		_ = store.Close()
	})

	if err := store.InitSchema(); err != nil {
		t.Fatalf("InitSchema returned error: %v", err)
	}
	if err := store.ensureLocalProfile("profile-1"); err != nil {
		t.Fatalf("ensureLocalProfile returned error: %v", err)
	}
	if err := store.ensureLocalProfile("profile-1"); err != nil {
		t.Fatalf("ensureLocalProfile returned error on second call: %v", err)
	}

	var createdAt string
	if err := store.db.QueryRow(`SELECT created_at FROM local_profiles WHERE local_profile_id = ?`, "profile-1").Scan(&createdAt); err != nil {
		t.Fatalf("query local_profiles failed: %v", err)
	}
	if createdAt == "" {
		t.Fatal("local_profiles.created_at was empty")
	}

	var totalScore, highScore, shipDeaths, gamesPlayed int
	var statsCreatedAt string
	err = store.db.QueryRow(`
		SELECT total_score, high_score, ship_deaths, games_played, created_at
		FROM local_player_stats
		WHERE local_profile_id = ?`,
		"profile-1",
	).Scan(&totalScore, &highScore, &shipDeaths, &gamesPlayed, &statsCreatedAt)
	if err != nil {
		t.Fatalf("query local_player_stats failed: %v", err)
	}
	if totalScore != 0 || highScore != 0 || shipDeaths != 0 || gamesPlayed != 0 {
		t.Fatalf("stats row = (%d, %d, %d, %d), want all zeroes", totalScore, highScore, shipDeaths, gamesPlayed)
	}
	if statsCreatedAt == "" {
		t.Fatal("local_player_stats.created_at was empty")
	}
}

func TestStoreEnsureLocalProfileRejectsEmptyID(t *testing.T) {
	store, err := New(Config{Path: ":memory:"})
	if err != nil {
		t.Fatalf("NewStore returned error: %v", err)
	}
	t.Cleanup(func() {
		_ = store.Close()
	})

	if err := store.InitSchema(); err != nil {
		t.Fatalf("InitSchema returned error: %v", err)
	}
	if err := store.ensureLocalProfile(""); err == nil {
		t.Fatal("ensureLocalProfile returned nil error for empty localProfileID")
	}
}

func TestStoreLoadStats(t *testing.T) {
	t.Run("new profile returns zero stats", func(t *testing.T) {
		store, err := New(Config{Path: ":memory:"})
		if err != nil {
			t.Fatalf("NewStore returned error: %v", err)
		}
		t.Cleanup(func() {
			_ = store.Close()
		})

		if err := store.InitSchema(); err != nil {
			t.Fatalf("InitSchema returned error: %v", err)
		}

		stats, found, err := store.LoadStats(protocol.PlayerDataIdentity{
			IdentityKind:   playerdata.IdentityKindLocalProfile,
			LocalProfileID: "profile-1",
		})
		if err != nil {
			t.Fatalf("LoadStats returned error: %v", err)
		}
		if !found {
			t.Fatal("LoadStats returned found=false for new local profile")
		}
		if stats.TotalScore != 0 || stats.HighScore != 0 || stats.ShipDeaths != 0 || stats.GamesPlayed != 0 {
			t.Fatalf("LoadStats returned non-zero stats: %+v", stats)
		}
		if stats.Wins != 0 {
			t.Fatalf("LoadStats returned wins=%d, want 0", stats.Wins)
		}
	})

	t.Run("invalid identity", func(t *testing.T) {
		store, err := New(Config{Path: ":memory:"})
		if err != nil {
			t.Fatalf("NewStore returned error: %v", err)
		}
		t.Cleanup(func() {
			_ = store.Close()
		})

		if err := store.InitSchema(); err != nil {
			t.Fatalf("InitSchema returned error: %v", err)
		}

		if _, _, err := store.LoadStats(protocol.PlayerDataIdentity{IdentityKind: playerdata.IdentityKindGuest}); err == nil {
			t.Fatal("LoadStats returned nil error for invalid identity kind")
		}
		if _, _, err := store.LoadStats(protocol.PlayerDataIdentity{IdentityKind: playerdata.IdentityKindLocalProfile}); err == nil {
			t.Fatal("LoadStats returned nil error for missing local profile id")
		}
	})
}

func TestStoreRecordMatchResult(t *testing.T) {
	t.Run("first record", func(t *testing.T) {
		store, err := New(Config{Path: ":memory:"})
		if err != nil {
			t.Fatalf("NewStore returned error: %v", err)
		}
		t.Cleanup(func() {
			_ = store.Close()
		})

		if err := store.InitSchema(); err != nil {
			t.Fatalf("InitSchema returned error: %v", err)
		}

		stats, duplicate, err := store.RecordMatchResult(protocol.PlayerDataRecordMatchResult{
			ResultID: "result-1",
			MatchID:  "match-1",
			Identity: protocol.PlayerDataIdentity{
				IdentityKind:   playerdata.IdentityKindLocalProfile,
				LocalProfileID: "profile-1",
			},
			Score:      10,
			ShipDeaths: 2,
			Won:        true,
		})
		if err != nil {
			t.Fatalf("RecordMatchResult returned error: %v", err)
		}
		if duplicate {
			t.Fatal("RecordMatchResult returned duplicate=true for first record")
		}
		if stats.TotalScore != 10 || stats.HighScore != 10 || stats.ShipDeaths != 2 || stats.GamesPlayed != 1 {
			t.Fatalf("RecordMatchResult returned stats %+v, want first record stats", stats)
		}
		if stats.Wins != 0 {
			t.Fatalf("RecordMatchResult returned wins=%d, want 0", stats.Wins)
		}
	})

	t.Run("high score increasing and not decreasing", func(t *testing.T) {
		store, err := New(Config{Path: ":memory:"})
		if err != nil {
			t.Fatalf("NewStore returned error: %v", err)
		}
		t.Cleanup(func() {
			_ = store.Close()
		})

		if err := store.InitSchema(); err != nil {
			t.Fatalf("InitSchema returned error: %v", err)
		}

		first, duplicate, err := store.RecordMatchResult(protocol.PlayerDataRecordMatchResult{
			ResultID: "result-1",
			MatchID:  "match-1",
			Identity: protocol.PlayerDataIdentity{
				IdentityKind:   playerdata.IdentityKindLocalProfile,
				LocalProfileID: "profile-1",
			},
			Score: 10,
			Won:   false,
		})
		if err != nil {
			t.Fatalf("first RecordMatchResult returned error: %v", err)
		}
		if duplicate {
			t.Fatal("first RecordMatchResult returned duplicate=true")
		}
		second, duplicate, err := store.RecordMatchResult(protocol.PlayerDataRecordMatchResult{
			ResultID: "result-2",
			MatchID:  "match-2",
			Identity: protocol.PlayerDataIdentity{
				IdentityKind:   playerdata.IdentityKindLocalProfile,
				LocalProfileID: "profile-1",
			},
			Score: 25,
			Won:   false,
		})
		if err != nil {
			t.Fatalf("second RecordMatchResult returned error: %v", err)
		}
		if duplicate {
			t.Fatal("second RecordMatchResult returned duplicate=true")
		}
		if second.HighScore != 25 {
			t.Fatalf("HighScore = %d, want 25", second.HighScore)
		}

		third, duplicate, err := store.RecordMatchResult(protocol.PlayerDataRecordMatchResult{
			ResultID: "result-3",
			MatchID:  "match-3",
			Identity: protocol.PlayerDataIdentity{
				IdentityKind:   playerdata.IdentityKindLocalProfile,
				LocalProfileID: "profile-1",
			},
			Score: 7,
			Won:   false,
		})
		if err != nil {
			t.Fatalf("third RecordMatchResult returned error: %v", err)
		}
		if duplicate {
			t.Fatal("third RecordMatchResult returned duplicate=true")
		}
		if third.HighScore != 25 {
			t.Fatalf("HighScore = %d, want 25", third.HighScore)
		}
		if first.Wins != 0 || second.Wins != 0 || third.Wins != 0 {
			t.Fatalf("wins changed in local stats: first=%d second=%d third=%d", first.Wins, second.Wins, third.Wins)
		}
	})

	t.Run("duplicate idempotency", func(t *testing.T) {
		store, err := New(Config{Path: ":memory:"})
		if err != nil {
			t.Fatalf("NewStore returned error: %v", err)
		}
		t.Cleanup(func() {
			_ = store.Close()
		})

		if err := store.InitSchema(); err != nil {
			t.Fatalf("InitSchema returned error: %v", err)
		}

		first, duplicate, err := store.RecordMatchResult(protocol.PlayerDataRecordMatchResult{
			ResultID: "result-1",
			MatchID:  "match-1",
			Identity: protocol.PlayerDataIdentity{
				IdentityKind:   playerdata.IdentityKindLocalProfile,
				LocalProfileID: "profile-1",
			},
			Score: 10,
			Won:   true,
		})
		if err != nil {
			t.Fatalf("first RecordMatchResult returned error: %v", err)
		}
		if duplicate {
			t.Fatal("first RecordMatchResult returned duplicate=true")
		}

		second, duplicate, err := store.RecordMatchResult(protocol.PlayerDataRecordMatchResult{
			ResultID: "result-1",
			MatchID:  "match-1",
			Identity: protocol.PlayerDataIdentity{
				IdentityKind:   playerdata.IdentityKindLocalProfile,
				LocalProfileID: "profile-1",
			},
			Score: 99,
			Won:   true,
		})
		if err != nil {
			t.Fatalf("duplicate RecordMatchResult returned error: %v", err)
		}
		if !duplicate {
			t.Fatal("duplicate RecordMatchResult returned duplicate=false")
		}
		if second != first {
			t.Fatalf("duplicate stats = %+v, want %+v", second, first)
		}
	})

	t.Run("invalid identity", func(t *testing.T) {
		store, err := New(Config{Path: ":memory:"})
		if err != nil {
			t.Fatalf("NewStore returned error: %v", err)
		}
		t.Cleanup(func() {
			_ = store.Close()
		})

		if err := store.InitSchema(); err != nil {
			t.Fatalf("InitSchema returned error: %v", err)
		}

		if _, _, err := store.RecordMatchResult(protocol.PlayerDataRecordMatchResult{
			Identity: protocol.PlayerDataIdentity{IdentityKind: playerdata.IdentityKindGuest},
			ResultID: "result-1",
			MatchID:  "match-1",
		}); err == nil {
			t.Fatal("RecordMatchResult returned nil error for invalid identity kind")
		}
		if _, _, err := store.RecordMatchResult(protocol.PlayerDataRecordMatchResult{
			Identity: protocol.PlayerDataIdentity{IdentityKind: playerdata.IdentityKindLocalProfile},
			ResultID: "result-1",
			MatchID:  "match-1",
		}); err == nil {
			t.Fatal("RecordMatchResult returned nil error for missing local profile id")
		}
		if _, _, err := store.RecordMatchResult(protocol.PlayerDataRecordMatchResult{
			Identity: protocol.PlayerDataIdentity{
				IdentityKind:   playerdata.IdentityKindLocalProfile,
				LocalProfileID: "profile-1",
			},
			MatchID: "match-1",
		}); err == nil {
			t.Fatal("RecordMatchResult returned nil error for missing result id")
		}
		if _, _, err := store.RecordMatchResult(protocol.PlayerDataRecordMatchResult{
			Identity: protocol.PlayerDataIdentity{
				IdentityKind:   playerdata.IdentityKindLocalProfile,
				LocalProfileID: "profile-1",
			},
			ResultID: "result-1",
		}); err == nil {
			t.Fatal("RecordMatchResult returned nil error for missing match id")
		}
	})
}

func TestStoreDeleteLocalProfileDeletesRelatedRows(t *testing.T) {
	store, err := New(Config{Path: ":memory:"})
	if err != nil {
		t.Fatalf("NewStore returned error: %v", err)
	}
	t.Cleanup(func() {
		_ = store.Close()
	})

	if err := store.InitSchema(); err != nil {
		t.Fatalf("InitSchema returned error: %v", err)
	}

	_, err = store.CreateLocalProfile("profile-1", "Pilot One", protocol.PlayerDataStats{
		TotalScore:  7,
		HighScore:   7,
		ShipDeaths:  1,
		GamesPlayed: 1,
	})
	if err != nil {
		t.Fatalf("CreateLocalProfile returned error: %v", err)
	}

	if _, _, err := store.RecordMatchResult(protocol.PlayerDataRecordMatchResult{
		ResultID: "result-1",
		MatchID:  "match-1",
		Identity: protocol.PlayerDataIdentity{
			IdentityKind:   playerdata.IdentityKindLocalProfile,
			LocalProfileID: "profile-1",
		},
		Score:      12,
		ShipDeaths:  2,
		Won:        true,
	}); err != nil {
		t.Fatalf("RecordMatchResult returned error: %v", err)
	}

	if err := store.DeleteLocalProfile("profile-1"); err != nil {
		t.Fatalf("DeleteLocalProfile returned error: %v", err)
	}

	if err := store.db.QueryRow(
		`SELECT local_profile_id
		 FROM local_profiles
		 WHERE local_profile_id = ?`,
		"profile-1",
	).Scan(new(string)); err != sql.ErrNoRows {
		t.Fatalf("local_profiles row still present or unexpected error: %v", err)
	}
	if err := store.db.QueryRow(
		`SELECT local_profile_id
		 FROM local_player_stats
		 WHERE local_profile_id = ?`,
		"profile-1",
	).Scan(new(string)); err != sql.ErrNoRows {
		t.Fatalf("local_player_stats row still present or unexpected error: %v", err)
	}
	if err := store.db.QueryRow(
		`SELECT result_id
		 FROM local_player_match_results
		 WHERE local_profile_id = ?`,
		"profile-1",
	).Scan(new(string)); err != sql.ErrNoRows {
		t.Fatalf("local_player_match_results row still present or unexpected error: %v", err)
	}
}

func TestStoreDeleteLocalProfileResetsDefaultToGuest(t *testing.T) {
	store, err := New(Config{Path: ":memory:"})
	if err != nil {
		t.Fatalf("NewStore returned error: %v", err)
	}
	t.Cleanup(func() {
		_ = store.Close()
	})

	if err := store.InitSchema(); err != nil {
		t.Fatalf("InitSchema returned error: %v", err)
	}

	_, err = store.CreateLocalProfile("profile-1", "Pilot One", protocol.PlayerDataStats{})
	if err != nil {
		t.Fatalf("CreateLocalProfile returned error: %v", err)
	}
	if _, err := store.SetDefaultLocalProfile(playerdata.IdentityKindLocalProfile, "profile-1"); err != nil {
		t.Fatalf("SetDefaultLocalProfile returned error: %v", err)
	}

	if err := store.DeleteLocalProfile("profile-1"); err != nil {
		t.Fatalf("DeleteLocalProfile returned error: %v", err)
	}

	defaultProfile, err := store.GetDefaultLocalProfile()
	if err != nil {
		t.Fatalf("GetDefaultLocalProfile returned error: %v", err)
	}
	if defaultProfile.IdentityKind != playerdata.IdentityKindGuest {
		t.Fatalf("IdentityKind = %q, want %q", defaultProfile.IdentityKind, playerdata.IdentityKindGuest)
	}
	if defaultProfile.LocalProfileID != "" {
		t.Fatalf("LocalProfileID = %q, want empty", defaultProfile.LocalProfileID)
	}
	if defaultProfile.DisplayName != "Guest" {
		t.Fatalf("DisplayName = %q, want Guest", defaultProfile.DisplayName)
	}
}

func TestStoreUpdateLocalProfileDisplayName(t *testing.T) {
	store, err := New(Config{Path: ":memory:"})
	if err != nil {
		t.Fatalf("NewStore returned error: %v", err)
	}
	t.Cleanup(func() {
		_ = store.Close()
	})

	if err := store.InitSchema(); err != nil {
		t.Fatalf("InitSchema returned error: %v", err)
	}

	_, err = store.CreateLocalProfile("profile-1", "Pilot One", protocol.PlayerDataStats{
		TotalScore:  4,
		HighScore:   4,
		ShipDeaths:  1,
		GamesPlayed: 1,
	})
	if err != nil {
		t.Fatalf("CreateLocalProfile returned error: %v", err)
	}

	if _, _, err := store.RecordMatchResult(protocol.PlayerDataRecordMatchResult{
		ResultID: "result-1",
		MatchID:  "match-1",
		Identity: protocol.PlayerDataIdentity{
			IdentityKind:   playerdata.IdentityKindLocalProfile,
			LocalProfileID: "profile-1",
		},
		Score:      9,
		ShipDeaths: 2,
		Won:        false,
	}); err != nil {
		t.Fatalf("RecordMatchResult returned error: %v", err)
	}

	profile, err := store.UpdateLocalProfileDisplayName("profile-1", "Ace Pilot")
	if err != nil {
		t.Fatalf("UpdateLocalProfileDisplayName returned error: %v", err)
	}
	if profile.LocalProfileID != "profile-1" {
		t.Fatalf("LocalProfileID = %q, want profile-1", profile.LocalProfileID)
	}
	if profile.DisplayName != "Ace Pilot" {
		t.Fatalf("DisplayName = %q, want Ace Pilot", profile.DisplayName)
	}

	profiles, err := store.ListLocalProfiles()
	if err != nil {
		t.Fatalf("ListLocalProfiles returned error: %v", err)
	}
	if len(profiles) != 1 {
		t.Fatalf("ListLocalProfiles length = %d, want 1", len(profiles))
	}
	if profiles[0].LocalProfileID != "profile-1" {
		t.Fatalf("ListLocalProfiles[0].LocalProfileID = %q, want profile-1", profiles[0].LocalProfileID)
	}
	if profiles[0].DisplayName != "Ace Pilot" {
		t.Fatalf("ListLocalProfiles[0].DisplayName = %q, want Ace Pilot", profiles[0].DisplayName)
	}

	stats, found, err := store.LoadStats(protocol.PlayerDataIdentity{
		IdentityKind:   playerdata.IdentityKindLocalProfile,
		LocalProfileID: "profile-1",
	})
	if err != nil {
		t.Fatalf("LoadStats returned error: %v", err)
	}
	if !found {
		t.Fatal("LoadStats returned found=false")
	}
	if stats.TotalScore != 13 {
		t.Fatalf("TotalScore = %d, want 13", stats.TotalScore)
	}
	if stats.HighScore != 9 {
		t.Fatalf("HighScore = %d, want 9", stats.HighScore)
	}
	if stats.ShipDeaths != 3 {
		t.Fatalf("ShipDeaths = %d, want 3", stats.ShipDeaths)
	}
	if stats.GamesPlayed != 2 {
		t.Fatalf("GamesPlayed = %d, want 2", stats.GamesPlayed)
	}
}

func TestStoreUpdateLocalProfileDisplayNameMissingID(t *testing.T) {
	store, err := New(Config{Path: ":memory:"})
	if err != nil {
		t.Fatalf("NewStore returned error: %v", err)
	}
	t.Cleanup(func() {
		_ = store.Close()
	})

	if err := store.InitSchema(); err != nil {
		t.Fatalf("InitSchema returned error: %v", err)
	}

	_, err = store.UpdateLocalProfileDisplayName("missing-profile", "Ace Pilot")
	if err == nil {
		t.Fatal("UpdateLocalProfileDisplayName returned nil error for missing local profile")
	}
	if !errors.Is(err, playerdata.ErrLocalProfileNotFound) {
		t.Fatalf("UpdateLocalProfileDisplayName error = %v, want ErrLocalProfileNotFound", err)
	}
}

func TestStoreUpdateLocalProfileDisplayNameRejectsEmptyDisplayName(t *testing.T) {
	store, err := New(Config{Path: ":memory:"})
	if err != nil {
		t.Fatalf("NewStore returned error: %v", err)
	}
	t.Cleanup(func() {
		_ = store.Close()
	})

	if err := store.InitSchema(); err != nil {
		t.Fatalf("InitSchema returned error: %v", err)
	}

	_, err = store.CreateLocalProfile("profile-1", "Pilot One", protocol.PlayerDataStats{})
	if err != nil {
		t.Fatalf("CreateLocalProfile returned error: %v", err)
	}

	_, err = store.UpdateLocalProfileDisplayName("profile-1", "")
	if err == nil {
		t.Fatal("UpdateLocalProfileDisplayName returned nil error for empty display name")
	}
	if !strings.Contains(err.Error(), "display_name is required") {
		t.Fatalf("UpdateLocalProfileDisplayName error = %v, want it to contain %q", err, "display_name is required")
	}
}

func TestStoreDeleteLocalProfileMissingID(t *testing.T) {
	store, err := New(Config{Path: ":memory:"})
	if err != nil {
		t.Fatalf("NewStore returned error: %v", err)
	}
	t.Cleanup(func() {
		_ = store.Close()
	})

	if err := store.InitSchema(); err != nil {
		t.Fatalf("InitSchema returned error: %v", err)
	}

	err = store.DeleteLocalProfile("missing-profile")
	if err == nil {
		t.Fatal("DeleteLocalProfile returned nil error for missing local profile")
	}
	if !errors.Is(err, playerdata.ErrLocalProfileNotFound) {
		t.Fatalf("DeleteLocalProfile error = %v, want ErrLocalProfileNotFound", err)
	}
}

func TestStorePersistenceAcrossReopen(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "player-data.sqlite")

	store, err := New(Config{Path: dbPath})
	if err != nil {
		t.Fatalf("NewStore returned error: %v", err)
	}

	if err := store.InitSchema(); err != nil {
		t.Fatalf("InitSchema returned error: %v", err)
	}

	_, duplicate, err := store.RecordMatchResult(protocol.PlayerDataRecordMatchResult{
		ResultID: "result-1",
		MatchID:  "match-1",
		Identity: protocol.PlayerDataIdentity{
			IdentityKind:   playerdata.IdentityKindLocalProfile,
			LocalProfileID: "profile-1",
		},
		Score:      14,
		ShipDeaths: 3,
		Won:        true,
	})
	if err != nil {
		t.Fatalf("RecordMatchResult returned error: %v", err)
	}
	if duplicate {
		t.Fatal("RecordMatchResult returned duplicate=true for first record")
	}
	if err := store.Close(); err != nil {
		t.Fatalf("Close returned error: %v", err)
	}

	reopened, err := New(Config{Path: dbPath})
	if err != nil {
		t.Fatalf("NewStore reopened returned error: %v", err)
	}
	t.Cleanup(func() {
		_ = reopened.Close()
	})

	if err := reopened.InitSchema(); err != nil {
		t.Fatalf("reopened InitSchema returned error: %v", err)
	}

	stats, found, err := reopened.LoadStats(protocol.PlayerDataIdentity{
		IdentityKind:   playerdata.IdentityKindLocalProfile,
		LocalProfileID: "profile-1",
	})
	if err != nil {
		t.Fatalf("LoadStats returned error: %v", err)
	}
	if !found {
		t.Fatal("LoadStats returned found=false after reopen")
	}
	if stats.TotalScore != 14 {
		t.Fatalf("TotalScore = %d, want 14", stats.TotalScore)
	}
	if stats.HighScore != 14 {
		t.Fatalf("HighScore = %d, want 14", stats.HighScore)
	}
	if stats.ShipDeaths != 3 {
		t.Fatalf("ShipDeaths = %d, want 3", stats.ShipDeaths)
	}
	if stats.GamesPlayed != 1 {
		t.Fatalf("GamesPlayed = %d, want 1", stats.GamesPlayed)
	}
	if stats.Wins != 0 {
		t.Fatalf("Wins = %d, want 0", stats.Wins)
	}
}
