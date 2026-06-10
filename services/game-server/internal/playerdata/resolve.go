package playerdata

// ResolveWinners returns a new slice with winner flags resolved for the match mode.
func ResolveWinners(mode MatchMode, players []PlayerMatchSummary) []PlayerMatchSummary {
	if len(players) == 0 {
		return []PlayerMatchSummary{}
	}

	result := make([]PlayerMatchSummary, len(players))
	copy(result, players)

	if mode == MatchModeSinglePlayer {
		for i := range result {
			result[i].Won = false
		}
		return result
	}

	maxScore := result[0].Score
	maxCount := 1
	maxIndex := 0

	for i := 1; i < len(result); i++ {
		score := result[i].Score
		switch {
		case score > maxScore:
			maxScore = score
			maxCount = 1
			maxIndex = i
		case score == maxScore:
			maxCount++
		}
	}

	for i := range result {
		result[i].Won = false
	}

	if maxCount == 1 {
		result[maxIndex].Won = true
	}

	return result
}
