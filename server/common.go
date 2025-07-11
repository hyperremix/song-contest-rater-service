package server

import (
	"github.com/hyperremix/song-contest-rater-service/event"
)

var broker = event.NewBroker()
