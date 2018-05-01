package adminBlock

import (
	"fmt"

	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/factoid"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
)

// AddFactoidAddress Entry -------------------------
type AddFactoidAddress struct {
	AdminIDType     uint32 `json:"adminidtype"`
	IdentityChainID interfaces.IHash
	FactoidAddress  interfaces.IAddress
}

var _ interfaces.IABEntry = (*AddFactoidAddress)(nil)
var _ interfaces.BinaryMarshallable = (*AddFactoidAddress)(nil)

func (e *AddFactoidAddress) Init() {
	e.AdminIDType = uint32(e.Type())
	if e.IdentityChainID == nil {
		e.IdentityChainID = primitives.NewZeroHash()
	}

	if e.FactoidAddress == nil {
		e.FactoidAddress = &factoid.Address{*(primitives.NewZeroHash().(*primitives.Hash))}
	}
}

func (a *AddFactoidAddress) IsSameAs(b *AddFactoidAddress) bool {
	if a.Type() != b.Type() {
		return false
	}

	if !a.IdentityChainID.IsSameAs(b.IdentityChainID) {
		return false
	}

	if !a.FactoidAddress.IsSameAs(b.FactoidAddress) {
		return false
	}

	return true
}

func (e *AddFactoidAddress) String() string {
	e.Init()
	var out primitives.Buffer
	out.WriteString(fmt.Sprintf("    E: %20s -- %17s %8x %12s %s",
		"AddAuditServer",
		"IdentityChainID", e.IdentityChainID.Bytes()[3:6],
		"Address", e.FactoidAddress.String()))
	return (string)(out.DeepCopyBytes())
}

func (c *AddFactoidAddress) UpdateState(state interfaces.IState) error {
	c.Init()
	//state.AddAuditServer(c.DBHeight, c.IdentityChainID)
	state.UpdateAuthorityFromABEntry(c)

	return nil
}

func NewAddFactoidAddress(chainID interfaces.IHash, add interfaces.IAddress) (e *AddFactoidAddress) {
	e = new(AddFactoidAddress)
	e.Init()
	e.IdentityChainID = chainID
	e.FactoidAddress = add
	return
}

func (e *AddFactoidAddress) Type() byte {
	return constants.TYPE_ADD_FACTOID_ADDRESS
}

func (e *AddFactoidAddress) SortedIdentity() interfaces.IHash {
	return e.IdentityChainID
}

func (e *AddFactoidAddress) MarshalBinary() ([]byte, error) {
	e.Init()
	var buf primitives.Buffer

	err := buf.PushByte(e.Type())
	if err != nil {
		return nil, err
	}

	// Need the size of the body
	var bodybuf primitives.Buffer
	err = bodybuf.PushIHash(e.IdentityChainID)
	if err != nil {
		return nil, err
	}

	err = bodybuf.PushBinaryMarshallable(e.FactoidAddress)
	if err != nil {
		return nil, err
	}
	// end body

	err = buf.PushVarInt(uint64(bodybuf.Len()))
	if err != nil {
		return nil, err
	}

	err = buf.Push(bodybuf.Bytes())
	if err != nil {
		return nil, err
	}

	return buf.DeepCopyBytes(), nil
}

func (e *AddFactoidAddress) UnmarshalBinaryData(data []byte) ([]byte, error) {
	buf := primitives.NewBuffer(data)
	e.Init()

	b, err := buf.PopByte()
	if err != nil {
		return nil, err
	}

	if b != e.Type() {
		return nil, fmt.Errorf("Invalid Entry type")
	}

	bl, err := buf.PopVarInt()
	if err != nil {
		return nil, err
	}

	body := make([]byte, bl)
	n, err := buf.Read(body)
	if err != nil {
		return nil, err
	}

	if uint64(n) != bl {
		return nil, fmt.Errorf("Expected to read %d bytes, but got %d", bl, n)
	}

	bodyBuf := primitives.NewBuffer(body)

	if uint64(n) != bl {
		return nil, fmt.Errorf("Unable to unmarshal body")
	}

	e.IdentityChainID, err = bodyBuf.PopIHash()
	if err != nil {
		return nil, err
	}

	err = bodyBuf.PopBinaryMarshallable(e.FactoidAddress)
	if err != nil {
		return nil, err
	}

	return buf.DeepCopyBytes(), nil
}

func (e *AddFactoidAddress) UnmarshalBinary(data []byte) (err error) {
	_, err = e.UnmarshalBinaryData(data)
	return
}

func (e *AddFactoidAddress) JSONByte() ([]byte, error) {
	e.AdminIDType = uint32(e.Type())
	return primitives.EncodeJSON(e)
}

func (e *AddFactoidAddress) JSONString() (string, error) {
	e.AdminIDType = uint32(e.Type())
	return primitives.EncodeJSONString(e)
}

func (e *AddFactoidAddress) IsInterpretable() bool {
	return false
}

func (e *AddFactoidAddress) Interpret() string {
	return ""
}

func (e *AddFactoidAddress) Hash() interfaces.IHash {
	bin, err := e.MarshalBinary()
	if err != nil {
		panic(err)
	}
	return primitives.Sha(bin)
}
