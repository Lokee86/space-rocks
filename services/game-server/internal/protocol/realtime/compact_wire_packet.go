package realtime

var compactWireKeyMap = map[string]string{
	"type":                      "t",
	"lane":                      "l",
	"sequence":                  "q",
	"baseline_id":               "b",
	"snapshot_id":               "sid",
	"server_sent_msec":          "ms",
	"snapshot_kind":             "k",
	"chunk_index":               "ci",
	"chunk_count":               "cc",
	"is_final_chunk":            "fc",
	"ship_creates":              "sc",
	"ship_updates":              "su",
	"ship_deletes":              "sx",
	"bullet_creates":            "bc",
	"bullet_updates":            "bu",
	"bullet_deletes":            "bx",
	"asteroid_creates":          "ac",
	"asteroid_updates":          "au",
	"asteroid_deletes":          "ax",
	"pickup_creates":            "pc",
	"pickup_updates":            "pu",
	"pickup_deletes":            "px",
	"receiver_creates":          "rc",
	"receiver_updates":          "ru",
	"receiver_deletes":          "rx",
	"players":                   "pl",
	"player_session_updates":    "psu",
	"player_session_deletes":    "psx",
	"player_lifecycle":          "plc",
	"player_lifecycle_updates":   "plu",
	"player_lifecycle_deletes":   "plx",
	"total_asteroids":           "ta",
	"id":                        "i",
	"player_id":                 "pid",
	"self_id":                   "self",
	"status":                    "stat",
	"rotation":                  "r",
	"health":                    "h",
	"score":                     "sco",
	"lives":                     "lv",
	"respawn_cooldown":          "rcd",
	"ship_type":                 "st",
	"shields":                   "sh",
	"thrusting":                 "th",
	"target_kind":               "tk",
	"target_id":                 "tid",
	"owner_id":                  "oi",
	"weapon_id":                 "wid",
	"projectile_type":           "pt",
	"size":                      "sz",
	"scale":                     "sl",
	"variant":                   "v",
	"pickup_class":              "pcl",
	"age_seconds":               "age",
	"lifespan_seconds":          "life",
	"primary_weapon_id":         "pwid",
	"primary_ammo_policy":       "pap",
	"primary_cooldown_remaining": "pcr",
	"primary_ammo_remaining":    "par",
	"secondary_weapon_id":       "swid",
	"secondary_ammo_policy":     "sap",
	"secondary_cooldown_remaining": "scr",
	"secondary_ammo_remaining":   "sar",
	"spawn_x":                   "spx",
	"spawn_y":                   "spy",
}

var compactWireValueMaps = map[string]map[string]string{
	"type": {
		"world_full":   "wf",
		"world_delta":  "wd",
		"overlay_full": "of",
		"overlay_delta":"od",
		"session_full": "sf",
		"session_delta":"sd",
	},
	"lane": {
		"world":   "w",
		"overlay": "o",
		"session": "s",
	},
	"snapshot_kind": {
		"full":  "f",
		"delta": "d",
	},
}

func CompactWirePacket(packet map[string]any) map[string]any {
	expanded := compactWireValue(packet, "")
	if compacted, ok := expanded.(map[string]any); ok {
		return compacted
	}
	return map[string]any{}
}

func compactWireValue(value any, parentKey string) any {
	switch typed := value.(type) {
	case map[string]any:
		expanded := make(map[string]any, len(typed))
		for key, child := range typed {
			compactKey := key
			if mapped, ok := compactWireKeyMap[key]; ok {
				compactKey = mapped
			}
			expanded[compactKey] = compactWireValue(child, key)
		}
		return expanded
	case []any:
		expanded := make([]any, len(typed))
		for i := range typed {
			expanded[i] = compactWireValue(typed[i], parentKey)
		}
		return expanded
	default:
		if mapping, ok := compactWireValueMaps[parentKey]; ok {
			if compacted, ok := mapping[asString(typed)]; ok {
				return compacted
			}
		}
		return value
	}
}

func asString(value any) string {
	switch typed := value.(type) {
	case string:
		return typed
	default:
		return ""
	}
}
