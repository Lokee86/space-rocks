package game

import (
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/damage"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/teams"
)

func (game *Game) SetDamagePolicy(policy damage.Policy) {
	game.mu.Lock()
	defer game.mu.Unlock()
	if policy.ID != damage.PolicyStandardV1 {
		policy = damage.NewStandardPolicy()
	}
	game.damagePolicy = policy
}

func (game *Game) DamagePolicy() damage.Policy {
	game.mu.Lock()
	defer game.mu.Unlock()
	return game.damagePolicy
}

func (game *Game) resolveDamageRequest(request damage.DamageResolutionRequest) damage.DamageResult {
	if request.Source.ResponsiblePlayerID != "" && request.Source.ResponsibleTeamID == "" {
		if teamID, ok := game.playerTeamLocked(request.Source.ResponsiblePlayerID); ok {
			request.Source.ResponsibleTeamID = string(teamID)
		}
	}
	eligibility := damage.EvaluateEligibility(damage.EligibilityRequest{
		Policy:                   game.damagePolicy,
		Source:                   request.Source,
		Target:                   request.Target,
		Relationship:             game.damageRelationship(request.Source, request.Target),
		Permissions:              request.Source.Permissions,
		UseExplicitPermissions:   request.Source.UseExplicitPermissions,
		SourceIsPlayerControlled: request.Source.ResponsiblePlayerID != "",
		TargetInvulnerable:       request.Target.Invulnerable,
		BypassInvulnerability:    request.Source.BypassInvulnerability,
		AuthorizedDevAdminSource: request.Source.AuthorizedDevAdminSource,
	})
	if !eligibility.Eligible {
		return damage.BlockedResult(request, eligibility.Reason)
	}
	return damage.ResolveSingle(request)
}

func (game *Game) damageRelationship(source damage.DamageSource, target damage.DamageTarget) damage.Relationship {
	switch target.EntityType {
	case damage.EntityTypeAsteroid, damage.EntityTypeEnvironment:
		return damage.RelationshipDestructible
	case damage.EntityTypeEnemy:
		if source.ResponsiblePlayerID != "" {
			return damage.RelationshipEnemy
		}
		return damage.RelationshipNeutral
	case damage.EntityTypePlayer:
		if source.ResponsiblePlayerID == "" {
			return damage.RelationshipNeutral
		}
		leftTeam, leftOK := game.playerTeamLocked(source.ResponsiblePlayerID)
		rightTeam, rightOK := game.playerTeamLocked(target.EntityID)
		if !leftOK || !rightOK {
			if source.ResponsiblePlayerID == target.EntityID {
				return damage.RelationshipSelf
			}
			return damage.RelationshipNeutral
		}
		relationship, err := teams.RelationshipBetween(game.teamStructure, source.ResponsiblePlayerID, leftTeam, target.EntityID, rightTeam)
		if err != nil {
			return damage.RelationshipNeutral
		}
		switch relationship {
		case teams.RelationshipSelf:
			return damage.RelationshipSelf
		case teams.RelationshipSameTeam:
			return damage.RelationshipAlly
		case teams.RelationshipOpposing:
			return damage.RelationshipEnemy
		default:
			return damage.RelationshipNeutral
		}
	default:
		return damage.RelationshipNeutral
	}
}
