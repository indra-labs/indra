# protocols



## Hidden Services

### 1. Creating a hidden service

In this process, an intro, which is advertised, and the intro route, which is held, provides a way for a client to contact the hidden service via a temporary proxy called an "Introducer":

![creating hidden service sequence diagram as svg](/src/github.com/indra-labs/indra/docs/hidden1.svg)

```sequence
Title: Creating a hidden service
Note over Alice: [intro route]\n[intro]\n"hidden service"\n->
Alice-->Bob: forward 1\n[Alice]
Bob-->Charlie: forward 1\n[Alice]
Charlie-->Dave: forward 1\n[Alice]
note over Dave: ->\n[intro]\n[intro route]\n"introducer"
Dave-->Charlie: gossip [intro]
Dave-->Eve: gossip [intro]
Dave-->Bob: gossip [intro]
Dave-->Faith: gossip [intro]
note over Bob,Faith: all received [intro] 
note over Dave: can introduce
```
At this point Bob, Charlie and Eve now know about Alice's hidden service 

### 2. Requesting connection from Introducer (Routing Request)

The client uses the information from the intro, attaches a reply routing header, which includes ciphers and nonces for a two key ECDH cryptosystem used to return a reply. The routing request is bundled inside the intro route reply header, and sent on, following the path prescribed by the hidden service when the introduction was created.

![requesting connection from introducer sequence diagram as svg](/src/github.com/indra-labs/indra/docs/hidden2.svg)

```sequence
Title: Routing Request (establishing connection to hidden service)
Note over Faith: route\n[reply Faith]\n->
Faith-->Eve: forward\n(Faith)
Eve-->Bob: forward\n(Faith)
Bob-->Dave: forward\n(Faith)
Note over Faith,Dave: forward Faith
note over Dave: ->\nroute\n[reply Faith]\n[intro route]
Dave-->Charlie: forward\n[intro route]
Charlie-->Gavin: forward\n[intro route]
Gavin-->Alice: forward\n[intro route]
Note over Dave,Alice: intro route
note over Alice: <- ready\n[reply Faith]\n[reply Alice]
Alice-->Bob: forward\n(Faith)
Bob-->Eve: forward\n(Faith)
Eve-->Faith: forward\n(Faith)
note over Faith,Alice: reply Faith
note over Faith: <-\n[ready]\n[reply Alice]
note over Faith,Alice: <- connection established ->
```

### 3. Request/Response Cycle

Once the receiver has the 'ready' signal, it can then begin a process of request and response wherein each reply carries the reply route header for the return path. If a message fails the protocol assumes that old keys should be available for a few cycles after the current one for this case so the connection can resume rather than forcing the client back to step one.

```sequence
Title: Hidden Service Request and Response
Note over Faith: forward [Alice]\nrequest faith\n[reply faith]\n->
Faith-->Eve: forward Faith
Eve-->Dave: forward Faith
Note over Faith,Dave: Faith's Forward Prefix -->
Dave-->Charlie: forward Alice
Charlie-->Bob: forward reply Alice
Bob-->Alice: forward reply Alice
Note over Dave,Alice: Alices's Reply Path -->
Note over Alice: forward[Faith]\nresponse Alice\n[reply Alice]\n<-
Alice-->Bob: forward Alice
Bob-->Charlie: forward Alice
Note over Alice,Charlie: <-- Alice's Forward Prefix
Charlie-->Dave: forward Faith
Dave-->Eve: forward Faith
Eve-->Faith: forward Faith
Note over Charlie,Faith: <-- Alice's Reply Path
```

