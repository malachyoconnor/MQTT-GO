package network

import (
	"errors"
	"fmt"
)

func decodeLongHeaderPacket(toDecode []byte) (LongHeaderPacket, error) {

	longHeader := LongHeader{}

	fmt.Printf("%0b\n", toDecode[0])

	longHeader.HeaderForm = getBits(toDecode[0], 7) == 1
	if !longHeader.HeaderForm {
		return nil, errors.New("error: headerForm value not equal to 1")
	}

	longHeader.FixedBit = getBits(toDecode[0], 6) == 1
	if !longHeader.HeaderForm {
		return nil, errors.New("error: Incorrect fixed bit")
	}

	longHeader.PacketType = getBits(toDecode[0], 5, 4)
	longHeader.TypeSpecificBits = getBits(toDecode[0], 3, 2, 1, 0)
	longHeader.Version = BytesToUint32(toDecode[4], toDecode[3], toDecode[2], toDecode[1])
	longHeader.DestinationConnectionIDLength = toDecode[5]
	offset := byte(6)
	longHeader.DestinationConnectionID = toDecode[offset : offset+longHeader.DestinationConnectionIDLength]
	offset += longHeader.DestinationConnectionIDLength
	longHeader.SourceConnectionIDLength = toDecode[offset]
	offset += 1
	longHeader.SourceConnectionID = toDecode[offset : offset+longHeader.SourceConnectionIDLength]

	// TODO: Implement the other packet types

	switch longHeader.PacketType {
	case LH_INITIAL:
		{
			initialPacket, err := decodeInitialPacket(&longHeader, toDecode[offset:])
			if err != nil {
				return nil, err
			}
			return initialPacket, nil
		}
	default:
		panic("error: Packet type not defined")
	}
}

func decodeInitialPacket(lh_packet *LongHeader, remainingHeader []byte) (*InitialPacket, error) {

	result := InitialPacket{}
	result.LongHeader = lh_packet

	tokenLength, offset, err := DecodeVarInt(remainingHeader)

	//TODO Remove this:
	if offset != 1 {
		panic("Token length not equal to 1!!")
	}

	if err != nil {
		return nil, err
	}

	if tokenLength != 0 {
		result.Token = remainingHeader[offset : offset+tokenLength]
		offset += tokenLength
	} else {
		result.Token = []byte{0}
	}

	length, packetLenOffset, err := DecodeVarInt(remainingHeader[offset:])
	result.Length = length
	if err != nil {
		return nil, err
	}
	offset += packetLenOffset

	packetNumberLength := 1 + (lh_packet.TypeSpecificBits & 192)
	fmt.Printf("Type specific bits: %08b \n", lh_packet.TypeSpecificBits)
	fmt.Println("Packet number length", packetNumberLength)

	fmt.Printf("Supposed packet number: %08b \n", remainingHeader[offset:offset+uint64(packetNumberLength)])

	encodedPacketNumber := BytesToUint32(remainingHeader[offset : offset+uint64(packetNumberLength)]...)
	result.PacketNumber = uint32(decodePacketNumber(0, int(encodedPacketNumber), int(packetNumberLength)))

	result.PacketPayload = remainingHeader[offset+uint64(packetNumberLength):]

	fmt.Printf("%08b \n", result.PacketPayload[0:2])

	return &result, nil

}
