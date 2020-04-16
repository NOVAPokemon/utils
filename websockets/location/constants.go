package location

import "time"

const (
	UpdateCooldown          = 3 * time.Second
	UpdateCooldownInSeconds = 3
	Timeout                 = 3 * UpdateCooldown
	PingCooldown            = 1 * time.Second

	UpdateLocation = "UPDATE_LOCATION"
	Gyms           = "GYMS"
)
