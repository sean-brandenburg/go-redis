package command

import (
	"fmt"
)

const HARDCODEC_REPL_ID = "8371b4fb1155b71f4a04d3e1bc3e18c4a990aeeb"

type PSync struct {
	ReplicationID string
	MasterOffset  string
}

func (psync PSync) String() string {
	return fmt.Sprintf("PSYNC: (ReplicationID=%q) (MasterOffset=%q)", psync.ReplicationID, psync.MasterOffset)
}

func (psync PSync) EncodedCommand() (string, error) {
	e := Encoder{UseBulkStrings: true}
	return e.Encode([]any{string(pSyncCmd), psync.ReplicationID, psync.MasterOffset})
}

func toPSync(data []any) (PSync, error) {
	if len(data) != 2 {
		return PSync{}, fmt.Errorf("expected exactly 2 data elements for PSYNC command, but found %d: %v", len(data), data)
	}

	replicationID, ok := data[0].(string)
	if !ok {
		return PSync{}, fmt.Errorf("expected the 1st1st1st1st1st1st1st1st1st1st1st parameter to the PSYNC command to be a string but it was %[1]v of type %[1]v", data[0])
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
