package pubsub

var SubFactory subFactory

type subFactory struct{}

func (subFactory) Channel(buffer int) *SubChannel { return NewSubChannel(buffer) }
func (subFactory) Value() *SubValue               { return NewSubValue() }
func (subFactory) Counter() *SubCounter           { return NewSubCounter() }

var PubFactory pubFactory

type pubFactory struct{}

func (pubFactory) Base() *PubBase                       { return new(PubBase) }
func (pubFactory) Multi(buffer int) *PubSimpleMulti     { return NewPubMulti(buffer) }
func (pubFactory) RoundRobin(buffer int) *PubRoundRobin { return NewPubRoundRobin(buffer) }
func (pubFactory) Threaded(buffer int) *PubThreaded     { return NewPubThreaded(buffer) }
