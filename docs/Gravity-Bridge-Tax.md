# Gravity Bridge Tax

All outbound transactions from the bridge to the target EVM and other chains pay
a tax on the gravity bridge. When sending a Paloma token to EVM, the net tokens
minted on the target chain will be the total tokens less the token tax % amount.
A governance vote is needed to define the tax percentage amount, as well as a
list of tokens and addresses that are exempt from the bridge tax.

## Tax Rate

The tax rate must be defined as a value between 0 and 1, inclusive. 0 means no
tax is applied, and 1 would mean the transfer will not go through and just be
spent on tax instead.

The tax is subtracted from the intended transfer amount at the moment the
transfer is queued for processing. The final amount transferred is calculated as
`<intended_amount>*(1-<tax_rate>)`.
If the transfer is successful, the taxed amount is burned on the Paloma side.
If a transfer is canceled before being executed, the full initial amount,
including tax, is refunded.

## Excluded Tokens

The governance vote can define a list of tokens that are excluded from the
bridge tax. Transfers of these tokens will never pay bridge tax and will be
transferred in the full amount.

## Exempt Addresses

Similarly, the governance vote can define a list of addresses that are exempt
from paying the bridge tax. Transfers from these senders will never pay bridge
tax.
