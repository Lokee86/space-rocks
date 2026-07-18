package matchresults

func cloneParticipantFacts(source []ParticipantFact) []ParticipantFact {
	return append([]ParticipantFact(nil), source...)
}

func cloneObjectives(source []ObjectiveResolution) []ObjectiveResolution {
	return append([]ObjectiveResolution(nil), source...)
}

func cloneMissions(source []MissionResolution) []MissionResolution {
	return append([]MissionResolution(nil), source...)
}

func cloneChallenges(source []ChallengeResolutionAggregate) []ChallengeResolutionAggregate {
	clone := make([]ChallengeResolutionAggregate, len(source))
	for index, challenge := range source {
		clone[index] = challenge
		if challenge.Values != nil {
			clone[index].Values = make(map[string]float64, len(challenge.Values))
			for key, value := range challenge.Values {
				clone[index].Values[key] = value
			}
		}
	}
	return clone
}

func cloneSummary(source MatchSummary) MatchSummary {
	clone := source
	clone.Participants = append([]ParticipantResult(nil), source.Participants...)
	clone.Teams = append([]TeamResult(nil), source.Teams...)
	clone.WinningPlayerRefs = append([]PlayerRef(nil), source.WinningPlayerRefs...)
	clone.WinningTeamRefs = append(clone.WinningTeamRefs[:0:0], source.WinningTeamRefs...)
	clone.Objectives = cloneObjectives(source.Objectives)
	clone.Missions = cloneMissions(source.Missions)
	clone.Challenges = cloneChallenges(source.Challenges)
	return clone
}
