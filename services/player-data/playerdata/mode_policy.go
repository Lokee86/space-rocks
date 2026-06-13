package playerdata

import (
	"fmt"

	"github.com/Lokee86/space-rocks/player-data/protocol"
)

func ValidateModeIdentity(playMode string, identity protocol.PlayerDataIdentity) error {
	switch playMode {
	case PlayModeSinglePlayer:
		switch identity.IdentityKind {
		case IdentityKindGuest, IdentityKindLocalProfile:
			return nil
		case IdentityKindAuthenticatedAccount:
			return fmt.Errorf("identity_kind %q is not allowed for play_mode %q", identity.IdentityKind, playMode)
		case "":
			return fmt.Errorf("identity_kind is required for play_mode %q", playMode)
		default:
			return fmt.Errorf("unknown identity_kind %q", identity.IdentityKind)
		}
	case PlayModeMultiplayer:
		switch identity.IdentityKind {
		case IdentityKindAuthenticatedAccount:
			return nil
		case IdentityKindGuest, IdentityKindLocalProfile:
			return fmt.Errorf("identity_kind %q is not allowed for play_mode %q", identity.IdentityKind, playMode)
		case "":
			return fmt.Errorf("identity_kind is required for play_mode %q", playMode)
		default:
			return fmt.Errorf("unknown identity_kind %q", identity.IdentityKind)
		}
	case PlayModeMultiplayerSimulation:
		switch identity.IdentityKind {
		case IdentityKindAuthenticatedAccount:
			return nil
		case IdentityKindGuest, IdentityKindLocalProfile:
			return fmt.Errorf("identity_kind %q is not allowed for play_mode %q", identity.IdentityKind, playMode)
		case "":
			return fmt.Errorf("identity_kind is required for play_mode %q", playMode)
		default:
			return fmt.Errorf("unknown identity_kind %q", identity.IdentityKind)
		}
	case "":
		return fmt.Errorf("play_mode is required")
	default:
		return fmt.Errorf("unknown play_mode %q", playMode)
	}
}
