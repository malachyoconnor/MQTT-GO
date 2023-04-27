package packets_test

import (
	"MQTT-GO/packets"
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestEncodingVarInts(test *testing.T) {
	for i := 0; i < 200; i++ {
		arr := packets.EncodeVarLengthInt(i)
		x, _, _ := packets.DecodeVarLengthInt(arr)

		if x != i {
			test.Error("Encoding and decoding are not symmetrical")
		}
	}
}

func TestEncodingFixedHeader(t *testing.T) {
	for _, header := range []packets.ControlHeader{
		{Type: 4, RemainingLength: 40, Flags: 10},
		{Type: 1, RemainingLength: 25, Flags: 11},
		{Type: 12, RemainingLength: 400, Flags: 2},
		{Type: 2, RemainingLength: 5, Flags: 0},
	} {

		arr := packets.EncodeFixedHeader(header)
		result, _, _ := packets.DecodeFixedHeader(arr)

		if !cmp.Equal(*result, header) {
			t.Error("Encode and Decode Fixed Header are not symmetrical")
		}

	}
}

func TestEncodingAndDecodingConnect(t *testing.T) {
	packet := packets.Packet{}
	packet.ControlHeader = &packets.ControlHeader{Type: packets.CONNECT, RemainingLength: 25, Flags: 11}
	packet.VariableLengthHeader = &packets.ConnectVariableHeader{ProtocolName: "MQTT", ProtocolLevel: 4, ConnectFlags: 0, KeepAlive: 60}
	packet.Payload = &packets.PacketPayload{ClientID: "test"}

	encodedPacket, err := packets.EncodeConnect(&packet)
	if err != nil {
		t.Error(err)
	}

	decodedPacket, err := packets.DecodeConnect(encodedPacket)
	if err != nil {
		t.Error(err)
	}

	if !reflect.DeepEqual(*packet.ControlHeader, *decodedPacket.ControlHeader) {
		t.Error("Control headers are not symmetrical")
	}
	if !reflect.DeepEqual(packet.VariableLengthHeader, decodedPacket.VariableLengthHeader) {
		t.Error("Variable length headers are not symmetrical")
	}
	if !reflect.DeepEqual(*packet.Payload, *decodedPacket.Payload) {
		t.Error("Payloads are not symmetrical")
	}

}

func TestEncodingAndDecodingSubscribe(t *testing.T) {
	packet := packets.Packet{}
	packet.ControlHeader = &packets.ControlHeader{Type: packets.SUBSCRIBE, RemainingLength: 25, Flags: 2}
	packet.VariableLengthHeader = &packets.SubscribeVariableHeader{PacketIdentifier: 1}
	packet.Payload = &packets.PacketPayload{RawApplicationMessage: []byte{1, 2, 3, 4, 5}}

	encodedPacket, err := packets.EncodeSubscribe(&packet)
	if err != nil {
		t.Error(err)
	}

	decodedPacket, err := packets.DecodeSubscribe(encodedPacket)
	if err != nil {
		t.Error(err)
	}

	if !reflect.DeepEqual(*packet.ControlHeader, *decodedPacket.ControlHeader) {
		t.Error("Control headers are not symmetrical")
	}
	if !reflect.DeepEqual(packet.VariableLengthHeader, decodedPacket.VariableLengthHeader) {
		t.Error("Variable length headers are not symmetrical")
	}
	if !reflect.DeepEqual(*packet.Payload, *decodedPacket.Payload) {
		t.Error("Payloads are not symmetrical")
	}

}

func TestEncodingAndDecodingPublish(t *testing.T) {
	packet := packets.Packet{}
	packet.ControlHeader = &packets.ControlHeader{Type: packets.PUBLISH, RemainingLength: 25, Flags: 0}
	packet.VariableLengthHeader = &packets.PublishVariableHeader{PacketIdentifier: 0, TopicFilter: "test"}
	packet.Payload = &packets.PacketPayload{RawApplicationMessage: []byte{1, 2, 3, 4, 5}}

	encodedPacket, err := packets.EncodePublish(&packet)
	if err != nil {
		t.Error(err)
	}

	decodedPacket, err := packets.DecodePublish(encodedPacket)
	if err != nil {
		t.Error(err)
	}

	if !reflect.DeepEqual(*packet.ControlHeader, *decodedPacket.ControlHeader) {
		t.Error("Control headers are not symmetrical")
	}
	if !reflect.DeepEqual(packet.VariableLengthHeader, decodedPacket.VariableLengthHeader) {
		t.Error("Variable length headers are not symmetrical")
	}
	if !reflect.DeepEqual(*packet.Payload, *decodedPacket.Payload) {
		t.Error("Payloads are not symmetrical")
	}
}

func TestEncodingAndDecodingDisconnect(t *testing.T) {
	packet := packets.Packet{}
	packet.ControlHeader = &packets.ControlHeader{Type: packets.DISCONNECT, RemainingLength: 0, Flags: 0}

	encodedPacket := packets.EncodeFixedHeader(*packet.ControlHeader)

	decodedPacket, err := packets.DecodeDisconnect(encodedPacket)
	if err != nil {
		t.Error(err)
	}

	if !reflect.DeepEqual(*packet.ControlHeader, *decodedPacket.ControlHeader) {
		t.Error("Control headers are not symmetrical")
	}

}

func TestEncodingAndDecodingConnack(t *testing.T) {
	packet := packets.Packet{}
	packet.ControlHeader = &packets.ControlHeader{Type: packets.CONNACK, RemainingLength: 2, Flags: 0}
	packet.VariableLengthHeader = &packets.ConnackVariableHeader{ConnectAcknowledgementFlags: 0, ConnectReturnCode: 0}

	encodedPacket := packets.CreateConnACK(false, 0)
	decodedPacket, err := packets.DecodeCONNACK(encodedPacket)
	if err != nil {
		t.Error(err)
	}

	if !reflect.DeepEqual(*packet.ControlHeader, *decodedPacket.ControlHeader) {
		t.Error("Control headers are not symmetrical")
	}

	if !reflect.DeepEqual(packet.VariableLengthHeader, decodedPacket.VariableLengthHeader) {
		t.Error("Control headers are not symmetrical")
	}
}

func TestEncodingAndDecodingPingreq(t *testing.T) {
	packet := packets.Packet{}
	packet.ControlHeader = &packets.ControlHeader{Type: packets.PINGREQ, RemainingLength: 0, Flags: 0}

	encodedPacket := packets.EncodeFixedHeader(*packet.ControlHeader)

	decodedPacket, err := packets.DecodePingreq(encodedPacket)
	if err != nil {
		t.Error(err)
	}

	if !reflect.DeepEqual(*packet.ControlHeader, *decodedPacket.ControlHeader) {
		t.Error("Control headers are not symmetrical")
	}
}

func TestEncodingAndDecodingSuback(t *testing.T) {
	packet := packets.Packet{}
	packet.ControlHeader = &packets.ControlHeader{Type: packets.SUBACK, RemainingLength: 25, Flags: 2}
	packet.VariableLengthHeader = &packets.SubackVariableHeader{PacketIdentifier: 1}
	packet.Payload = &packets.PacketPayload{RawApplicationMessage: []byte{1, 2, 3, 4, 5}}

	encodedPacket, err := packets.EncodeSuback(&packet)
	if err != nil {
		t.Error(err)
	}

	decodedPacket, err := packets.DecodeSuback(encodedPacket)
	if err != nil {
		t.Error(err)
	}

	if !reflect.DeepEqual(*packet.ControlHeader, *decodedPacket.ControlHeader) {
		t.Error("Control headers are not symmetrical")
	}
	if !reflect.DeepEqual(packet.VariableLengthHeader, decodedPacket.VariableLengthHeader) {
		t.Error("Variable length headers are not symmetrical")
	}
	if !reflect.DeepEqual(*packet.Payload, *decodedPacket.Payload) {
		t.Error("Payloads are not symmetrical")
	}

}

func TestEncodingAndDecodingUnsubscribe(t *testing.T) {
	packet := packets.Packet{}
	packet.ControlHeader = &packets.ControlHeader{Type: packets.UNSUBSCRIBE, RemainingLength: 25, Flags: 2}
	packet.VariableLengthHeader = &packets.UnsubscribeVariableHeader{PacketIdentifier: 1}
	packet.Payload = &packets.PacketPayload{TopicList: []packets.TopicWithQoS{{Topic: "test1"}, {Topic: "test2"}}}

	encodedPacket, err := packets.EncodeUnsubscribe(&packet)
	if err != nil {
		t.Error(err)
	}

	decodedPacket, err := packets.DecodeUnsubscribe(encodedPacket)
	if err != nil {
		t.Error(err)
	}

	if !reflect.DeepEqual(*packet.ControlHeader, *decodedPacket.ControlHeader) {
		t.Error("Control headers are not symmetrical")
	}
	if !reflect.DeepEqual(packet.VariableLengthHeader, decodedPacket.VariableLengthHeader) {
		t.Error("Variable length headers are not symmetrical")
	}
	if !reflect.DeepEqual(*packet.Payload, *decodedPacket.Payload) {
		t.Error("Payloads are not symmetrical")
	}

}
