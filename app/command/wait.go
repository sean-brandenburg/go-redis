package command

import (
	"fmt"
	"strconv"
)

type Wait struct {
	NumReplicas int64
	WaitForMs   int64
}

func (Wait) String() string {
	return "Wait"
}

func (w Wait) EncodedCommand() (string, error) {
	e := Encoder{UseBulkStrings: true}

	return e.EncodeArray(
		[]any{
			string(WaitCmd),
			strconv.FormatInt(w.NumReplicas, 10),
			strconv.FormatInt(w.WaitForMs, 10),
		},
	)
}

func (Wait) CommandType() CommandType {
	return WaitCmd
}

func toWait(data []any) (Wait, error) {
	if len(data) != 2 {
		return Wait{}, fmt.Errorf("expected wait to have parameters <NumReplicas> and <WaitForMs> but got: %v", data)
	}

	rawNumReplias := data[0]
	rawNumReplicasString, ok := rawNumReplias.(string)
	if !ok {
		return Wait{}, fmt.Errorf("expected <NumReplicas> Wait parameter to be a string representing an integer, but got %v", rawNumReplicasString)
	}

	numReplicas, err := strconv.ParseInt(rawNumReplicasString, 10, 64)
	if err != nil {
		return Wait{}, fmt.Errorf("expected <NumReplicas> Wait parameter to be an integer, but got %v", rawNumReplicasString)
	}

	rawWaitForMs := data[1]
	rawWaitForMsString, ok := rawWaitForMs.(string)
	if !ok {
		return Wait{}, fmt.Errorf("expected <WaitForMs> Wait parameter to be a string representing an integer, but got %v", rawWaitForMs)
	}

	waitForMs, err := strconv.ParseInt(rawWaitForMsString, 10, 64)
	if err != nil {
		return Wait{}, fmt.Errorf("expected <WaitForMs> Wait parameter to be an integer, but got %v", rawWaitForMsString)
	}

	return Wait{
		NumReplicas: numReplicas,
		WaitForMs:   waitForMs,
	}, nil
}
