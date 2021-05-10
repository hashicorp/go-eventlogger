# go-eventlogger

`go-eventlogger` is a flexible event system libray implemented as a pub/sub model supporting middleware. 

The library's clients submit events to a Broker that routes them through a pipeline. Each node in pipline can modifying the event, filtering it, persisting it, etc.  

# Stability Notice 

While this library is fully open source and HashiCorp will be maintaining it (since we are and will be making extensive use of it), the API and output format is subject to minor changes as we fully bake and vet it in our projects. This notice will be removed once it's fully integrated into our major projects and no further changes are anticipated.


# Usage

An Event is a collection of data, analogous to a log entry, that we want to process in a Graph.  The client provides an event type and payload, and any other fields are generated as part of processing. The library will not attempt to discover whether configured formatter/marshaller nodes can actually handle the arbitrary payloads; it is up to the encapsulating program to put any such constraints on the user via its API.

The library's clients submit events to a Broker that routes them through a Graph based on their type.  A Graph is composed of Nodes.  A Node processes an Event in some way -- modifying it, filtering it, persisting it, etc.  A Sink is a Node that persists an Event.

## Broker

Clients interact with the library via the Broker.

A Broker processes incoming Events, by sending them to the Graph associated with the Event's type.  A given Broker, along with its associated set of Graphs, will be configured programmatically. 


## Nodes 

A Node is a node in a Pipeline, that can perform any kind of operation that it wants to on an Event.  A node has a Type, one of: Processor, Formatter, Sink.

Examples of things that a Node might do to an Event include:

Modify the Event, by storing a change description in Mutations.  Changes could be described as a (jsonpointer, interface{}) key-value pair.
Filter the Event, by returning nil.
Get the Event ready for a sink by rendering it in some way, e.g. as JSON, so that downstream Sinks in the graph can then write it without any extra work.  Rendered events will be stored in the Formatted map.


## Pipeline 

A Pipeline is a pointer to the root of an interconnected sequence of Nodes. All pipelines must start with a JSON Formatter node. All pipelines must end with a sink node.


# Contributing 

First: if you're unsure or afraid of anything, just ask or submit the issue or pull request anyways. You won't be yelled at for giving your best effort. The worst that can happen is that you'll be politely asked to change something. We appreciate any sort of contributions, and don't want a wall of rules to get in the way of that.

That said, if you want to ensure that a pull request is likely to be merged, talk to us! A great way to do this is in issues themselves. When you want to work on an issue, comment on it first and tell us the approach you want to take.
