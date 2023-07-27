# protocols



## Hidden Services

### Creating a hidden service

```sequence
Note left of alice: intro ->
note left of alice: intro route ->
alice-->bob: forward
bob-->charlie: forward
charlie-->dave: forward
note left of dave: -> intro route
note left of dave: propagate intro <-
dave-->charlie: gossip [intro]
dave-->eve: gossip [intro]
dave-->bob: gossip [intro]
```
### Requesting connection from Introducer (Routing Request)

```sequence
Note left of eve: routing request [reply header A] ->
eve-->bob: forward
bob-->charlie: forward
charlie-->alice: forward
note right of alice: <- Fwd alice, intro route[ready[reply eve]]
alice-->charlie: forward

charlie-->bob: forward
bob-->eve: forward
note left of eve: received ready
```

### Request/Response Cycle

```sequence
Note left of faith: forward(faith,alice)[request[reply faith]] ->
faith-->eve: forward eve
eve-->dave: forward eve
dave-->charlie: forward faith
charlie-->bob: forward reply faith
bob-->alice: forward reply faith
Note right of alice: <- forward(alice,faith)[response[reply alice]]
alice-->bob: forward alice
bob-->charlie: forward alice
charlie-->dave: forward faith
dave-->eve: forward faith
eve-->faith: forward faith
note left of faith: forward(faith,alice)[request[reply faith]] ->
faith-->eve: etc
eve-->dave: etc
dave-->charlie: etc
charlie-->bob: etc
bob-->alice: and so on
Note right of alice: <- forward(alice,faith)[response[reply alice]]

```

