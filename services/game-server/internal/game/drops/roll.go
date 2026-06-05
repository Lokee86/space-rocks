package drops

func (tables Tables) Roll(tableID string, source Source, roll Roll) []Result {
	table, ok := tables.ByID[tableID]
	if !ok {
		return nil
	}
	if table.SourceType != source.Type {
		return nil
	}

	results := make([]Result, 0, min(table.MaxDropsPerSource, len(table.Entries)))
	rollIndex := 0
	for _, entry := range table.Entries {
		if source.Size < entry.MinSourceSize || source.Size > entry.MaxSourceSize {
			continue
		}

		rollValue := 1.0
		if rollIndex < len(roll.Values) {
			rollValue = roll.Values[rollIndex]
		}
		rollIndex++

		if rollValue >= entry.Chance {
			continue
		}

		results = append(results, Result{
			TableID:    table.ID,
			PickupType: entry.PickupType,
			X:          source.X,
			Y:          source.Y,
		})
		if table.DropMode == DropModeSingle || len(results) >= table.MaxDropsPerSource {
			break
		}
	}

	return results
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
