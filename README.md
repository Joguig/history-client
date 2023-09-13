# history.v2

This is the version 2 architecture of [history-service]. Major change is the
entrypoint to the architecture is now [kinesis stream]. Currently,
[history-service] is dual writing to both architectures.


[history-service]: https://git-aws.internal.justin.tv/foundation/history-service/commits/chore/admin-387/use-kinesis-for-es
[kinesis stream]: https://aws.amazon.com/kinesis/data-streams/
