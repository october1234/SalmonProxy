package mc

import (
	"errors"
	"io"
)

const (
	VARINT_SEGMENT_BITS = 0x7F
	VARINT_CONTINUE_BIT = 0x80
)

type FirstPacket struct {
	ProtocolVersion int
	ServerAddress   string
	ServerPort      uint16
}

func ReadFirstPacket(inputStream io.Reader, fp *FirstPacket) ([]byte, error) {
	buffer := make([]byte, (5+5)+(5+768+2+5))
	_, err := inputStream.Read(buffer)
	if err != nil {
		return buffer, err
	}

	i := 0
	_, err = readVarInt(&i, buffer)
	if err != nil {
		return buffer, err
	}

	packetId, err := readVarInt(&i, buffer)
	if err != nil {
		return buffer, err
	}
	if packetId != 0x00 {
		return buffer, errors.New("first packet id should be 0x00")
	}

	ver, err := readVarInt(&i, buffer)
	if err != nil {
		return buffer, err
	}
	fp.ProtocolVersion = ver

	addLen, err := readVarInt(&i, buffer)
	if err != nil {
		return buffer, err
	}

	address := buffer[i : i+addLen]
	fp.ServerAddress = string(address)

	return buffer, nil
}

func readVarInt(index *int, buffer []byte) (int, error) {
	var value int
	var position int

	for {
		value |= (int(buffer[*index]) & VARINT_SEGMENT_BITS) << position
		if (int(buffer[*index]) & VARINT_CONTINUE_BIT) == 0 {
			break
		}
		position += 7
		if position >= 32 {
			return 0, errors.New("VarInt is too big")
		}
		*index++
	}

	*index++
	return value, nil
}
