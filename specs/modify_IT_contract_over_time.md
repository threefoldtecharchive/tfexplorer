# The problem: update IT contract over time

At the moment, there is no way for a grid user to modify anything in the workloads that are live on the grid. The IT contract is something that is agreed by the user and all involved farmers before.

Once an IT contract is signed by all parties, it is  provisioned by the 3 nodes and the workloads come to life.

If the reservation has been made for a duration of 6 months, there is no way for the user to cancel it after 3 months and not loose the tokens for the last 3 months. A manual discussion between the user and all the farmers  may be be initiated but there is no garantee the farmer will accept to refund the remaining 3 months and this process is impossible to scale.

## Possible solutions

### Decouple the reservation of capacity from the actual workloads description

This solution splits the reservation of capacity from the workloads consuming this capacity.
In order to implement this solution a possible approach would be to create some kind of capacity pool entity. A user would reserve capacity from different farms. Once the capacity is reserved, they could provision/decommission workloads on these farms freely without worrying when deleting a workloads since then the capacity would just be given back to the pool and no tokens would be "lost".

If a user needs to extend the life of its workloads he just has to make sure the reserved capacity from its pools is always funded enough.

Here are the 2 new flows that would need to be implemented:

#### Creation of the capacity pool

![reserve_capacity](reserve_capacity.png)

This flow is very similar to the current flow to deploy workloads. The main difference is only that the amount of resource units is defined instead of a full workloads definition.

Both the farmer and the user still need to sign the reservation to mark their agreement on the deal. Multi signature is still possible using the same signing request construct that exists today.

The main difference is during the call numbered 3 in the diagram. When the user sends the capacity reservation to the explorer, the explorer will block. During this time, the explorer will send the capacity reservation to the farmer and wait for him to answer (until the farmer 3bot is reality, the explorer takes the role of the farmer and is responsible to make sure the farm has enough capacity to sell).
If enough capacity is available in the farm, the explorer will mark the capacity reservation as to be paid and sends the payment detail to the user.
If there is not enough free capacity in the farm, the explorer will return an error to the user. The error will include the reason details. The user can then modify it's request and retry.

When the farmer confirms the capacity is locked, the explorer will create a capacity pool object that defines the amount of capacity reserved by the pool. The pool object also contains the expiration date of the pool. This will be used by the node over time to periodically check back in the explorer and make sure the pool has been extended. If a node has some workloads that are part of a pool that is about to expire, a call is made to the explorer to ask what is the expiration date of the pool. If the pool has not been extended, all the workloads linked to this pool will be decommissioned.

With this system in place, expirations are not needed on the workload definition anymore. It also greatly reduces the amount of time the tokens are locked in the explorer, since as soon as the tokens are received on the escrow account, they can be forwarded directly to the farmer.

If for some reason the client fails to pay the capacity reservation in time, no token transfer is involved at all (unless the client pays only a part of the amount, in which case a refund needs to happen)

#### Workloads deployment

![deploy_workload](deploy_workload.png)

This flow is greatly simplified compared to how it works today. The only party involved are the user, the explorer and the nodes. The farmer and blockchain are no longer involved.

There are still some things to take in to account here. In order to avoid over-provisioning and try to return early, the explorer will keep track of how much resource is available in a capacity pool. So when a workloads definition is received, it first checks if the total amount of resource needed is available in the pool.
This check needs to be atomic in the explorer, meaning that 2 workloads for the same pool need to be processed one at a time.
This is what we see in the step 3 and 4 of the diagram. The modification of the pool capacity is done before the workload definition is sent to the node. This is a protection against concurrent requests trying to deploy workloads using the same capacity pool. By modifying the pool capacity early we prevent over-provisioning.

The pool resource is also updated when a workloads is decommissioned from a node. The explorer does this when it receives a result with the state deleted from a node.

### Todo

- I did not yet measure the impact of those change regarding all the different currency supported. We should try to make sure to avoid the network split generated by the FreeTFT token. cf. https://github.com/threefoldtech/tfexplorer/issues/50
- The exact schema of the different new objects still needs to be defined in detail
- Define clearly all the modification in the code that would be needed for all component to implement this proposal

### Extra ideas

With this design, the capacity reservation is not linked to a single farm but to a list of nodes. This property would allow multiple farmers to create some kind of "virtual farm" where all farmers contribute some capacity and are rewarded based on the % of CPR they allocated to the virtual farm. Think crypto mining pool but for IT capacity. This is especially interesting for very small farmers that have less chance to get market shares. This needs to be further thought out and is not part of this reflection just yet.

### Extension of the capacity pool

The capacity pool solves the problem of the workload extension, but there is still
the issue of user potentially paying more than what he uses. This is the result
of reserving a fixed amount of capacity over some period of time, and we assume
the full capacity to be utilized the entire time. Even though the system could
detect that some workload has failed to deploy, its effect of the pool will not
be refunded. In the proposed scheme this is not really a problem since a user
would cancel his reservation and redeploy anyway. But the capacity pool does open
up an interesting possibility: now that we have an abstraction which can be
considered as representing the users payment, we can implement a system which
only charges the user for the amount of capacity that he uses at any given point
in time.

