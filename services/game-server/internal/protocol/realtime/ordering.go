package realtime

import "sort"

type OrderedRecordKey struct {
	Lane   string
	Family string
	Kind   string
	ID     string
}

func SortOrderedRecordKeys(keys []OrderedRecordKey) {
	sort.SliceStable(keys, func(i, j int) bool {
		left := keys[i]
		right := keys[j]
		if left.Lane != right.Lane {
			return left.Lane < right.Lane
		}
		if left.Family != right.Family {
			return left.Family < right.Family
		}
		if left.Kind != right.Kind {
			return left.Kind < right.Kind
		}
		return left.ID < right.ID
	})
}
