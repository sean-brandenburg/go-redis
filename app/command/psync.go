package command

import (
	"encoding/hex"
	"fmt"
)

const (
	// TODO: Remove these
	HARDCODE_REPL_ID   = "8371b4fb1155b71f4a04d3e1bc3e18c4a990aeeb"
	HARDCODE_EMPTY_RDB = "524544495330303131fa0972656469732d76657205372e322e30fa0a72656469732d62697473c040fa056374696d65c26d08bc65fa08757365642d6d656dc2b0c41000fa08616f662d62617365c000fff06e3bfec0ff5aa2"
)

func GetHardedCodedEmptyRDBBytes() []byte {
	emptyRDB, _ := hex.DecodeString(HARDCODE_EMPTY_RDB)
	return emptyRDB
}

type PSync struct {
	ReplicationID string
	MasterOffset  string
}

func (psync PSync) String() string {
	return fmt.Sprintf("PSYNC: (ReplicationID=%q) (MasterOffset=%q)", psync.ReplicationID, psync.MasterOffset)
}

func (psync PSync) EncodedCommand() (string, error) {
	e := Encoder{UseBulkStrings: true}
	return e.EncodeArray([]any{string(PSyncCmd), psync.ReplicationID, psync.MasterOffset})
}

func (PSync) CommandType() CommandType {
	return PSyncCmd
}

func toPSync(data []any) (PSync, error) {
	if len(data) != 2 {
		return PSync{}, fmt.Errorf("expected exactly 2 data elements for PSYNC command, but found %d: %v", len(data), data)
	}

	replicationID, ok := data[0].(string)
	if !ok {
		return PSync{}, fmt.Errorf("expected the 1st parameter to the PSYNC command to be a string but it was %[1]v of type %[1]v", data[0])
	}

	masterOffset, ok := data[1].(string)
	if !ok {
		return PSync{}, fmt.Errorf("expected the 2nd parameter to the PSYNC command to be a string but it was %[1]v of type %[1]v", data[1])
	}

	return PSync{
		ReplicationID: replicationID,
		MasterOffset:  masterOffset,
	}, nil
}
