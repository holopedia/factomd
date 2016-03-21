// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package messages

//https://docs.google.com/spreadsheets/d/1wy9JDEqyM2uRYhZ6Y1e9C3hIDm2prIILebztQ5BGlr8/edit#gid=1997221100

import (
	"fmt"
	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/interfaces"
)

func UnmarshalMessage(data []byte) (interfaces.IMsg, error) {
	if data == nil {
		return nil, fmt.Errorf("No data provided")
	}
	if len(data) == 0 {
		return nil, fmt.Errorf("No data provided")
	}
	messageType := int(data[0])
	var msg interfaces.IMsg
	switch messageType {
	case constants.EOM_MSG:
		msg = new(EOM)
	case constants.ACK_MSG:
		msg = new(Ack)
	case constants.AUDIT_SERVER_FAULT_MSG:
		msg = new(AuditServerFault)
	case constants.COMMIT_CHAIN_MSG:
		msg = new(CommitChainMsg)
	case constants.COMMIT_ENTRY_MSG:
		msg = new(CommitEntryMsg)
	case constants.DIRECTORY_BLOCK_SIGNATURE_MSG:
		msg = new(DirectoryBlockSignature)
	case constants.EOM_TIMEOUT_MSG:
		msg = new(EOMTimeout)
	case constants.FACTOID_TRANSACTION_MSG:
		msg = new(FactoidTransaction)
	case constants.HEARTBEAT_MSG:
		msg = new(Heartbeat)
	case constants.INVALID_ACK_MSG:
		msg = new(InvalidAck)
	case constants.INVALID_DIRECTORY_BLOCK_MSG:
		msg = new(InvalidDirectoryBlock)
	case constants.MISSING_ACK_MSG:
		msg = new(MissingAck)
	case constants.REVEAL_ENTRY_MSG:
		msg = new(RevealEntryMsg)
	case constants.REQUEST_BLOCK_MSG:
		msg = new(RequestBlock)
	case constants.SIGNATURE_TIMEOUT_MSG:
		msg = new(SignatureTimeout)
	case constants.DBSTATE_MISSING_MSG:
		msg = new(DBStateMissing)
	case constants.DBSTATE_MSG:
		msg = new(DBStateMsg)
	default:
		return nil, fmt.Errorf("Unknown message type %d %x", messageType, data[0])
	}

	err := msg.UnmarshalBinary(data[:])
	if err != nil {
		return nil, err
	}
	return msg, nil

}

type Signable interface {
	Sign(interfaces.Signer) error
	MarshalForSignature() ([]byte, error)
	GetSignature() interfaces.IFullSignature
}

func SignSignable(s Signable, key interfaces.Signer) (interfaces.IFullSignature, error) {
	toSign, err := s.MarshalForSignature()
	if err != nil {
		return nil, err
	}
	sig := key.Sign(toSign)
	return sig, nil
}

func VerifyMessage(s Signable) (bool, error) {
	toSign, err := s.MarshalForSignature()
	if err != nil {
		return false, err
	}
	sig := s.GetSignature()
	return sig.Verify(toSign), nil
}
