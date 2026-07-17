package teams

import "fmt"

func ResolveAssignments(config Config, participantIDs []string, requested Assignments) (Assignments, error) {
	if err := ValidateConfig(config); err != nil {
		return nil, err
	}
	if len(requested) != 0 && config.Structure != StructureCustom {
		return nil, fmt.Errorf("requested assignments are only valid for custom teams")
	}

	switch config.Structure {
	case StructureFFA:
		return ResolveFFA(participantIDs)
	case StructureCoOp:
		return ResolveCoop(participantIDs)
	case StructureCustom:
		return ResolveCustom(participantIDs, requested)
	case StructureAutoBalanced:
		return ResolveAutoBalanced(participantIDs, config.AutoTeamCount)
	default:
		return nil, fmt.Errorf("unknown team structure %q", config.Structure)
	}
}
