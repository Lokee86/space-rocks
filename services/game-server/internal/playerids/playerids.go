package playerids

import "fmt"

const MaxPlayers = 8

func Format(number int) string {
	return fmt.Sprintf("Player-%d", number)
}
