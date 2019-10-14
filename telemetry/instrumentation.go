package telemetry

var (
	// Top Level Metrics
	MapSize = NewGaugeVec(
		"queue_size",
		"common metric for all queues",
		[]string{"package", "name", "thread", "message"},
	)

	ChannelSize = NewGaugeVec(
		"channel_size",
		"common metric for all channels",
		[]string{"package", "name", "thread", "message"},
	)
)
