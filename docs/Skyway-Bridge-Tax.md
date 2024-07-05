# Skyway Bridge Tax

All outbound transactions from the bridge to the target EVM and other chains pay
a tax on the skyway bridge. This tax is added to the cost of the transfer.
A governance vote is needed to define the tax rate for a specific token, as well
as a list of addresses that are exempt from the bridge tax.

## Tax Rate

The tax rate must be defined as a non-negative value, with 0 meaning no tax is
applied. This is the ratio of the tax applied on top of the transfer, so a tax
rate value of 0.2 means a 20% tax.

The tax is added to the cost of the transfer and will stay locked until the
transfer is finished.
If the transfer is successful, the taxed amount is burned on the Paloma side.
If a transfer is canceled before being executed, the full initial amount, plus
tax, is refunded.

## Exempt Addresses

Similarly, the governance vote can define a list of addresses that are exempt
from paying the bridge tax. Transfers from these senders will never pay bridge
tax.
