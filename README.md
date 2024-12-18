Go Redis implementation for [codecrafters.io](https://codecrafters.io)

## Usage

Run `./spawn_redis_server.sh` to start the server and try running some commands using the [redis-cli](https://redis.io/docs/latest/develop/connect/cli/)

Ex.)

- `redis-cli PING` -> `PONG`

- `redis-cli ECHO test` -> `test`

- `redis-cli SET key value` -> `OK`

- `redis-cli GET key` -> `value`

Set a value with a lifetime of one second
`redis-cli SET key value px 1000` -> `OK`

## Replica Set

A replica set can be set up using the by setting up a master and pointing some replica nodes at it

`./spawn_redis_server.sh --port <PORT-A>"`
`./spawn_redis_server.sh --port <PORT-B> --replicaof "localhost <PORT-A>"`
`./spawn_redis_server.sh --port <PORT-C> --replicaof "localhost <PORT-A>"`
