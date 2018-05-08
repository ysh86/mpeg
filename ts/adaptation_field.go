package ts

import "github.com/32bitkid/bitreader"
import "io"

// AdaptationFieldControl is the two bit code that appears in a transport
// stream packet header that determines whether an Adapation Field appears
// in the bit stream.
type AdaptationFieldControl uint32

const (
	_                AdaptationFieldControl = iota
	PayloadOnly                             // 0b01
	FieldOnly                               // 0b10
	FieldThenPayload                        //0b11
)

// AdaptationField is an optional field in a transport stream packet header.
// TODO(jh): Needs implementation
type AdaptationField struct {
	Length uint32

	DiscontinuityIndicator            bool
	RandomAccessIndicator             bool
	ElementaryStreamPriorityIndicator bool
	PCRFlag                           bool
	OPCRFlag                          bool
	SplicingPointFlag                 bool
	TransportPrivateDataFlag          bool
	AdaptationFieldExtensionFlag      bool

	PCR uint64

	Junk []byte
}

func newAdaptationField(br bitreader.BitReader) (*AdaptationField, uint32, error) {
	adaptationField := AdaptationField{}
	length, err := br.Read32(8)
	if err != nil {
		return nil, 0, err
	}

	adaptationField.Length = length
	adaptationField.Junk = make([]byte, length)
	_, err = io.ReadFull(br, adaptationField.Junk)
	if err == io.EOF {
		return nil, 0, io.ErrUnexpectedEOF
	} else if err != nil {
		return nil, 0, err
	}

	// flags
	if length > 0 {
		var flags = adaptationField.Junk[0]

		adaptationField.AdaptationFieldExtensionFlag = (flags&1 == 1)
		flags >>= 1
		adaptationField.TransportPrivateDataFlag = (flags&1 == 1)
		flags >>= 1
		adaptationField.SplicingPointFlag = (flags&1 == 1)
		flags >>= 1
		adaptationField.OPCRFlag = (flags&1 == 1)
		flags >>= 1
		adaptationField.PCRFlag = (flags&1 == 1)
		flags >>= 1
		adaptationField.ElementaryStreamPriorityIndicator = (flags&1 == 1)
		flags >>= 1
		adaptationField.RandomAccessIndicator = (flags&1 == 1)
		flags >>= 1
		adaptationField.DiscontinuityIndicator = (flags&1 == 1)
		flags >>= 1
	}
	// PCR: 48bits
	if adaptationField.PCRFlag {
		if length < 1+6 {
			// TODO: error!
		}
		var temp uint64
		temp = uint64(adaptationField.Junk[1+0])
		temp <<= 8
		temp |= uint64(adaptationField.Junk[1+1])
		temp <<= 8
		temp |= uint64(adaptationField.Junk[1+2])
		temp <<= 8
		temp |= uint64(adaptationField.Junk[1+3])
		temp <<= 8
		temp |= uint64(adaptationField.Junk[1+4])
		temp <<= 8
		temp |= uint64(adaptationField.Junk[1+5])

		adaptationField.PCR = (temp>>15)*300 + (temp & 0x7fff)
	}

	return &adaptationField, length, nil
}
