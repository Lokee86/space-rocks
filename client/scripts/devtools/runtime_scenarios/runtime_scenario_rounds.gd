extends RefCounted
class_name RuntimeScenarioRounds


static func expand(rounds: Array) -> Array:
	var expanded: Array = []
	for raw_round in rounds:
		if !(raw_round is Dictionary):
			expanded.append(raw_round)
			continue
		var round: Dictionary = raw_round
		var repeat_count := maxi(int(round.get("repeat", 1)), 1)
		var base_name := str(round.get("name", "round"))
		for repeat_index in range(repeat_count):
			var expanded_round: Dictionary = round.duplicate(true)
			expanded_round.erase("repeat")
			if repeat_count > 1:
				expanded_round["name"] = "%s-%03d" % [base_name, repeat_index + 1]
			expanded.append(expanded_round)
	return expanded
