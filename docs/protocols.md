# protocols



## Hidden Services

### 1. Creating a hidden service

In this process, an intro, which is advertised, and the intro route, which is held, provides a way for a client to contact the hidden service via a temporary proxy called an "Introducer":

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
### 2. Requesting connection from Introducer (Routing Request)

The client uses the information from the intro, attaches a reply routing header, which includes ciphers and nonces for a two key ECDH cryptosystem used to return a reply. The routing request is bundled inside the intro route reply header, and sent on, following the path prescribed by the hidden service when the introduction was created.

```sequence
Note left of eve: routing request [reply header alice] ->
eve-->bob: forward
bob-->charlie: forward
charlie-->alice: forward
note right of alice: <- Fwd alice, intro route[ready[reply eve]]
alice-->charlie: forward

charlie-->bob: forward
bob-->eve: forward
note left of eve: received ready
```

### 3. Request/Response Cycle

Once the receiver has the 'ready' signal, it can then begin a process of request and response wherein each reply carries the reply route header for the return path. If a message fails the protocol assumes that old keys should be available for a few cycles after the current one for this case so the connection can resume rather than forcing the client back to step one.

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

