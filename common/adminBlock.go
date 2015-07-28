// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package common

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"sync"
)

// Administrative Chain
type AdminChain struct {
	ChainID *Hash
	Name    [][]byte

	NextBlock       *AdminBlock
	NextBlockHeight uint32
	BlockMutex      sync.Mutex
}

// Administrative Block
// This is a special block which accompanies this Directory Block.
// It contains the signatures and organizational data needed to validate previous and future Directory Blocks.
// This block is included in the DB body. It appears there with a pair of the Admin AdminChainID:SHA256 of the block.
// For more details, please go to:
// https://github.com/FactomProject/FactomDocs/blob/master/factomDataStructureDetails.md#administrative-block
type AdminBlock struct {
	//Marshalized
	Header    *ABlockHeader
	ABEntries []ABEntry //Interface

	//Not Marshalized
	ABHash *Hash
}

// Create an empty Admin Block
func CreateAdminBlock(chain *AdminChain, prev *AdminBlock, cap uint) (b *AdminBlock, err error) {
	if prev == nil && chain.NextBlockHeight != 0 {
		return nil, errors.New("Previous block cannot be nil")
	} else if prev != nil && chain.NextBlockHeight == 0 {
		return nil, errors.New("Origin block cannot have a parent block")
	}

	b = new(AdminBlock)

	b.Header = new(ABlockHeader)
	b.Header.AdminChainID = chain.ChainID

	if prev == nil {
		b.Header.PrevFullHash = NewHash()
	} else {

		if prev.ABHash == nil {
			prev.BuildABHash()
		}
		b.Header.PrevFullHash = prev.ABHash
	}

	b.Header.DBHeight = chain.NextBlockHeight
	b.ABEntries = make([]ABEntry, 0, cap)

	return b, err
}

// Build the sha hash for the admin block
func (b *AdminBlock) BuildABHash() (err error) {

	binaryAB, _ := b.MarshalBinary()
	b.ABHash = Sha512Half(binaryAB)

	return
}

// Add an Admin Block entry to the block
func (b *AdminBlock) AddABEntry(e ABEntry) (err error) {
	b.ABEntries = append(b.ABEntries, e)
	return
}

// Add the end-of-minute marker into the admin block
func (b *AdminBlock) AddEndOfMinuteMarker(eomType byte) (err error) {
	eOMEntry := &EndOfMinuteEntry{
		entryType: TYPE_MINUTE_NUM,
		EOM_Type:  eomType}

	b.AddABEntry(eOMEntry)

	return
}

// Write out the AdminBlock to binary.
func (b *AdminBlock) MarshalBinary() (data []byte, err error) {
	var buf bytes.Buffer

	data, _ = b.Header.MarshalBinary()
	buf.Write(data)

	for i := uint32(0); i < b.Header.MessageCount; i++ {
		data, _ := b.ABEntries[i].MarshalBinary()
		buf.Write(data)
	}
	return buf.Bytes(), err
}

// Admin Block size
func (b *AdminBlock) MarshalledSize() uint64 {
	var size uint64 = 0

	size += b.Header.MarshalledSize()

	for _, entry := range b.ABEntries {
		size += entry.MarshalledSize()
	}

	return size
}

func (b *AdminBlock) UnmarshalBinaryData(data []byte) (newData []byte, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("Error unmarshalling: %v", r)
		}
	}()
	newData = data
	h := new(ABlockHeader)
	newData, err = h.UnmarshalBinaryData(newData)
	if err != nil {
		return
	}
	b.Header = h

	b.ABEntries = make([]ABEntry, b.Header.MessageCount)
	for i := uint32(0); i < b.Header.MessageCount; i++ {
		if newData[0] == TYPE_DB_SIGNATURE {
			b.ABEntries[i] = new(DBSignatureEntry)
		} else if newData[0] == TYPE_MINUTE_NUM {
			b.ABEntries[i] = new(EndOfMinuteEntry)
		}
		newData, err = b.ABEntries[i].UnmarshalBinaryData(newData)
		if err != nil {
			return
		}
	}
	return
}

// Read in the binary into the Admin block.
func (b *AdminBlock) UnmarshalBinary(data []byte) (err error) {
	_, err = b.UnmarshalBinaryData(data)
	return
}

// Read in the binary into the Admin block.
func (b *AdminBlock) GetDBSignature() ABEntry {

	for i := uint32(0); i < b.Header.MessageCount; i++ {
		if b.ABEntries[i].Type() == TYPE_DB_SIGNATURE {
			return b.ABEntries[i]
		}
	}

	return nil
}

// Admin Block Header
type ABlockHeader struct {
	AdminChainID *Hash
	PrevFullHash *Hash
	DBHeight     uint32

	HeaderExpansionSize uint64
	HeaderExpansionArea []byte

	MessageCount uint32
	BodySize     uint32
}

// Write out the ABlockHeader to binary.
func (b *ABlockHeader) MarshalBinary() (data []byte, err error) {
	var buf bytes.Buffer

	data, err = b.AdminChainID.MarshalBinary()
	if err != nil {
		return nil, err
	}
	buf.Write(data)

	data, err = b.PrevFullHash.MarshalBinary()
	if err != nil {
		return nil, err
	}
	buf.Write(data)

	binary.Write(&buf, binary.BigEndian, b.DBHeight)

	EncodeVarInt(&buf, b.HeaderExpansionSize)
	buf.Write(b.HeaderExpansionArea)

	binary.Write(&buf, binary.BigEndian, b.MessageCount)
	binary.Write(&buf, binary.BigEndian, b.BodySize)

	return buf.Bytes(), err
}

