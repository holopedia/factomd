// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package primitives_test

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/interfaces"
	. "github.com/FactomProject/factomd/common/primitives"
	"github.com/FactomProject/factomd/common/primitives/random"
)

func junk(x string) {
	defer func() {
		LogNilHashBug(x)
	}()
}

func junk2(x string) {
	defer func() {
		LogNilHashBug(x)
	}()
}

func Example_LogNilHashBugOnce() {
	os.Stderr = os.Stdout
	junk("GotA")
	// Output: GotA. Called from goroutine 1 -/common/primitives/hash_test.go:25
}

func Example_LogNilHashBugMultiple() {
	os.Stderr = os.Stdout
	for i := 0; i < 5; i++ {
		junk2(fmt.Sprintf("Got%d", i))
	}
	// Output: Got0. Called from goroutine 1 -/common/primitives/hash_test.go:31
}

func TestUnmarshalNilHash(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Panic caught during the test - %v", r)
		}
	}()

	a := new(Hash)
	err := a.UnmarshalBinary(nil)
	if err == nil {
		t.Errorf("Error is nil when it shouldn't be")
	}

	err = a.UnmarshalBinary([]byte{})
	if err == nil {
		t.Errorf("Error is nil when it shouldn't be")
	}
}

func TestHashCopyAndIsEqual(t *testing.T) {
	for i := 0; i < 1000; i++ {
		h := random.RandByteSliceOfLen(constants.HASH_LENGTH)

		h1 := new(Hash)
		h2 := new(Hash)

		if h1.IsSameAs(h2) == false { // Out of the box, hashes should be equal
			t.Errorf("Hashes are not equal")
		}

		h1.SetBytes(h[:])

		if h1.IsSameAs(h2) == true { // Now they should not be equal
			t.Errorf("Hashes are equal")
		}

		h2.SetBytes(h[:])

		if h1.IsSameAs(h2) == false { // Back to equality!
			t.Errorf("Hashes are not equal")
		}

		hash2 := h1.Fixed()
		for i := range h {
			if h[i] != hash2[i] {
				t.Errorf("Hashes are not equal")
			}
		}
	}
}

func TestHashMarshalUnmarshal(t *testing.T) {
	for i := 0; i < 1000; i++ {
		h1 := RandomHash()

		b, err := h1.MarshalBinary()
		if err != nil {
			t.Errorf("Error: %v", err)
		}

		h2 := new(Hash)
		err = h2.UnmarshalBinary(b)
		if err != nil {
			t.Errorf("Error: %v", err)
		}

		if h1.String() != h2.String() {
			t.Errorf("Hashes are not equal - %v vs %v", h1.String(), h2.String())
		}
	}
}

//Test vectors: http://www.di-mgt.com.au/sha_testvectors.html
func TestSha(t *testing.T) {
	testVector := map[string]string{
		"abc": "ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad",
		"":    "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
		"abcdbcdecdefdefgefghfghighijhijkijkljklmklmnlmnomnopnopq":                                                         "248d6a61d20638b8e5c026930c3e6039a33ce45964ff2167f6ecedd419db06c1",
		"abcdefghbcdefghicdefghijdefghijkefghijklfghijklmghijklmnhijklmnoijklmnopjklmnopqklmnopqrlmnopqrsmnopqrstnopqrstu": "cf5b16a778af8380036ce59e7b0492370b249b11e8f07a51afac45037afee9d1",
	}

	for k, v := range testVector {
		answer, err := DecodeBinary(v)
		if err != nil {
			t.Error(err)
			t.FailNow()
		}
		hash := Sha([]byte(k))

		if bytes.Compare(hash.Bytes(), answer) != 0 {
			t.Errorf("Wrong SHA hash for %v", k)
		}
		if hash.String() != v {
			t.Errorf("Wrong SHA hash string for %v", k)
		}
	}
}

