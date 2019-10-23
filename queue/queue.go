package queue

import (
	"time"

	"github.com/FactomProject/factomd/common"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/generated"
	"github.com/FactomProject/factomd/telemetry"
	"github.com/FactomProject/factomd/worker"
)

//FactomGenerate template accountedqueue typename Queue_IMsg type interfaces.IMsg
//FactomGenerate template accountedqueue import github.com/FactomProject/factomd/common/interfaces
//FactomGenerate template accountedqueue import github.com/FactomProject/factomd/common
//FactomGenerate template accountedqueue import github.com/FactomProject/factomd/telemetry

//FactomGenerate template accountedqueue_test typename Queue_IMsg type interfaces.IMsg testelement new(messages.Bounce) import time,github.com/FactomProject/factomd/common/interfaces
//FactomGenerate template accountedqueue_test import github.com/FactomProject/factomd/common
//FactomGenerate template accountedqueue_test import github.com/FactomProject/factomd/telemetry
//FactomGenerate template accountedqueue_test import github.com/FactomProject/factomd/worker
//FactomGenerate template accountedqueue_test import testing

type MsgQueue struct {
	generated.Queue_IMsg
	Package string
	Thread  *worker.Thread
}

func NewMsgQueue(thread *worker.Thread, parent common.NamedObject, packagename string, name string, size int) *MsgQueue {
	q := new(MsgQueue).Init(thread, packagename, parent, name, size)
	return q
}

func (q *MsgQueue) Init(thread *worker.Thread, packagename string, parent common.NamedObject, name string, size int) *MsgQueue {
	q.Thread = thread
	q.Package = packagename
	q.Queue_IMsg.Init(parent, name, size)
	return q
}

// construct gauge w/ proper labels
func (q *MsgQueue) Metric(msg interfaces.IMsg) telemetry.Gauge {
	label := "nil"
	if msg != nil {
		label = msg.Label()
	}

	return telemetry.ChannelSize.WithLabelValues(q.Package, q.GetPath()+label, "thread", "current")
}

// construct counter for tracking totals
func (q *MsgQueue) TotalMetric(msg interfaces.IMsg) telemetry.Counter {
	label := "nil"
	if msg != nil {
		label = msg.Label()
	}
	return telemetry.TotalCounter.WithLabelValues(q.Package, q.GetPath()+label, "thread", "total")
}

// construct counter for intermittent polling of queue size
func (q *MsgQueue) PollMetric() telemetry.Gauge {
	return telemetry.ChannelSize.WithLabelValues(q.Package, q.GetPath(), "thread", "aggregate")
}

// add metric to poll size of queue
func (q *MsgQueue) RegisterPollMetric() {
	q.Thread.RegisterMetric(func(poll *time.Ticker, exit chan bool) {
		gauge := q.PollMetric()

		for {
			select {
			case <-exit:
				return
			case <-poll.C:
				gauge.Set(float64(q.Length()))
			}
		}
	})
}

// Enqueue adds item to channel and instruments based on type
func (q *MsgQueue) Enqueue(m interfaces.IMsg) {
	q.Queue_IMsg.Enqueue(m)
	q.TotalMetric(m).Inc()
	q.Metric(m).Inc()
}

// Dequeue removes an item from channel and instruments based on type.
// Returns nil if nothing in // queue
func (q *MsgQueue) Dequeue() interfaces.IMsg {
	v := q.Queue_IMsg.Dequeue()
	if v != nil {
		q.Metric(v).Dec()
		return v
	}
	return nil
}

// Dequeue removes an item from channel and instruments based on type.
func (q *MsgQueue) BlockingDequeue() interfaces.IMsg {
	v := q.Queue_IMsg.BlockingDequeue()
	q.Metric(v).Dec()
	return v
}