func (b *ABlockHeader) MarshalledSize() uint64 {
	var size uint64 = 0

	size += uint64(HASH_LENGTH)                         //AdminChainID
	size += uint64(HASH_LENGTH)                         //PrevFullHash
	size += 4                                           //DBHeight
	size += uint64(VarIntLength(b.HeaderExpansionSize)) //HeaderExpansionSize
	size += b.HeaderExpansionSize                       //HeadderExpansionArea
	size += 4                                           //MessageCount
	size += 4                                           //BodySize

	return size
}

func (b *ABlockHeader) UnmarshalBinaryData(data []byte) (newData []byte, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("Error unmarshalling: %v", r)
		}
	}()
	newData = data
	b.AdminChainID = new(Hash)
	newData, err = b.AdminChainID.UnmarshalBinaryData(newData)
	if err != nil {
		return
	}

	b.PrevFullHash = new(Hash)
	newData, err = b.PrevFullHash.UnmarshalBinaryData(newData)
	if err != nil {
		return
	}

	b.DBHeight, newData = binary.BigEndian.Uint32(newData[0:4]), newData[4:]

	b.HeaderExpansionSize, newData = DecodeVarInt(newData)
	b.HeaderExpansionArea, newData = newData[:b.HeaderExpansionSize], newData[b.HeaderExpansionSize:]

	b.MessageCount, newData = binary.BigEndian.Uint32(newData[0:4]), newData[4:]
	b.BodySize, newData = binary.BigEndian.Uint32(newData[0:4]), newData[4:]

	return
}

// Read in the binary into the ABlockHeader.
func (b *ABlockHeader) UnmarshalBinary(data []byte) (err error) {
	_, err = b.UnmarshalBinaryData(data)
	return
}

// Generic admin block entry type
type ABEntry interface {
	Type() byte
	MarshalBinary() ([]byte, error)
	MarshalledSize() uint64
	UnmarshalBinary(data []byte) (err error)
	UnmarshalBinaryData(data []byte) (newData []byte, err error)
}

// DB Signature Entry -------------------------
type DBSignatureEntry struct {
	ABEntry              //interface
	entryType            byte
	IdentityAdminChainID *Hash
	PubKey               PublicKey
	PrevDBSig            *[64]byte
}

// Create a new DB Signature Entry
func NewDBSignatureEntry(identityAdminChainID *Hash, sig Signature) (e *DBSignatureEntry) {
	e = new(DBSignatureEntry)
	e.entryType = TYPE_DB_SIGNATURE
	e.IdentityAdminChainID = identityAdminChainID
	e.PubKey = sig.Pub
	e.PrevDBSig = sig.Sig
	return
}

func (e *DBSignatureEntry) Type() byte {
	return e.entryType
}

func (e *DBSignatureEntry) MarshalBinary() (data []byte, err error) {
	var buf bytes.Buffer

	buf.Write([]byte{e.entryType})

	data, err = e.IdentityAdminChainID.MarshalBinary()
	if err != nil {
		return nil, err
	}
	buf.Write(data)

	_, err = buf.Write(e.PubKey.Key[:])
	if err != nil {
		return nil, err
	}

	_, err = buf.Write(e.PrevDBSig[:])
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func (e *DBSignatureEntry) MarshalledSize() uint64 {
	var size uint64 = 0
	size += 1 // Type (byte)
	size += uint64(HASH_LENGTH)
	size += uint64(HASH_LENGTH)
	size += uint64(SIG_LENGTH)

	return size
}

func (e *DBSignatureEntry) UnmarshalBinaryData(data []byte) (newData []byte, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("Error unmarshalling: %v", r)
		}
	}()
	newData = data
	e.entryType, newData = newData[0], newData[1:]

	e.IdentityAdminChainID = new(Hash)
	newData, err = e.IdentityAdminChainID.UnmarshalBinaryData(newData)
	if err != nil {
		return
	}

	e.PubKey.Key = new([HASH_LENGTH]byte)
	copy(e.PubKey.Key[:], newData[:HASH_LENGTH])
	newData = newData[HASH_LENGTH:]

	e.PrevDBSig = new([SIG_LENGTH]byte)
	copy(e.PrevDBSig[:], newData[:SIG_LENGTH])

	newData = newData[SIG_LENGTH:]

	return
}

func (e *DBSignatureEntry) UnmarshalBinary(data []byte) (err error) {
	_, err = e.UnmarshalBinaryData(data)
	return
}

type EndOfMinuteEntry struct {
	entryType byte
	EOM_Type  byte
}

func (m *EndOfMinuteEntry) Type() byte {
	return m.entryType
}

func (e *EndOfMinuteEntry) MarshalBinary() (data []byte, err error) {
	var buf bytes.Buffer

	buf.Write([]byte{e.entryType})

	buf.Write([]byte{e.EOM_Type})

	return buf.Bytes(), nil
}

func (e *EndOfMinuteEntry) MarshalledSize() uint64 {
	var size uint64 = 0
	size += 1 // Type (byte)
	size += 1 // EOM_Type (byte)

	return size
}

func (e *EndOfMinuteEntry) UnmarshalBinaryData(data []byte) (newData []byte, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("Error unmarshalling: %v", r)
		}
	}()
	newData = data

	e.entryType, newData = newData[0], newData[1:]
	e.EOM_Type, newData = newData[0], newData[1:]

	return
}

func (e *EndOfMinuteEntry) UnmarshalBinary(data []byte) (err error) {
	_, err = e.UnmarshalBinaryData(data)
	return
}