func TestSha512Half(t *testing.T) {
	testVector := map[string]string{
		"abc": "ddaf35a193617abacc417349ae20413112e6fa4e89a97ea20a9eeee64b55d39a",
		"":    "cf83e1357eefb8bdf1542850d66d8007d620e4050b5715dc83f4a921d36ce9ce",
		"abcdbcdecdefdefgefghfghighijhijkijkljklmklmnlmnomnopnopq":                                                         "204a8fc6dda82f0a0ced7beb8e08a41657c16ef468b228a8279be331a703c335",
		"abcdefghbcdefghicdefghijdefghijkefghijklfghijklmghijklmnhijklmnoijklmnopjklmnopqklmnopqrlmnopqrsmnopqrstnopqrstu": "8e959b75dae313da8cf4f72814fc143f8f7779c6eb9f7fa17299aeadb6889018",
	}

	for k, v := range testVector {
		answer, err := DecodeBinary(v)
		if err != nil {
			t.Error(err)
			t.FailNow()
		}
		hash := Sha512Half([]byte(k))

		if bytes.Compare(hash.Bytes(), answer) != 0 {
			t.Errorf("Wrong SHA512Half hash for %v", k)
		}
		if hash.String() != v {
			t.Errorf("Wrong SHA512Half hash string for %v", k)
		}
	}
}

func TestHashStrings(t *testing.T) {
	base := "ddaf35a193617abacc417349ae20413112e6fa4e89a97ea20a9eeee64b55d39a"
	hash, err := HexToHash(base)
	if err != nil {
		t.Error(err)
	}
	if hash.String() != base {
		t.Error("Invalid conversion to string")
	}

	text, err := hash.JSONByte()
	if err != nil {
		t.Error(err)
	}

	if string(text) != fmt.Sprintf("\"%v\"", base) {
		t.Errorf("JSONByte failed - %v vs %v", string(text), base)
	}

	str, err := hash.JSONString()
	if err != nil {
		t.Error(err)
	}

	if str != fmt.Sprintf("\"%v\"", base) {
		t.Errorf("JSONString failed - %v vs %v", string(text), base)
	}
}

func TestIsSameAs(t *testing.T) {
	base := "ddaf35a193617abacc417349ae20413112e6fa4e89a97ea20a9eeee64b55d39a"
	hash, err := HexToHash(base)
	if err != nil {
		t.Error(err)
	}
	hex, err := DecodeBinary(base)
	if err != nil {
		t.Error(err)
	}
	hash2, err := NewShaHash(hex)
	if err != nil {
		t.Error(err)
	}
	if hash.IsSameAs(hash2) == false {
		t.Error("Identical hashes not recognized as such")
	}

	hash3 := hash.Copy()
	if hash.IsSameAs(hash3) == false {
		t.Errorf("Copied hash is not identical")
	}
}

func TestHashMisc(t *testing.T) {
	base := "4040404040404040404040404040404040404040404040404040404040404040"
	hash, err := HexToHash(base)
	if err != nil {
		t.Error(err)
	}
	if hash.String() != base {
		t.Error("Error in String")
	}

	hash2, err := NewShaHashFromStr(base)
	if err != nil {
		t.Error(err)
	}

	if hash2.ByteString() != "@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@" {
		t.Errorf("Error in ByteString - received %v", hash2.ByteString())
	}

	h, err := hex.DecodeString(base)
	if err != nil {
		t.Error(err)
	}
	hash = NewHash(h)
	if hash.String() != base {
		t.Error("Error in NewHash")
	}

	//***********************

	if hash.IsSameAs(nil) != false {
		t.Error("Error in IsSameAs")
	}

	//***********************

	minuteHash, err := HexToHash("0000000000000000000000000000000000000000000000000000000000000001")
	if err != nil {
		t.Error(err)
	}
	if minuteHash.IsMinuteMarker() == false {
		t.Error("Error in IsMinuteMarker")
	}

	hash = NewZeroHash()
	if hash.String() != "0000000000000000000000000000000000000000000000000000000000000000" {
		t.Error("Error in NewZeroHash")
	}
}

