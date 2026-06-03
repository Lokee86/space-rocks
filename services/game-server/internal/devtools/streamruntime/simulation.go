package streamruntime

func StepGameRuntime(gameInstance interface{ DevtoolsStepContinuousBulletStreams(float64) }, delta float64) {
	gameInstance.DevtoolsStepContinuousBulletStreams(delta)
}
