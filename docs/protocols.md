# protocols



## Hidden Services

### 1. Creating a hidden service

In this phase, Alice, who wants to receive inbound connections via a hidden service, sends out an introduction request to Dave, who will serve as introducer, via a 3 hop forward path.

As introducer, Dave will now gossip the `intro` over the gossip network (Kademlia DHT Pub/Sub), and everyone will have this intro or be able to query neighbours in case they didn't receive it.

```sequence
Title: Creating a hidden service
Note over Alice: "hidden service"\nforward Alice\n[intro route]\n[intro Alice]\n->
Alice-->Bob: forward\n[Alice]
Bob-->Charlie: forward\n[Alice]
Charlie-->Dave: forward\n[Alice]
note over Alice,Dave: forward Alice ->
note over Dave: ->\n[intro Alice]\n[intro route Alice]\n"introducer"
note over Alice,Faith: <- Dave gossip [intro Alice] ->
```
At this point Bob, Charlie and Eve now know about Alice's hidden service 

### 2. Requesting connection from Introducer (Routing Request)

Faith has acquired an `intro` by some means and wishes to establish a connection to the hidden service. 

She forwards a route request, with an attached reply header to route a reply back to her, to Dave, the introducer for this illustration.

Dave then wraps up the route request and reply messages in the `intro route` it received along with the `intro`, which is used to forward one (1) request back to the hidden service.

> !!!! Here it might be good to mention that there is a flood attack vector here with creating unlimited numbers of `intro route`/`intro` over the gossip network. This, and the current lack of accounting for traffic to hidden services are both intertwined elements that fix this vulnerability. Nodes simply will require a high fee to accept an introduction, big enough that the spam use case is linearly more expensive than honest use.
>
> For the MVP this functionality will not be implemented, but we are already aware of it and will complete this after MVP.
>
> Also note there is no risk here of these 3 `intro route` packages being an avenue to unmasking, though in the request/response cycle this requires two forwards from the sender to ensure the receiver is not trying to unmask the client.

Alice then will receive the route request, with Faith's reply packet. Alice places her forward prefix, in case Faith also controls the first hop and thus would unmask Alice. This prefix is required in all messages in addition to the 3 hop reply in order to prevent either party unmasking each other.

This is not required for the first two steps of this part of the protocol because everyone knows the introducer, and neither client nor server would gain anything by controlling the adjacent hops on Dave's end of the path (last inbound, first outbound). But an attacker would want to attempt to unmask Alice, or a malicious hidden service would try to unmask Faith, and both cases are covered by each side adding their own two hops prior to the provided reply path.

```sequence
Title: Routing Request (establishing connection to hidden service)
Note over Faith: forward->Dave\n[route request Alice]\n[reply Faith]
Faith-->Eve: forward\n[Faith]
Eve-->Bob: forward\n[Faith]
Bob-->Dave: forward\n[Faith]
Note over Faith,Dave: forward->Dave
note over Dave: -> intro route Alice\n[route request Alice]\n[reply Faith]
Dave-->Charlie: intro route\n[Alice]
Charlie-->Gavin: intro route\n[Alice]
Gavin-->Alice: intro route\n[Alice]
Note over Dave,Alice: intro route Alice ->
Note over Faith,Alice: intro route Alice via Dave ->
note right of Alice: -> route request Alice\n[reply Faith]
note over Alice: forward Alice\n[Faith<-reply]\n[ready]\n[reply Alice]
Alice-->Charlie: forward\n[Alice]
Charlie-->Dave: forward\n[Alice]
note over Alice,Dave: <- prefix Alice
Dave-->Bob: forward\n[Faith]
Bob-->Eve: forward\n[Faith]
Eve-->Faith: forward\n[Faith]
note over Dave, Faith: <- forward path Faith
note over Faith,Alice: <- ready reply Alice to Faith
note over Faith: <-\nready\n[reply Alice]
Faith -> Alice: connection established
Alice -> Faith:
```

### 3. Request/Response Cycle

Once the receiver has the 'ready' signal, it can then begin a process of request and response wherein each reply carries the reply route header for the return path.

If a message fails the the parties keep past keys to decrypt latent messages or if it appears the outbound message may have got lost to retry a message using an older key since the key change message may have failed to get across before it arrived in the receiver's buffer.

Engineering more reliability into this requires the use of split/join message layers and layer two error correction compositions.

```sequence
Title: Hidden Service Request and Response
Note over Faith: forward->Dave\n[reply Alice]\n[request Faith]\n[reply Faith]\n->
Faith-->Eve: forward\n[Faith]
Eve-->Dave: forward\n[Faith]
Note over Faith,Dave: Faith's Forward Prefix -->
Dave-->Charlie: forward\n[reply Alice]
Charlie-->Bob: forward\n[reply Alice]
Bob-->Alice: forward\n[reply Alice]
Note over Dave,Alice: Alices's Reply Path -->
Note right of Alice: -> request Faith\n[reply Alice]
Note Over Faith,Alice: message round 1 (request from hidden client)
Note over Alice: forward Alice\n[forward Faith]\n[response Alice]\n[reply Alice]\n<-
Alice-->Bob: forward\n[Alice]
Bob-->Charlie: forward\n[Alice]
Note over Alice,Charlie: <-- Alice's Forward Prefix
Charlie-->Dave: forward\n[reply Faith]
Dave-->Eve: forward\n[reply Faith]
Eve-->Faith: forward\n[reply Faith]
Note over Charlie,Faith: <-- Faith's Reply Path
Note left of Faith: <-\nresponse Alice\n[reply Alice]
Note Over Faith,Alice: message round 2 (reply from hidden service)
```

