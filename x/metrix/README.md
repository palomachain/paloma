# Paloma Metrix module

The `metrix` module is responsible for capturing all kinds of domain events
around relayer performance, aggregating them in a meaningful way and
making them accessible to consumers.

## The collected metrics

The following metrics are currently actively being gathered:

### Relayer Fee

> [!WARNING]  
> This is still very WIP, as the concept of user fees is yet to be embossed into Paloma.

### Uptime

The update is a percentile value of representing the network uptime of the valid
ator node attributed with the relayer.

It is updated once per block, and calculated by evaluating the nodes `missed
blocks` during the `signed blocks window` using the following formula:

`((params.signed_blocks_window - signingInfo.missed_blocks_counter) /
params.signed_blocks_window) * 100`

> [!WARNING]  
> This still needs to be adapted for jailed Validators, who will still be reported
> with an uptime of 100% (although not eligible for relaying).

### SuccessRate

This weight represents how important pigeon relaying success rate should be
during validator selection. The metric is collected every time a Pigeon reports
a message as processed, either successfully or not. 

The success rate goes up by 1% for every successfully relayed message, and down
by 2% for every failed attempt. The rate is capped between 0% and 100%, and will
slowly shift back towards a base rate of 50% from either end over the period of
1000 messages. 

### ExecutionTime

I have no idea what this one was intended to represent. My guess is pigeon
relaying time (i.e. faster → better). The metrics work much the same way as
on success rate, with a moving average built on the last 100 messages delivered,
self purging after 1000 messages.

### FeatureSet

This weight represents how important pigeon feature sets (MEV, TXtype,etc…)
should be during validator selection. 

> [!WARNING]  
> This is still a work in progress.