If we ignore discounts, we can look at the price of a workload as:
`RU * T * RU_PRICE/T`. Since this is a simple multiplication of 2 linear scaling
values (price and time) and a constant (the RU price over time), we can come to
the same price in multiple ways. For instance, a client could reserve 10 SU and
10 CU for 10 days. He pays X amount of tokens for this, regardless of whether he
uses the pool or not. In fact, in this scheme, a pool could be thought of as the
ability to deploy a set of workloads for some time, which is an allocation rather
than a pool. If another client reserves 20 CU and 20 SU for 5 days, he would
effectively pay just as much.

This proposal would merge the time factor with the resource unit factor. As such,
in the pool, CU and SU would be presented as "CUs" and "SUs", Compute unit seconds
and Storage unit seconds. These values represent the maximum amount of seconds a
single CU and SU can be deployed from the pool. In effect, every second a CU or SU
is deployed, they drain a single CUs or SUs from the pool. Of course, if a workload
is calculated to be 2 SU, it will drain 2 SU per second it is deployed, and so on.
With this, there is no longer a fixed period of time for which a pool is valid.
The result of this is, that a user can cancel workloads at any time. If he does,
the pool will simply last longer. If no workload is defined, the pool does not drain
at all. Since the pool can be mapped to the tokens the user paid, he no longer
pays if he cancels a workload sooner, effectively allowing us to have a "pay for
what you use" strategy.

On the explorer side of things, it knows all workloads which are deployed. As such
it knows exactly when a pool will expire. This allows us to track the available
capacity in an efficient way. Every time it receives a reservation, or a result
from the node, it can check the pool, decrement the available cloud units in the
pool (since it knows when the last time it decremented its values was, what the
time now is, and how much CU and SU is deployed), and recompute when it will expire
based on the deployed workloads. The pool can then set up a timer at this time,
in case nothing changes. Once the pool runs out, the workloads tied to it are
decomissioned. It is the clients job to send reservations for the pool to extend
its available capacity, to ensure the pool remains active. Should this be needed,
the explorer could, for example, reports the amount of time left in a pool every
time it gets a workload reservation (assuming all workloads deploy successfully).
Since the explorer knows when a reservation has been deployed (with or without error),
tracking of the workloads influence on the pool can be started at this moment. This
means that the user also no longer pays for the time it takes to deploy a workload,
as is the case currently. Likewise, if a deployment failed, it will not be taken
into account in the pool usage, even if the user does not cancel the reservation.

In the future, the explorer can be extended to also detect if a workload is still
live. To do this, a node could send a status of the workloads every X amount of time
to the explorer. If the workload is not healthy, it will no longer affect the pool,
until such a time where the workload is repaired. This is an improvement over the
current situation where a node going down after a workload is deployed will still
give all tokens to the farmer, even though the user does not really get anything.
Obviously, next to the dedicated workload health reporting, the explorer can verify
that nodes are actually online through their uptime reports, and stop accounting
for a workload in a pool if a node is considered to have stopped.

This also slots nicely into the inevitable concept of network units (yet to
be defined and implemented into zos). In this scenario, the NU can represent an
amount of total bandwidth. The bandwidth limit (max throughput) can be set on
the network itself. Periodically, nodes report how much throughput there was over
a period of time, and the explorer can update the NU in the pool. If NU is empty,
the explorer could instruct the nodes to shut down inter-network communication.
Note: this paragraph is speculation about how the NU implementation could look,
in this approach to the capacity reservation.

Now some downsides. Since it is no longer known when and how the pool will be used,
capacity planning is pretty hard. The idea for now would be to not worry about this,
however. Is is the responsibility of the user to select nodes which have enough
capacity to deploy individual workloads. If there is no capacity available on any
of the nodes in the pool, this likely means that there is a lot of usage of the grid.
In reality, today this is not the case. But even if it were, this is something
that might be approached completely differently on a farmer by farmer basis.
Therefore, this is a non issue which should be solved by the arrival of the
(farmer) threebot. In case a user deploys after only a certain time, the same
logic applies. Importantly, an extension of an existing workload can never
fail since it is already deployed. The client simply extends the pool, and the
workload won't be decomissioned.

All in all, there is also still no refund for cancelling a workload before it
would expire. At least, not automated through our system. Again, this would not
really be an issue. Any decent sized workloads could be started from a pool
which only has a small initial amount of capacity. As time passes, a user simply
needs to extend the pool, by making a reservation to add capacity and paying it.
Since this is limited functionality, it could even be integrated into the
3bot wallet app.

But the implementation of the capacity pool means that an actual refund is no longer
required. In fact the payment flow is greatly simplified. The user reserves a
capacity pool, and the explorer sends back how much the user should pay. If the user
paid enough, the pool is created (or extended if it already exists). The farmer
gets paid, and any overpaid amount is refunded. Similarly, if the pool is not paid
in time, the client is refunded everything he already paid. Regular workloads no
longer require payments. This also means that for a regular workload, the deployment
process is started instantly, as we no longer need to monitor the blockchain
permanently. And in fact, once the farmer threebot is here, an approach could
be taken where a user pays the farmer directly. In that scenario, a reservation
would require an additional signature from the foundation. The explorer simply
monitors the blockchain for payments, and checks is a payment of (at least) the
required amount is done to the farmer and the foundation (and tftech for certified
capacity).

#### TL;DR  

This solution/extension would bring a more "pay for what you use" style
approach, which should slot in nicely with the eventual usage of NU. The main
downside is the largely difficult capacity planning, though the low amount
of usage of the grid today, together with the fact that this is actually something
a farmer should decide on, once the farmer threebot is here, mean that this
is largely a non issue at the time of writing.
