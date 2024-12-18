package server

import (
	"errors"
	"fmt"
	"slices"
	"strings"

	"go.uber.org/zap"

	"github.com/codecrafters-io/redis-starter-go/app/command"
)

func RunCommand(server Server, conn Connection, cmd command.Command) error {
	if server.NodeType() == ReplicaNodeType && server.IsSteadyState() {
		// Once in steady state, replica nodes should only reply to replconf messages so set the conn to Noop
		if _, ok := cmd.(command.ReplConf); !ok {
			conn = LogNoopConn{
				Logger:   server.Logger(),
				ConnType: conn.ConnectionType(),
			}
		}
	}

	cmdExec := commandExecutor{
		server: server,
		conn:   conn,
	}
	return cmdExec.execute(cmd)
}

type commandExecutor struct {
	server Server
	conn   Connection
}

func (e commandExecutor) execute(cmd command.Command) error {
	switch typedCommand := cmd.(type) {
	case command.Ping:
		return e.executePing(typedCommand)
	case command.Echo:
		return e.executeEcho(typedCommand)
	case command.Info:
		return e.executeInfo(typedCommand)
	case command.Get:
		return e.executeGet(typedCommand)
	case command.Set:
		return e.executeSet(typedCommand)

	// Replication Handling
	case command.ReplConf:
		return e.executeReplConf(typedCommand)
	case command.PSync:
		return e.executePSync(typedCommand)
	}

	return fmt.Errorf("unknown command: %T", cmd)
}

func (e commandExecutor) executePing(_ command.Ping) error {
	if _, err := e.conn.Write([]byte("+PONG\r\n")); err != nil {
		return fmt.Errorf("error writing reponse to PING command to client: %w", err)
	}
	return nil
}

func (e commandExecutor) executeEcho(echo command.Echo) error {
	resStr, err := command.Encoder{}.Encode(echo.Payload)
	if err != nil {
		return fmt.Errorf("error encoding response for ECHO command: %w", err)
	}

	if _, err := e.conn.Write([]byte(resStr)); err != nil {
		return fmt.Errorf("error writing reponse to ECHO command to client: %w", err)
	}
	return nil
}

func (e commandExecutor) executeGet(get command.Get) error {
	responseString := command.NullBulkString

	data, ok := e.server.Get(get.Payload)
	if ok {
		var err error
		responseString, err = command.Encoder{}.Encode(data)
		if err != nil {
			return fmt.Errorf("error encoding response for GET command: %w", err)
		}
	}

	if _, err := e.conn.Write([]byte(responseString)); err != nil {
		return fmt.Errorf("error writing reponse to GET command to client: %w", err)
	}

	return nil
}

func (e commandExecutor) executeSet(set command.Set) error {
	e.server.Set(set.KeyPayload, set.ValuePayload, set.ExpiryTimeMs)

	if _, err := e.conn.Write([]byte(command.OKString)); err != nil {
		return fmt.Errorf("error writing reponse to SET command to client: %w", err)
	}

	return nil
}

// TODO: Testing once fn returns are a bit more stable
func (e commandExecutor) executeInfo(info command.Info) error {
	serverInfo, err := GetServerInfo(e.server, info.Payload)
	if err != nil {
		return fmt.Errorf("error getting server info: %w", err)
	}

	// Sort and join info with new lines
	infoToEncode := make([]string, 0, len(serverInfo))
	for key, val := range serverInfo {
		infoToEncode = append(infoToEncode, fmt.Sprintf("%s:%s\n", key, val))
	}
	slices.Sort(infoToEncode)

	encoder := command.Encoder{UseBulkStrings: true}
	res, err := encoder.Encode(strings.Join(infoToEncode, ""))
	if err != nil {
		return fmt.Errorf("error encoding response for INFO command: %w", err)
	}

	if _, err := e.conn.Write([]byte(res)); err != nil {
		return fmt.Errorf("error writing reponse to INFO command to client: %w", err)
	}

	return nil
}

func (e commandExecutor) executeReplConf(replConf command.ReplConf) error {
	nodeType := e.server.NodeType()
	switch nodeType {
	case MasterNodeType:
		if replConf.IsAck() {
			e.server.Logger().Info("Master node received ACK from replica", zap.String("offset", replConf.ValuePayload))
			return nil
		}
	case ReplicaNodeType:
		if replConf.IsGetAck() {
			encoder := command.Encoder{UseBulkStrings: true}
			res, err := encoder.Encode([]any{"REPLCONF", "ACK", "0"})
			if err != nil {
				return fmt.Errorf("error encoding response for REPLCONF ACK command: %w", err)
			}

			if _, err := e.conn.Write([]byte(res)); err != nil {
				return fmt.Errorf("error writing reponse to REPLCONF command to master: %w", err)
			}

			replica, ok := e.server.(*ReplicaServer)
			if !ok {
				return errors.New("REPLCONF GETACK processed for node with type ReplicaNode, but failed to cast to ReplicaServer")
			}
			replica.SetIsSteadyState(true)

			return nil
		} else if replConf.IsListeningPort() {
			if _, err := e.conn.Write([]byte(command.OKString)); err != nil {
				return fmt.Errorf("error writing reponse to REPLCONF command to client: %w", err)
			}
			return nil
		}
	}

	return fmt.Errorf("node of type %q received unknown REPLCONF command: %v", nodeType, replConf)
}

//////////////////////////
// Master Only Commands //
//////////////////////////

// TODO: Send full rdb file to replica
func (e commandExecutor) executePSync(_ command.PSync) error {
	master, ok := e.server.(*MasterServer)
	if !ok {
		return errors.New("received a PSYNC command on a non-master server")
	}

	if _, err := e.conn.Write([]byte(fmt.Sprintf("+FULLRESYNC %s 0\r\n", command.HARDCODE_REPL_ID))); err != nil {
		return fmt.Errorf("error writing reponse to PSYNC command to client: %w", err)
	}

	emptyRDB := command.GetHardedCodedEmptyRDBBytes()
	if _, err := e.conn.Write([]byte(fmt.Sprintf("$%d\r\n%s", len(emptyRDB), emptyRDB))); err != nil {
		return fmt.Errorf("error writing RDB file response to PSYNC command to client: %w", err)
	}

	master.registeredReplicaConns = append(master.registeredReplicaConns, e.conn)

	return nil
}
