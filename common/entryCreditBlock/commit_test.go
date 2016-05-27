package entryCreditBlock_test

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"testing"

	ed "github.com/FactomProject/ed25519"
	. "github.com/FactomProject/factomd/common/entryCreditBlock"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
	"github.com/FactomProject/factomd/testHelper"
)

var (
	_ = fmt.Sprint("testing")
)

func TestCommitEntryMarshal(t *testing.T) {
	fmt.Printf("---\nTestCommitEntryMarshal\n---\n")

	ce := NewCommitEntry()

	// test MarshalBinary on a zeroed CommitEntry
	if p, err := ce.MarshalBinary(); err != nil {
		t.Error(err)
	} else if z := make([]byte, CommitEntrySize); string(p) != string(z) {
		t.Errorf("Marshal failed on zeroed CommitEntry")
	}

	// build a CommitEntry for testing
	ce.Version = 0
	ce.MilliTime = (*primitives.ByteSlice6)(&[6]byte{1, 1, 1, 1, 1, 1})
	p, _ := hex.DecodeString("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")
	ce.EntryHash.SetBytes(p)
	ce.Credits = 1

	// make a key and sign the msg
	if pub, privkey, err := ed.GenerateKey(rand.Reader); err != nil {
		t.Error(err)
	} else {
		ce.ECPubKey = (*primitives.ByteSlice32)(pub)
		ce.Sig = (*primitives.ByteSlice64)(ed.Sign(privkey, ce.CommitMsg()))
	}

	// marshal and unmarshal the commit and see if it matches
	ce2 := NewCommitEntry()
	if p, err := ce.MarshalBinary(); err != nil {
		t.Error(err)
	} else {
		t.Logf("%x\n", p)
		ce2.UnmarshalBinary(p)
	}

	if !ce2.IsValid() {
		t.Errorf("signature did not match after unmarshalbinary")
	}
}

func TestCommitMarshalUnmarshal(t *testing.T) {
	blocks := testHelper.CreateFullTestBlockSet()
	for _, block := range blocks {
		for _, tx := range block.ECBlock.GetEntries() {
			h1, err := tx.MarshalBinary()
			if err != nil {
				t.Errorf("Error marshalling - %v", err)
			}
			var h2 []byte
			var e interfaces.BinaryMarshallable
			switch tx.ECID() {
			case ECIDChainCommit:
				e = new(CommitChain)
				break
			case ECIDEntryCommit:
				e = new(CommitEntry)
				break
			case ECIDBalanceIncrease:
				e = new(IncreaseBalance)
				break
			case ECIDMinuteNumber:
				e = new(MinuteNumber)
				break
			case ECIDServerIndexNumber:
				e = new(ServerIndexNumber)
				break
			default:
				t.Error("Wrong ECID")
				break
			}

			h2, err = e.UnmarshalBinaryData(h1)
			if err != nil {
				t.Errorf("Error unmarshalling - %v", err)
				continue
			}
			if len(h2) > 0 {
				t.Errorf("Leftovers from unmarshalling - %x", h2)
			}
			h2, err = e.MarshalBinary()
			if err != nil {
				t.Errorf("Error marshalling2 - %v", err)
				continue
			}

			if primitives.AreBytesEqual(h1, h2) == false {
				t.Error("ECEntries are not identical - %x vs %x", h1, h2)
			}
		}
	}
}
