package leader

import (
	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/fnode"
	"github.com/FactomProject/factomd/modules/event"
	"github.com/FactomProject/factomd/pubsub"
	"github.com/FactomProject/factomd/worker"
)

type Pub struct {
	MsgOut pubsub.IPublisher
}

type Sub struct {
	MsgInput       *pubsub.SubChannel
	MovedToHeight  *pubsub.SubChannel
	BalanceChanged *pubsub.SubChannel
	DBlockCreated  *pubsub.SubChannel
	EomTicker      *pubsub.SubChannel
	LeaderConfig   *pubsub.SubChannel
	// TODO: add InternalAuthoritySet listener - probably add an update auth set msg instead of listening to all
	// Eventually also track swapping audits and leaders AddRemoveServer Messages (come in quads)
	// currently sent to election process

}

// block level events
type Events struct {
	Config         *event.LeaderConfig //
	*event.DBHT                        // from move-to-ht
	*event.Balance                     // REVIEW: does this relate to a specific VM
	*event.Directory
	*event.Ack // record of last sent ack by leader
	*event.LeaderConfig
}

func mkChan() *pubsub.SubChannel {
	return pubsub.SubFactory.Channel(50)
}

func (l *Leader) Start(w *worker.Thread) {

	w.Run(l.EOMTimer)

	w.Spawn(func(w *worker.Thread) {
		w.Init(&w.Name, "LeaderThread")
		w.OnReady(l.Ready)
		w.OnRun(l.Run)

		l.Pub.MsgOut = pubsub.PubFactory.Threaded(100).Publish(
			pubsub.GetPath("FNode0", event.Path.LeaderMsgOut),
		)
		go l.Pub.MsgOut.Start()

		l.Sub.MovedToHeight = mkChan()
		l.Sub.MsgInput = mkChan()
		l.Sub.BalanceChanged = mkChan()
		l.Sub.DBlockCreated = mkChan()
		l.Sub.EomTicker = mkChan()
		l.Sub.LeaderConfig = mkChan()
	})
}

func (l *Leader) Ready() {
	node0 := fnode.Get(0).State.GetFactomNodeName() // get follower name

	{ // temporary hooks for leader thread development
		// KLUDGE publish to Fnode01 bypassing networkOut
		// snoop on valid message inputs
		l.Sub.MsgInput.Subscribe(pubsub.GetPath(node0, "bmv", "rest"))
	}

	// subscribe to internal events
	l.Sub.MovedToHeight.Subscribe(pubsub.GetPath(node0, event.Path.Seq))
	l.Sub.DBlockCreated.Subscribe(pubsub.GetPath(node0, event.Path.Directory))
	l.Sub.BalanceChanged.Subscribe(pubsub.GetPath(node0, event.Path.Bank))
}

func (l *Leader) Run() {
	for {
		select {
		case v := <-l.MsgInput.Updates:
			m := v.(interfaces.IMsg)
			if constants.NeedsAck(m.Type()) {
				l.sendAck(m)
			}
		case v := <-l.MovedToHeight.Updates:
			l.DBHT = v.(*event.DBHT)
			log.LogPrintf("leader.txt", "SeqChange: %v", v)
			l.seqChanged()
		}
	}
}

func (l*Leader) seqChanged() {
	if l.DBHT.Minute != 0 {
		return
	}
	{ // possibly shut down this leader thread or maybe unsubscribe to events
		select {
		//case v := <-l.NewAuthoritySet
		case v := <-l.Sub.LeaderConfig.Updates:
			l.Config = v.(*event.LeaderConfig)
			// TODO: handle demotion/brainswap
		default:
		}
	}
	{
		v := <-l.Sub.BalanceChanged.Updates
		l.Balance = v.(*event.Balance)
		log.LogPrintf("leader.txt", "BalChange: %v", v)
	}
	{
		v := <-l.Sub.DBlockCreated.Updates
		l.Directory = v.(*event.Directory)
		log.LogPrintf("leader.txt", "Directory: %v", v)
	}
	l.VMIndex = 0 // KLUDGE hard coded for single leader
	l.SendDBSig()
}

