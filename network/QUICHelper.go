package network

import (
	"crypto/sha256"
	"errors"
	"fmt"

	"golang.org/x/crypto/hkdf"
)

func DecodeVarInt(data []byte) (uint64, uint64, error) {
	v := uint64(data[0])
	// Length is encoded as 2^(first 2 bits)
	length := uint64(1 << (v >> 6))

	if int(length) > len(data) {
		return 0, 0, errors.New("error: Data missing")
	}

	v = v & 63
	if len(data) == 1 {
		return v, length, nil
	}

	for i := uint64(1); i < length; i++ {
		v = (v << 8) + uint64(data[i])
	}
	return v, length, nil
}

func EncodeVarInt(toEncode uint64) ([]byte, error) {
	if toEncode&(1<<63) == 1 || toEncode&(1<<62) == 1 {
		return nil, errors.New("error: Higher bits set, cannot encode")
	}
	byteMask := uint64(255)

	masks := []uint64{
		byteMask<<(64-8) + byteMask<<(64-2*8) + byteMask<<(64-3*8) + byteMask<<(64-4*8) + 1<<31 + 1<<30,
		byteMask<<(64-5*8) + byteMask<<(64-6*8) + 1<<15 + 1<<14,
		byteMask<<(64-7*8) + 1<<7 + 1<<6}

	bytesRequired := byte(1)

	for i, mask := range masks {
		// fmt.Printf("%064b <- Mask\n", mask)
		// fmt.Printf("%064b <- To encode\n", toEncode)
		if toEncode&mask > 0 {
			bytesRequired = 1 << (3 - i)
			break
		}
	}

	result := make([]byte, bytesRequired)

	switch bytesRequired {
	case 8:
		result[0] = 3 << 6
	case 4:
		result[0] = 2 << 6
	case 2:
		result[0] = 1 << 6
	}
	result[0] += byte(toEncode >> (uint64(bytesRequired-1) * 8))

	for i := int(bytesRequired) - 2; i >= 0; i-- {
		result[bytesRequired-byte(i)-1] = byte(toEncode >> (i * 8))
	}

	return result, nil
}

func getBits(value byte, bitsToExtract ...byte) byte {
	resultByte := byte(0)
	for i, bitIndex := range bitsToExtract {
		bit := (value >> bitIndex) & 1
		resultByte += (bit << i)
	}

	return resultByte
}

func BytesToUint32(bitsToExtract ...byte) (result uint32) {
	if len(bitsToExtract) > 4 {
		panic("Error: Given more than 4 bytes to convert to 32 bit unsigned integer")
	}

	for i := 0; i < len(bitsToExtract); i++ {
		result += uint32(bitsToExtract[i]) << (i * 8)
	}
	return result
}

func decodePacketNumber(client_dst_connection_id []byte) uint32 {

	initial_salt := []byte{56, 118, 44, 247, 245, 89, 52, 179, 77, 23, 154, 230, 164, 200, 12, 173, 204, 187, 127, 10}
	hash := sha256.New
	inital_secret := hkdf.Extract(hash, client_dst_connection_id, initial_salt)

	client_initial_secret := hkdf.Expand(hash, inital_secret, []byte("client in"))
	server_initial_secret := hkdf.Expand(hash, inital_secret, []byte("server in"))

	fmt.Println(client_initial_secret, server_initial_secret)
	return 0
}
