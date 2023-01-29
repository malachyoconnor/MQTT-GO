package network

import "errors"

func DecodeVarInt(data []byte) uint64 {
	v := uint64(data[0])
	// Length is encoded as 2^(first 2 bits)
	length := 1 << (v >> 6)
	v = v & 63
	if len(data) == 1 {
		return v
	}

	for i := 1; i < length; i++ {
		v = (v << 8) + uint64(data[i]) // this overflows...
	}
	return v
}

func EncodeVarInt(toEncode uint64) ([]byte, error) {
	if toEncode&(1<<63) == 1 || toEncode&(1<<62) == 1 {
		return nil, errors.New("error: Higher bits set, cannot encode")
	}
	byteMask := uint64(255)

	masks := []uint64{
		byteMask<<(64-8) + byteMask<<(64-2*8) + byteMask<<(64-3*8) + byteMask<<(64-4*8),
		byteMask<<(64-5*8) + byteMask<<(64-6*8),
		byteMask << (64 - 7*8)}

	bytesRequired := byte(1)

	for i, mask := range masks {
		if toEncode&mask > 0 {
			bytesRequired = 1 << (3 - i)
		}
	}

	result := make([]byte, bytesRequired)
	result[0] = (bytesRequired << 6) + byte(toEncode>>(64-8))

	for i := 1; i < int(bytesRequired); i++ {
		result[i] = byte(toEncode >> (64 - (i+1)*8))
	}

	return result, nil
}

// 1111 0000  Check first 4 bytes
// 1111 1100  Check next 2 bytes
// 1111 1110  Check next 1
// Otherwise just use one bit
