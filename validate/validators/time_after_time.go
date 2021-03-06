package validators

import (
	"fmt"
	"time"

	"github.com/gobuffalo/validate"
)

// TimeAfterTime validator
type TimeAfterTime struct {
	FirstName  string
	FirstTime  time.Time
	SecondName string
	SecondTime time.Time
}

// IsValid checks if FirstTime is after SecondTime
func (v *TimeAfterTime) IsValid(errors *validate.Errors) {
	if v.FirstTime.UnixNano() < v.SecondTime.UnixNano() {
		errors.Add(GenerateKey(v.FirstName), fmt.Sprintf("%s must be after %s.", v.FirstName, v.SecondName))
	}
}
