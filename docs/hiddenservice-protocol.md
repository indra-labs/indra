# Hidden Service Message Protocol

## Creating a hidden service

```sequence
Note left of alice: return header A ->
Note left of alice: hidden service ->
alice-->bob: relay
bob-->charlie: relay
charlie-->dave: relay
note right of dave: <-introducer
note left of dave: return header A <-
dave-->charlie: gossip intro
dave-->eve: gossip intro
dave-->bob: gossip intro
```
## Requesting connection from introducer

```sequence
Note left of eve: return header B ->
eve-->dave: routing request
Note left of dave: -> return header A
dave-->charlie: relay
charlie-->bob: relay
bob-->alice: relay
note right of alice: <- return header B
alice-->bob: ready
note left of alice: return header C <-
bob-->charlie: 
charlie-->eve: 
note left of eve: return header C ->
```
## Request/Response Cycle

```sequence
note left of eve: return header C ->
Note left of eve: request message ->
Note right of eve: -> return header D
eve-->charlie: return onion
charlie-->bob: return onion
bob-->alice: return onion
Note right of alice: <- return header D
Note left of alice: reply message <-
note left of alice: return header E <-
alice-->bob: next request...
bob-->charlie:
note right of eve: <- return header E
charlie-->eve: 
note left of eve: new request message ->
note left of eve: return header F ->
eve-->charlie: etc
charlie-->bob: etc
bob-->alice: and so on

```

