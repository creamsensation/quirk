package quirk

import "time"

type subscription func(query string, duration time.Duration)
