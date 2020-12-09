package packets

import (
	"github.com/r4g3baby/mcserver/pkg/protocol"
	"github.com/r4g3baby/mcserver/pkg/util/bytes"
)

type PacketPlayOutServerDifficulty struct {
	Difficulty uint8
	Locked     bool
}

func (packet *PacketPlayOutServerDifficulty) GetID(proto protocol.Protocol) (int32, error) {
	return GetPacketID(proto, protocol.Play, protocol.ClientBound, packet)
}

func (packet *PacketPlayOutServerDifficulty) Read(_ protocol.Protocol, buffer *bytes.Buffer) error {
	difficulty, err := buffer.ReadUint8()
	if err != nil {
		return err
	}
	packet.Difficulty = difficulty

	locked, err := buffer.ReadBool()
	if err != nil {
		return err
	}
	packet.Locked = locked

	return nil
}

func (packet *PacketPlayOutServerDifficulty) Write(_ protocol.Protocol, buffer *bytes.Buffer) error {
	if err := buffer.WriteUint8(packet.Difficulty); err != nil {
		return err
	}

	if err := buffer.WriteBool(packet.Locked); err != nil {
		return err
	}

	return nil
}
