Go Redis implementation for [codecrafters.io](https://codecrafters.io)

Run `./spawn_redis_server.sh` to start the server and try running some commands using the [redis-cli](https://redis.io/docs/latest/develop/connect/cli/)

NOTE: Lots of missing support for fancier stuff

Ex.)
- `redis-cli PING` -> `PONG`

- `redis-cli ECHO test` -> `test`

- `redis-cli SET key value` -> `OK`

- `redis-cli GET key` -> `value`

Set a value with a lifetime of one second
`redis-cli SET key value px 1000` -> `OK`