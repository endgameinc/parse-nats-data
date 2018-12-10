# Parse NATS Streaming Server Data

This is a tool to parse the contents of [NATS streaming server](https://github.com/nats-io/nats-streaming-server/) durable subscription. 

From the [nats.io documentation](https://nats.io/documentation/streaming/nats-streaming-intro/):

> Durable subscriptions - Subscriptions may also specify a “durable name” which will survive client restarts. Durable subscriptions cause the server to track the last acknowledged message sequence number for a client and durable name. When the client restarts/resubscribes, and uses the same client ID and durable name, the server will resume delivery beginning with the earliest unacknowledged message for this durable subscription.

## Installation

```
go get -u github.com/endgameinc/parse-nats-data
```

## Usage

```
# ./parse-nats-data

Usage: ./parse-nats-data [options]

Options:
    -d, --msgsDat       msgs.dat file
    -i, --msgsIdx       msgs.idx file
    -s, --subsDat       subs.dat file
    -c, --clientsDat    clients.dat file
```

To read messages, run `parse-nats-data` with the following arguments: 

```
# ./parse-nats-data -d /path/to/msgs.1.dat -i /path/to/msgs.1.idx
[2018-07-10 15:00:31.716244931 +0000 UTC] seq: 1 | size: 936 | offset: 4 | crc32: 8e22ea12
[2018-07-10 15:00:31.836288178 +0000 UTC] seq: 2 | size: 959 | offset: 948 | crc32: 7cccde1c
[2018-07-10 15:00:36.202716151 +0000 UTC] seq: 3 | size: 969 | offset: 1915 | crc32: 562b1ffc
[2018-07-10 15:00:50.33685147 +0000 UTC] seq: 4 | size: 958 | offset: 2892 | crc32: 256e9151
[2018-07-10 15:00:53.790236525 +0000 UTC] seq: 5 | size: 994 | offset: 3858 | crc32: 3c5ad77
[2018-07-10 15:00:54.425182393 +0000 UTC] seq: 6 | size: 983 | offset: 4860 | crc32: 2fe42320
[2018-07-10 15:00:55.447482123 +0000 UTC] seq: 7 | size: 982 | offset: 5851 | crc32: fd8ff587
[2018-07-10 15:00:56.251222498 +0000 UTC] seq: 8 | size: 984 | offset: 6841 | crc32: 2ec77197
[2018-07-10 15:02:47.133969633 +0000 UTC] seq: 9 | size: 980 | offset: 7833 | crc32: 83c370d4
[2018-07-10 15:02:47.716276067 +0000 UTC] seq: 10 | size: 995 | offset: 8821 | crc32: bd8ce709
[2018-07-10 15:02:50.598110162 +0000 UTC] seq: 11 | size: 891 | offset: 9824 | crc32: f0bdc27b
[2018-07-10 15:02:51.031120836 +0000 UTC] seq: 12 | size: 899 | offset: 10723 | crc32: 175b6e1b
```

To read subscription events, run `parse-nats-data` with the following arguments:

```
# ./parse-nats-data -s /path/to/subs.dat
ID: 1 'Client_6715' Type: subRecNew LastSent: 0 qGroup:  Inbox: _INBOX.4SR1QZIfwYLHPtQYrb1AzQ AckInbox: _STAN.subacks.54V6hblM4Dt4B7KyUzRYy5.sub:queue.neCdF3TS8SIjOY1fpRavR0 MaxInFlight: 1024 AckWaitInSecs: 30 Durable:  IsDurable: false IsClosed: false
ID: 2 'Client_10792_bcff6c27' Type: subRecNew LastSent: 0 qGroup: sub:queue Inbox: _INBOX.Jby2tc5tiIBlZvYSnfBUOR AckInbox: _STAN.subacks.54V6hblM4Dt4B7KyUzRYy5.sub.neCdF3TS8SIjOY1fpRavss MaxInFlight: 8192 AckWaitInSecs: 30 Durable:  IsDurable: true IsClosed: false
ID: 3 'Client_11933_4794405c' Type: subRecNew LastSent: 0 qGroup: sub:queue Inbox: _INBOX.KoEbjf9Czr4kk62w3eLslO AckInbox: _STAN.subacks.54V6hblM4Dt4B7KyUzRYy5.sub.neCdF3TS8SIjOY1fpRawgQ MaxInFlight: 8192 AckWaitInSecs: 30 Durable:  IsDurable: true IsClosed: false
ID: 1 SeqNo: 1 subRecMsg
ID: 2 SeqNo: 1 subRecMsg
ID: 3 SeqNo: 1 subRecMsg
ID: 1 SeqNo: 1 subRecAck
ID: 3 SeqNo: 1 subRecAck
ID: 1 SeqNo: 2 subRecMsg
ID: 2 SeqNo: 2 subRecMsg
```

To read client events, run `parse-nats-data` as follows:

```
# ./parse-nats-data -c clients.dat 
CID: Client_40887 Inbox: _INBOX.cggINaUbPQHIjTwxsJSIuz ConnId: [99 103 103 73 78 97 85 98 80 81 72 73 106 84 119 120 115 74 83 73 115 73] Protocol: 1 PingInterval: 5 PingMaxTimeout: 3 addClient
CID: Client_40887 Inbox: _INBOX.cggINaUbPQHIjTwxsJSIpb ConnId: [99 103 103 73 78 97 85 98 80 81 72 73 106 84 119 120 115 74 83 73 109 117] Protocol: 1 PingInterval: 5 PingMaxTimeout: 3 addClient
CID: Client_40887 Inbox: _INBOX.cggINaUbPQHIjTwxsJSJGX ConnId: [99 103 103 73 78 97 85 98 80 81 72 73 106 84 119 120 115 74 83 74 68 113] Protocol: 1 PingInterval: 5 PingMaxTimeout: 3 addClient
CID: Client_40887 delClient
CID: Client_40887 Inbox: _INBOX.cggINaUbPQHIjTwxsJSJpY ConnId: [99 103 103 73 78 97 85 98 80 81 72 73 106 84 119 120 115 74 83 74 109 114] Protocol: 1 PingInterval: 5 PingMaxTimeout: 3 addClient
CID: Client_40887 Inbox: _INBOX.cggINaUbPQHIjTwxsJSKDn ConnId: [99 103 103 73 78 97 85 98 80 81 72 73 106 84 119 120 115 74 83 75 66 54] Protocol: 1 PingInterval: 5 PingMaxTimeout: 3 addClient
CID: Client_40887 delClient
CID: Client_40887 Inbox: _INBOX.cggINaUbPQHIjTwxsJSKZL ConnId: [99 103 103 73 78 97 85 98 80 81 72 73 106 84 119 120 115 74 83 75 87 101] Protocol: 1 PingInterval: 5 PingMaxTimeout: 3 addClient
CID: Client_40887 delClient
```