func TestHashIsZero(t *testing.T) {
	strs := []string{
		"0000000000000000000000000000000000000000000000000000000000000001",
		"0000000000000000000000000000000000000000000000000000000000000002",
		"0000000000000000000000000000000000000000000000000000000000000003",
		"0000000000000000000000000000000000000000000000000000000000000004",
		"0000000000000000000000000000000000000000000000000000000000000005",
		"0000000000000000000000000000000000000000000000000000000000000006",
		"0000000000000000000000000000000000000000000000000000000000000007",
		"0000000000000000000000000000000000000000000000000000000000000008",
		"0000000000000000000000000000000000000000000000000000000000000009",
		"000000000000000000000000000000000000000000000000000000000000000a",
		"000000000000000000000000000000000000000000000000000000000000000b",
		"000000000000000000000000000000000000000000000000000000000000000c",
		"000000000000000000000000000000000000000000000000000000000000000d",
		"000000000000000000000000000000000000000000000000000000000000000e",
		"000000000000000000000000000000000000000000000000000000000000000f"}
	for _, str := range strs {
		h, err := NewShaHashFromStr(str)
		if err != nil {
			t.Error(err)
		}
		if h.IsZero() == true {
			t.Errorf("Non-zero hash is zero")
		}
	}

	h, err := NewShaHashFromStr("0000000000000000000000000000000000000000000000000000000000000000")
	if err != nil {
		t.Error(err)
	}
	if h.IsZero() == false {
		t.Errorf("Zero hash is non-zero")
	}

}

