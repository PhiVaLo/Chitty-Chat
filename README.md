# How to run the program

A node can input the following commands:

Run Databases:
```bash
go run ./database/Database.go -port 5050 -id 1
go run ./database/Database.go -port 5051 -id 2
go run ./database/Database.go -port 5052 -id 3
```

Run Nodes:
```bash
go run ./nodes/AuctionNode.go -port 8080 -id 1
go run ./nodes/AuctionNode.go -port 8081 -id 2 
```





```bash
bid <amount>
result
```

<amount> is an integer

--------------------------------------------------------------------

# Distributed Mutual Exclusion
Distributed Systems (B-SWU & K-SD, Autumn 2023)
* Mandatory Hand-in 5 - Replication
* Submision Due Date: *`Tuesday 28, November 2023, 23:59`*

## Report

- [Google Docs](https://docs.google.com/document/d/1CT_8YhAHZcycnY6ix9731VRISfUEH5zQjuLDVVoCBTo/edit)

<div style="text-align: center;">
    <img src="https://i.imgur.com/NcX1hkX.png" alt="Centered Image" width="300">
</div>

--------------------------------------------------------------------

# == A Distributed Auction System ==

## ::Introduction::

You must implement a **distributed auction system** using replication: a distributed component which handles auctions, and provides operations for bidding and querying the state of an auction. The component must faithfully implement the semantics of the system described below, and must at least be resilient to one (1) crash failure.


## ::MA Learning Goal::

The goal of this mandatory activity is that you learn (by doing) how to use replication to design a service that is resilent to crashes. In particular, it is important that you can recognise what the key issues that may arise are and understand how to deal with them.

 

## ::API::

Your system must be implemented as some number of nodes,  running on distinct processes (no threads). Clients direct API requests to any node they happen to know (it is up to you to decide how many nodes can be known). Nodes must respond to the following API

Method:  bid
Inputs:  amount (an int)
Outputs: ack
Comment: given a bid, returns an outcome among {fail, success or exception}

 

Method:  result
Inputs:  void
Outputs: outcome
Comment:  if the auction is over, it returns the result, else highest bid.

 

## ::Semantics::

Your component must have the following behaviour, for any reasonable sequentialisation/interleaving of requests to it:

- The first call to "bid" registers the bidder.

- Bidders can bid several times, but a bid must be higher than the previous one(s).

- after a predefined timeframe, the highest bidder ends up as the winner of the auction, e.g, after 100 time units from the start of the system. 

- bidders can query the system in order to know the state of the auction.

 

## :: Faults :: 

- Assume a network that has reliable, ordered message transport, where transmissions to non-failed nodes complete within a known time-limit.

- Your component must be resilient to the failure-stop failure of one (1) node.

 

## :: Report ::

Write a report of at most 3 pages containing the following structure (exactly create four sections as below):

- Introduction. A short introduction to what you have done.

- Architecture. A description of the architecture of the system and its protocols (behaviour), including any protocols used internally between nodes of the system. 

- Correctness 1. Argue whether your implementation satisfies linearisability or sequential consistency. In order to do so, first, you must state precisely what such property is. 

- Correctness 2. An argument that your protocol is correct in the absence and the presence of failures.

 

## :: Implementation ::

- Implement your system in GoLang. We strongly recommend that you reuse the the frameworks and libraries used in the previous mandatory activities.

- Submit a log (as a separate file) documenting a correct system run under failures. Your log can be a collection of relevant print statements, that demonstrates the control flow trough the system. It must be clear from the log where crashes occur.

 - Provide a README file with instructions on how to run your implementation. 

 

## **>>** SUMMARY of what to submit on learnit **<<**

- a link to a GitHub repo containing the code (please make sure the repo is public so that we can easily access it. Alternatively, you can submit a .zip file in learnit.

- A file report.pdf containing a report (in PDF) with your answers; the file can be **at most** 3 A4 pages (it can be less), font cannot be smaller than 9pt.

- A text file log.txt containing log(s).