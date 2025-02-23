## IDs (process definition, process instance, job, events, etc.)

This engine does use an implementation of [Twitter's Snowflake algorithm](https://en.wikipedia.org/wiki/Snowflake_ID)
which combines some advantages, like it's time based and can be sorted, and it's collision free to a very large extend.
So you can rely on larger IDs were generated later in time, and they will not collide with IDs,
generated on e.g. other nodes of your application in a multi-node installation.

The IDs are structured like this ...
```
+-----------------------------------------------------------+
| 41 Bit Timestamp |  10 Bit NodeID  |   12 Bit Sequence ID |
+-----------------------------------------------------------+
```

The NodeID is generated out of a hash-function which reads all environment variables.
As a result, this approach allows 4096 unique IDs per node and per millisecond.
