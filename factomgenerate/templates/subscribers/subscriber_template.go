//Ͼ/*
// The FactomGenerate templates use Greek Capitol syllabary characters using "Ͼ" U+03FE, "Ͽ" U+03FF as the
// delimiters. This is done so the template can be valid go code and goimports and gofmt will work correctly on the
// code and it can be tested in unmodified form. For more information see factomgenerate/generate.go
//*/Ͽ

package subscribers // this is only here to make gofmt happy and is never in the generated code

//Ͼdefine "subscribe-imports"Ͽ

import (
	"github.com/FactomProject/factomd/pubsub/subscribers"
)

//ϾendϿ

type Ͼ_valuetypeϿ int // not used when generating, only used for testing

// Expects: valuetype <type>

//Ͼdefine "subscribebyvalue"Ͽ
// Start subscribeByValue generated go code

// Ͼ_typenameϿ subscriber has the basic necessary function implementations.
type Subscribe_ByValue_Ͼ_valuetypeϿ struct {
	subscribers.Value
}

func (s *Subscribe_ByValue_Ͼ_valuetypeϿ) Read() Ͼ_valuetypeϿ {
	o := s.Value.Read()     // call the generic implementation
	return o.(Ͼ_valuetypeϿ) // cast the return to the specific type
}

// End Subscribebyvalue generated code
//Ͼend Ͽ