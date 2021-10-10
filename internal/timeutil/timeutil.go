package timeutil

import (
	"fmt"
	"math"
	"time"
)

// FormatDuration formats duration from timestamp to now.
// Based in part on the javascript library date-fns:
// https://github.com/date-fns/date-fns/blob/master/src/formatDistance/index.ts
func FormatDuration(timestamp int64) string {
	const minutesInDay = 1440
	const minutesInAlmostTwoDays = 2520
	const minutesInMonth = 43200    // 30 days
	const minutesInAvgMonth = 43830 // 365.25/12 days
	const minutesInTwoMonths = 86400

	minutes := int64(math.Round(float64(time.Now().Unix()*1000-timestamp) / 1000 / 60))
	months := int64(math.Round(float64(minutes) / minutesInAvgMonth))

	switch {
	case minutes == 0:
		return "less than a minute"

	case minutes < 2:
		return "1 minute"

	case minutes < 45:
		return fmt.Sprintf("%d minutes", minutes)

	case minutes < 90:
		return "about 1 hour"

	case minutes < minutesInDay:
		return fmt.Sprintf("about %.0f hours", float64(minutes)/60)

	case minutes < minutesInAlmostTwoDays:
		return "1 day"

	case minutes < minutesInMonth:
		return fmt.Sprintf("%.0f days", float64(minutes)/minutesInDay)

	case minutes < 1.5*minutesInMonth:
		return "about 1 month"

	case minutes < minutesInTwoMonths:
		return "about 2 months"

	case months < 12:
		return fmt.Sprintf("%d months", months)

	case months < 15:
		return "about 1 year"

	case months < 21:
		return "over 1 year"

	default:
		return fmt.Sprintf("about %.0f years", float64(months)/12)
	}
}
