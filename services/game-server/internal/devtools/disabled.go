package devtools

func ShouldHandleCommand(packetType string) bool {
	return IsCommandType(packetType) && Enabled()
}
