# Paloma Metrix module

The `metrix` module is responsible for capturing all kinds of domain events
around relayer performance, aggregating them in a meaningful way and
making them accessible to consumers.

## The collected metrics

The following metrics are currently actively being gathered:

### Relayer Fee

**TODO** No idea if we need this one

### Uptime

The update is a percentile value of representing the network uptime of the valid
ator node attributed with the relayer.

It is updated once per block, and calculated by evaluating the nodes `missed
blocks` during the `signed blocks window` using the following formula:

`((params.signed_blocks_window - signingInfo.missed_blocks_counter /
params.signed_blocks_window) * 100`

### SuccessRate

This weight represents how important pigeon relaying success rate should be
during validator selection

### ExecutionTime

I have no idea what this one was intended to represent. My guess is pigeon
relaying time (i.e. faster → better)

### FeatureSet

This weight represents how important pigeon feature sets (MEV, TXtype,etc…)
should be during validator selection