func TestIsMinuteMarker(t *testing.T) {
	strs := []string{
		"0000000000000000000000000000000000000000000000000000000000000000",
		"0000000000000000000000000000000000000000000000000000000000000001",
		"0000000000000000000000000000000000000000000000000000000000000002",
		"0000000000000000000000000000000000000000000000000000000000000003",
		"0000000000000000000000000000000000000000000000000000000000000004",
		"0000000000000000000000000000000000000000000000000000000000000005",
		"0000000000000000000000000000000000000000000000000000000000000006",
		"0000000000000000000000000000000000000000000000000000000000000007",
		"0000000000000000000000000000000000000000000000000000000000000008",
		"0000000000000000000000000000000000000000000000000000000000000009",
		"000000000000000000000000000000000000000000000000000000000000000a",
		"000000000000000000000000000000000000000000000000000000000000000b",
		"000000000000000000000000000000000000000000000000000000000000000c",
		"000000000000000000000000000000000000000000000000000000000000000d",
		"000000000000000000000000000000000000000000000000000000000000000e",
		"000000000000000000000000000000000000000000000000000000000000000f"}
	for _, str := range strs {
		hash, err := HexToHash(str)
		if err != nil {
			t.Errorf("%v", err)
		}
		if hash.IsMinuteMarker() == false {
			t.Errorf("Entry %v is not a minute marker!", str)
		}
	}
	strs = []string{
		"1000000000000000000000000000000000000000000000000000000000000000",
		"0200000000000000000000000000000000000000000000000000000000000000",
		"0030000000000000000000000000000000000000000000000000000000000000",
		"0004000000000000000000000000000000000000000000000000000000000000",
		"0000500000000000000000000000000000000000000000000000000000000000",
		"0000060000000000000000000000000000000000000000000000000000000000",
		"0000007000000000000000000000000000000000000000000000000000000000",
		"0000000800000000000000000000000000000000000000000000000000000000",
		"0000000090000000000000000000000000000000000000000000000000000000",
		"000000000a000000000000000000000000000000000000000000000000000000",
		"0000000000b00000000000000000000000000000000000000000000000000000",
		"00000000000c0000000000000000000000000000000000000000000000000000",
		"000000000000d000000000000000000000000000000000000000000000000000",
		"0000000000000e00000000000000000000000000000000000000000000000000",
		"00000000000000f0000000000000000000000000000000000000000000000000",
		"0000000000000001000000000000000000000000000000000000000000000000",
		"0000000000000000200000000000000000000000000000000000000000000000",
		"0000000000000000030000000000000000000000000000000000000000000000",
		"0000000000000000004000000000000000000000000000000000000000000000",
		"0000000000000000000500000000000000000000000000000000000000000000",
		"0000000000000000000060000000000000000000000000000000000000000000",
		"0000000000000000000007000000000000000000000000000000000000000000",
		"0000000000000000000000800000000000000000000000000000000000000000",
		"0000000000000000000000090000000000000000000000000000000000000000",
		"000000000000000000000000a000000000000000000000000000000000000000",
		"0000000000000000000000000b00000000000000000000000000000000000000",
		"00000000000000000000000000c0000000000000000000000000000000000000",
		"000000000000000000000000000d000000000000000000000000000000000000",
		"0000000000000000000000000000e00000000000000000000000000000000000",
		"00000000000000000000000000000f0000000000000000000000000000000000",
		"0000000000000000000000000000001000000000000000000000000000000000",
		"0000000000000000000000000000000200000000000000000000000000000000",
		"0000000000000000000000000000000030000000000000000000000000000000",
		"0000000000000000000000000000000004000000000000000000000000000000",
		"0000000000000000000000000000000000500000000000000000000000000000",
		"0000000000000000000000000000000000060000000000000000000000000000",
		"0000000000000000000000000000000000007000000000000000000000000000",
		"0000000000000000000000000000000000000800000000000000000000000000",
		"0000000000000000000000000000000000000090000000000000000000000000",
		"000000000000000000000000000000000000000a000000000000000000000000",
		"0000000000000000000000000000000000000000b00000000000000000000000",
		"00000000000000000000000000000000000000000c0000000000000000000000",
		"000000000000000000000000000000000000000000d000000000000000000000",
		"0000000000000000000000000000000000000000000e00000000000000000000",
		"00000000000000000000000000000000000000000000f0000000000000000000",
		"0000000000000000000000000000000000000000000001000000000000000000",
		"0000000000000000000000000000000000000000000000200000000000000000",
		"0000000000000000000000000000000000000000000000030000000000000000",
		"0000000000000000000000000000000000000000000000004000000000000000",
		"0000000000000000000000000000000000000000000000000500000000000000",
		"0000000000000000000000000000000000000000000000000060000000000000",
		"0000000000000000000000000000000000000000000000000007000000000000",
		"0000000000000000000000000000000000000000000000000000800000000000",
		"0000000000000000000000000000000000000000000000000000090000000000",
		"000000000000000000000000000000000000000000000000000000a000000000",
		"0000000000000000000000000000000000000000000000000000000b00000000",
		"00000000000000000000000000000000000000000000000000000000c0000000",
		"000000000000000000000000000000000000000000000000000000000d000000",
		"0000000000000000000000000000000000000000000000000000000000e00000",
		"00000000000000000000000000000000000000000000000000000000000f0000",
		"0000000000000000000000000000000000000000000000000000000000001000",
		"0000000000000000000000000000000000000000000000000000000000000200"}

	for _, str := range strs {
		hash, err := HexToHash(str)
		if err != nil {
			t.Errorf("%v", err)
		}
		if hash.IsMinuteMarker() == true {
			t.Errorf("Entry %v is a minute marker!", str)
		}

		text, err := hash.(*Hash).MarshalText()
		if err != nil {
			t.Errorf("%v", err)
		}
		if string(text) != str {
			t.Errorf("Invalid marshalled text")
		}
	}
}

func TestStringUnmarshaller(t *testing.T) {
	for i := 0; i < 1000; i++ {
		base := RandomHash().String()

		hash, err := HexToHash(base)
		if err != nil {
			t.Error(err)
		}

		h2 := new(Hash)
		err = h2.UnmarshalText([]byte(base))
		if err != nil {
			t.Error(err)
		}
		if hash.IsSameAs(h2) == false {
			t.Errorf("Hash from UnmarshalText is incorrect - %v vs %v", hash, h2)
		}

		h3 := new(Hash)
		err = json.Unmarshal([]byte("\""+base+"\""), h3)
		if err != nil {
			t.Error(err)
		}
		if hash.IsSameAs(h3) == false {
			t.Errorf("Hash from json.Unmarshal is incorrect - %v vs %v", hash, h3)
		}
	}
}

