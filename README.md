# go-eventlogger

`go-eventlogger` is a flexible event system libray implemented as a pub/sub model supporting middleware. 

The library's clients submit events to a Broker that routes them through a pipeline. Each node in pipline can modifying the event, filtering it, persisting it, etc.  
