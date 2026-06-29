package realtime

type Priority string

const (
	PriorityCritical Priority = "critical"
	PriorityHigh     Priority = "high"
	PriorityMedium   Priority = "medium"
	PriorityLow      Priority = "low"
	PriorityDebug    Priority = "debug"
)

type DeliveryClass string

const (
	DeliveryClassRequired        DeliveryClass = "required"
	DeliveryClassEventOnce       DeliveryClass = "event_once"
	DeliveryClassHotSupersedable  DeliveryClass = "hot_supersedable"
	DeliveryClassDeferrable      DeliveryClass = "deferrable"
	DeliveryClassDebugOnly       DeliveryClass = "debug_only"
)