func TestDoubleSha(t *testing.T) {
	testVector := map[string]string{
		"abc": "4f8b42c22dd3729b519ba6f68d2da7cc5b2d606d05daed5ad5128cc03e6c6358",
		"":    "5df6e0e2761359d30a8275058e299fcc0381534545f55cf43e41983f5d4c9456",
		"abcdbcdecdefdefgefghfghighijhijkijkljklmklmnlmnomnopnopq":                                                         "0cffe17f68954dac3a84fb1458bd5ec99209449749b2b308b7cb55812f9563af",
		"abcdefghbcdefghicdefghijdefghijkefghijklfghijklmghijklmnhijklmnoijklmnopjklmnopqklmnopqrlmnopqrsmnopqrstnopqrstu": "accd7bd1cb0fcbd85cf0ba5ba96945127776373a7d47891eb43ed6b1e2ee60fe",
	}

	for k, v := range testVector {
		b := DoubleSha([]byte(k))
		h, err := NewShaHash(b)
		if err != nil {
			t.Error(err)
		}
		if h.String() != v {
			t.Errorf("DoubleSha failed %v != %v", h.String(), v)
		}
	}
}

func TestNewShaHashFromStruct(t *testing.T) {
	testVector := map[string]string{
		"abc": "c127d30fe315d2d3f2dfeae6b9d57c6aa6322c73fb3fd868963660d6cdcd471f",
		"":    "e2854aa639f07056d58cc02ab52d169c48af8b418fcb0df7842f22a1b2ab3ac2",
		"abcdbcdecdefdefgefghfghighijhijkijkljklmklmnlmnomnopnopq":                                                         "c226baeb2cad51713659f5e111aaaa6a5a4cfffe7d874c3974c212f4c77fe9d7",
		"abcdefghbcdefghicdefghijdefghijkefghijklfghijklmghijklmnhijklmnoijklmnopjklmnopqklmnopqrlmnopqrsmnopqrstnopqrstu": "cdc9eb98889856282bf26c78ffde24c46cbeed70442acf25577fd1aef48a5951",
	}

	for k, v := range testVector {
		h, err := NewShaHashFromStruct(k)
		if err != nil {
			t.Error(err)
		}
		if h.String() != v {
			t.Errorf("NewShaHashFromStruct failed %v != %v", h.String(), v)
		}
	}
}

func TestCreateHash(t *testing.T) {
	strs := []string{
		"0000000000000000000000000000000000000000000000000000000000000000",
		"0000000000000000000000000000000000000000000000000000000000000001",
		"0000000000000000000000000000000000000000000000000000000000000002",
		"0000000000000000000000000000000000000000000000000000000000000003",
		"0000000000000000000000000000000000000000000000000000000000000004",
		"0000000000000000000000000000000000000000000000000000000000000005",
		"0000000000000000000000000000000000000000000000000000000000000006",
		"0000000000000000000000000000000000000000000000000000000000000007",
		"0000000000000000000000000000000000000000000000000000000000000008",
		"0000000000000000000000000000000000000000000000000000000000000009",
		"000000000000000000000000000000000000000000000000000000000000000a",
		"000000000000000000000000000000000000000000000000000000000000000b",
		"000000000000000000000000000000000000000000000000000000000000000c",
		"000000000000000000000000000000000000000000000000000000000000000d",
		"000000000000000000000000000000000000000000000000000000000000000e",
		"000000000000000000000000000000000000000000000000000000000000000f"}
	var hashes []interfaces.BinaryMarshallable
	for _, str := range strs {
		h, err := NewShaHashFromStr(str)
		if err != nil {
			t.Error(err)
		}
		hashes = append(hashes, h)
	}
	h, err := CreateHash(hashes...)
	if err != nil {
		t.Error(err)
	}
	if h.String() != "f3635ea6ad7cd94849624a1d7d739a14611d76fe8e1607d9eba1f9a258442e63" {
		t.Errorf("Invalid hash - %v vs f3635ea6ad7cd94849624a1d7d739a14611d76fe8e1607d9eba1f9a258442e63", h.String())
	}
}
