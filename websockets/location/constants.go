package location

import "time"

const UpdateCooldown = 3 * time.Second
const Timeout = 3 * UpdateCooldown
const PingCooldown = 1 * time.Second
