package devtools

func PlayerShapeID(shipType string) string {
	if shipType == "" {
		shipType = "v_wing"
	}
	return "player:" + shipType
}

func AsteroidShapeID(variant int) string {
	return "asteroid:" + itoa(variant)
}

func BulletShapeID() string {
	return "bullet"
}

func PickupShapeID(pickupType string) string {
	return "pickup:" + pickupType
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}

	negative := n < 0
	if negative {
		n = -n
	}

	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}

	if negative {
		i--
		buf[i] = '-'
	}

	return string(buf[i:])
}
