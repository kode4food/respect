package resp

import (
	"fmt"
	"io"
)

type (
	VerbatimString struct {
		data []byte
		enc  encoding
	}

	encoding [encodingLength]byte
)

const encodingLength = 3

var (
	EmptyVerbatimString = &VerbatimString{
		data: []byte{},
		enc:  emptyEncoding,
	}

	colon         = []byte{':'}
	emptyEncoding = encoding{}
)

// compile-time check for interface implementation
var _ String = (*VerbatimString)(nil)

func MakeVerbatimString(enc string, data string) (*VerbatimString, error) {
	e, err := makeEncoding(enc)
	if err != nil {
		return nil, err
	}
	return &VerbatimString{
		data: ([]byte)(data),
		enc:  e,
	}, nil
}

func readVerbatimString(r *Reader) (*VerbatimString, error) {
	data, err := r.readBulk()
	if err != nil {
		return EmptyVerbatimString, err
	}
	if len(data) < encodingLength+1 {
		return EmptyVerbatimString, fmt.Errorf(ErrInvalidLength, len(data))
	}
	enc := encoding{}
	copy(enc[:], data[:encodingLength])
	return &VerbatimString{
		data: data[encodingLength+1:],
		enc:  enc,
	}, nil
}

func (*VerbatimString) Tag() Tag {
	return VerbatimStringTag
}

func (s *VerbatimString) Marshal(w io.Writer) error {
	if _, err := w.Write([]byte{byte(s.Tag())}); err != nil {
		return err
	}
	if err := writeInt(len(s.enc)+len(s.data)+1, w); err != nil {
		return err
	}
	if _, err := w.Write(s.enc[:]); err != nil {
		return err
	}
	if _, err := w.Write(colon); err != nil {
		return err
	}
	if _, err := w.Write(s.data); err != nil {
		return err
	}
	return writeNewline(w)
}

func (s *VerbatimString) Equal(v Value) bool {
	if v, ok := v.(*VerbatimString); ok {
		return s == v ||
			s.Encoding() == v.Encoding() && s.String() == v.String()
	}
	return false
}

func (s *VerbatimString) Encoding() string {
	return string(s.enc[:])
}

func (s *VerbatimString) String() string {
	return string(s.data)
}

func makeEncoding(enc string) (encoding, error) {
	if len(enc) != encodingLength {
		return emptyEncoding, fmt.Errorf(ErrInvalidEncoding, enc)
	}
	res := encoding{}
	copy(res[:], enc)
	return res, nil
}
