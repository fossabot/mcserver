package packets

import (
	"errors"
	"fmt"
	"github.com/imdario/mergo"
	"github.com/r4g3baby/mcserver/pkg/protocol"
	"reflect"
)

var (
	packets = map[protocol.Protocol]map[protocol.State]map[protocol.Direction]map[reflect.Type]int32{
		protocol.Unknown: {
			protocol.Handshaking: {
				protocol.ServerBound: {
					reflect.TypeOf((*PacketHandshakingStart)(nil)).Elem(): 0x00,
				},
			},
			protocol.Status: {
				protocol.ClientBound: {
					reflect.TypeOf((*PacketStatusOutResponse)(nil)).Elem(): 0x00,
					reflect.TypeOf((*PacketStatusOutPong)(nil)).Elem():     0x01,
				},
				protocol.ServerBound: {
					reflect.TypeOf((*PacketStatusInRequest)(nil)).Elem(): 0x00,
					reflect.TypeOf((*PacketStatusInPing)(nil)).Elem():    0x01,
				},
			},
			protocol.Login: {
				protocol.ClientBound: {
					reflect.TypeOf((*PacketLoginOutDisconnect)(nil)).Elem(): 0x00,
					reflect.TypeOf((*PacketLoginOutSuccess)(nil)).Elem():    0x02,
				},
				protocol.ServerBound: {
					reflect.TypeOf((*PacketLoginInStart)(nil)).Elem(): 0x00,
				},
			},
		},
	}

	packetsByID = map[protocol.Protocol]map[protocol.State]map[protocol.Direction]map[int32]reflect.Type{
		protocol.Unknown: {
			protocol.Handshaking: {
				protocol.ServerBound: {
					0x00: reflect.TypeOf((*PacketHandshakingStart)(nil)).Elem(),
				},
			},
			protocol.Status: {
				protocol.ClientBound: {
					0x00: reflect.TypeOf((*PacketStatusOutResponse)(nil)).Elem(),
					0x01: reflect.TypeOf((*PacketStatusOutPong)(nil)).Elem(),
				},
				protocol.ServerBound: {
					0x00: reflect.TypeOf((*PacketStatusInRequest)(nil)).Elem(),
					0x01: reflect.TypeOf((*PacketStatusInPing)(nil)).Elem(),
				},
			},
			protocol.Login: {
				protocol.ClientBound: {
					0x00: reflect.TypeOf((*PacketLoginOutDisconnect)(nil)).Elem(),
					0x02: reflect.TypeOf((*PacketLoginOutSuccess)(nil)).Elem(),
				},
				protocol.ServerBound: {
					0x00: reflect.TypeOf((*PacketLoginInStart)(nil)).Elem(),
				},
			},
		},
	}
)

func init() {
	if err := RegisterPackets(protocol.V1_16, map[protocol.State]map[protocol.Direction]map[reflect.Type]int32{
		protocol.Play: {
			protocol.ClientBound: {
				reflect.TypeOf((*PacketPlayOutServerDifficulty)(nil)).Elem(): 0x0D,
				reflect.TypeOf((*PacketPlayOutDisconnect)(nil)).Elem():       0x19,
				reflect.TypeOf((*PacketPlayOutKeepAlive)(nil)).Elem():        0x1F,
				reflect.TypeOf((*PacketPlayOutJoinGame)(nil)).Elem():         0x24,
				reflect.TypeOf((*PacketPlayOutPositionAndLook)(nil)).Elem():  0x34,
			},
			protocol.ServerBound: {
				reflect.TypeOf((*PacketPlayInKeepAlive)(nil)).Elem(): 0x10,
			},
		},
	}); err != nil {
		panic(err)
	}

	copyPackets(protocol.V1_16, protocol.V1_16_1, protocol.V1_16_2, protocol.V1_16_3, protocol.V1_16_4)
}

func GetPacketID(proto protocol.Protocol, state protocol.State, direction protocol.Direction, packet protocol.Packet) (int32, error) {
	if states, ok := packets[proto]; ok {
		if directions, ok := states[state]; ok {
			if pTypes, ok := directions[direction]; ok {
				if id, ok := pTypes[reflect.TypeOf(packet).Elem()]; ok {
					return id, nil
				}
			}
		}
	}
	return 0, errors.New("no packet id found for the given options")
}

func GetPacketByID(proto protocol.Protocol, state protocol.State, direction protocol.Direction, id int32) (protocol.Packet, error) {
	if states, ok := packetsByID[proto]; ok {
		if directions, ok := states[state]; ok {
			if pIDs, ok := directions[direction]; ok {
				if pType, ok := pIDs[id]; ok {
					return reflect.New(pType).Interface().(protocol.Packet), nil
				}
			}
		}
	}
	return nil, fmt.Errorf("no packet found with id %d", id)
}

func RegisterPackets(proto protocol.Protocol, packetsMap map[protocol.State]map[protocol.Direction]map[reflect.Type]int32) error {
	newPacketsMap := make(map[protocol.State]map[protocol.Direction]map[reflect.Type]int32)
	for state, directions := range packets[protocol.Unknown] {
		newPacketsMap[state] = make(map[protocol.Direction]map[reflect.Type]int32)
		for direction, pTypes := range directions {
			newPacketsMap[state][direction] = make(map[reflect.Type]int32)
			for pType, pID := range pTypes {
				newPacketsMap[state][direction][pType] = pID
			}
		}
	}

	if err := mergo.Merge(&newPacketsMap, packetsMap, mergo.WithOverride); err != nil {
		return err
	}

	packets[proto] = newPacketsMap

	packetsByID[proto] = make(map[protocol.State]map[protocol.Direction]map[int32]reflect.Type)
	for state, directions := range newPacketsMap {
		packetsByID[proto][state] = make(map[protocol.Direction]map[int32]reflect.Type)
		for direction, pTypes := range directions {
			packetsByID[proto][state][direction] = make(map[int32]reflect.Type)
			for pType, pID := range pTypes {
				packetsByID[proto][state][direction][pID] = pType
			}
		}
	}

	return nil
}

func copyPackets(src protocol.Protocol, destinations ...protocol.Protocol) {
	for _, dst := range destinations {
		packets[dst] = packets[src]
		packetsByID[dst] = packetsByID[src]
	}
}
