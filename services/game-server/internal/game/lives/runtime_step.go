package lives

func (runtime *Runtime) Step(delta float64) {
	if delta <= 0 {
		return
	}

	for _, participant := range runtime.participants {
		if participant.respawnCooldown > 0 {
			participant.respawnCooldown = max(0, participant.respawnCooldown-delta)
		}
	}
}
