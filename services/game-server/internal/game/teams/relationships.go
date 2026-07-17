package teams

import "fmt"

func RelationshipBetween(structure Structure, leftParticipantID string, leftTeamID ID, rightParticipantID string, rightTeamID ID) (Relationship, error) {
	if err := validateRelationshipInputs(structure, leftTeamID, rightTeamID); err != nil {
		return "", err
	}
	if leftParticipantID == rightParticipantID {
		return RelationshipSelf, nil
	}

	switch structure {
	case StructureFFA:
		return RelationshipOpposing, nil
	case StructureCoOp:
		return RelationshipSameTeam, nil
	case StructureCustom, StructureAutoBalanced:
		if leftTeamID == NoTeam || rightTeamID == NoTeam {
			return RelationshipUnaffiliated, nil
		}
		if leftTeamID == rightTeamID {
			return RelationshipSameTeam, nil
		}
		return RelationshipOpposing, nil
	default:
		return "", fmt.Errorf("unknown team structure %q", structure)
	}
}

func validateRelationshipInputs(structure Structure, leftTeamID ID, rightTeamID ID) error {
	switch structure {
	case StructureFFA, StructureCoOp, StructureCustom, StructureAutoBalanced:
	default:
		return fmt.Errorf("unknown team structure %q", structure)
	}
	if err := ValidateTeamID(leftTeamID); err != nil {
		return fmt.Errorf("left participant: %w", err)
	}
	if err := ValidateTeamID(rightTeamID); err != nil {
		return fmt.Errorf("right participant: %w", err)
	}
	return nil
}
