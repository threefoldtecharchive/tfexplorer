# The problem: update IT contract over time

At the moment, there is no way for a grid user to modify anything in the workloads that are live on the grid. The IT contract is something that is agreed by the user and all involved farmers before hands.

Once a IT contract is signed by all parties, it is then provisioned by the 3Nodes and the workloads comes to life.

If the reservation has been made for a duration of 6 month. There is no way for the user to cancel it after 3 month and not loose the tokens for the last 3months. A manual discussion between the user and all the farmers can maybe be initiated but there is no grantee the farmer will accept to refund the remaining 3 months and this process is impossible to scale.

## Possible solutions

### Decouple the reservation of capacity from the actual workloads description

This solution splits the reservation of capacity from the workloads consuming this capacity.
In order to implement this solution a possible approach would be to create some kind of capacity pool entity. User would reserve capacity from different farms. Then once the capacity is reserved, they could provision/decommission workloads on these farms freely without worrying when deleting a workloads cause then the capacity would just be given back to the pool and no token would be "lost".

If a user needs to extend the live of its workloads he just have to make sure the reserved capacity from its pools is always funded enough.

Here are the 2 new flows that would need to be implemented:

#### Creation of the capacity pool

![reserve_capacity](reserve_capacity.png)

This flow is very similar to the current flow to deploy workloads. The main difference is only the amount of resource units if defined instead of a full workloads definition.

Both the farmer and the user still needs to sign the reservation to mark their agreement on the deal. Multi signature is still possible using the same signing request construct that exists today.

The main difference is during the call numbered 3 in the diagram. When the user send the capacity reservation to the explorer, the explorer will block. During this time, the explorer will sends the capacity reservation to the farmer and wait for him to answer (until the farmer 3bot is reality, the explorer takes the role of the farmer and his responsible to make sure the farm has enough capacity to sell).
If enough capacity is available in the farm, the explorer will mark the capacity reservation as to be paid and send the payment detail to the user.  
If there is not enough free capacity in the farm, the explorer will return an error to the user. The error will give detail about the error. The user can then modify it's request and retry.

When the farmer confirm the capacity is locked, the explorer will create an capacity pool object that define define the amount of capacity reserved by the pool. The pool object also contains the expiration date of the pool. This will be used by the node over time to periodically check back in the explorer and make sure the pool has been extended. If a node has some workloads that are part of a pool that is about to expire, a call is make to the explorer to ask what is the expiration date of the pool. If the pool has not been extended, all the workloads linked to this pool will be decommissioned.

With this system in place, expiration are not needed on the workload definition anymore. It also greatly reduce the amount of time the token are locked in the explorer, cause now as soon as they token are received on the escrow account, they can be forwarded directly to the farmer.

If for some reason the client fails to pay the capacity reservation in time, no token transfer is involved at all (unless the client pay only a part of the amount, then refund needs to happens)

#### Workloads deployment

![deploy_workload](deploy_workload.png)

This flow is greatly simplified compare to how it works today. The only party involved are the user, the explorer and the nodes. The farmer and blockchain are no longer involved.

There are still some things to take in account here. In order to avoid over-provisioning and try to return early, the explorer will keep track of how much resource is available in a capacity pool. So when a workloads definition is received, it first check if the total amount of resource needed is available in the pool.
This check needs to be atomic in the explorer, meaning that 2 workloads for the same pool needs to be process one at a time.
This is what we see in the step 3 and 4 of the diagram. The modification of the pool capacity is done before the workload definition is send to the node. This is a protection against concurrent request trying to deploy workloads using the same capacity pool. By modifying the pool capacity early we prevent over-provisioning.

The pool resource is also updated when a workloads is decommissioned from a node. The explorer does this when it receives a result with the state deleted from a node.

### Todo

- I did not yet measure the impact of those change regarding all the different currency supported. We should try to make sure to avoid the network split generated by the FreeTFT token. cf. https://github.com/threefoldtech/tfexplorer/issues/50
- The exact schema of the different new objects still needs to be defined in detail
- Define clearly all the modification in the code that would be needed for all component to implement this proposal

### Extra ideas

With this design, the capacity reservation is not linked to a single farm but to a list of nodes. This property would allow multiple farmer to create some kind of "virtual farm" where all farmer contribute some capacity and are rewarded based on the % of CPR they allocated to the virtual farm. Think crypto mining pool but for IT capacity. This is specially interesting for very small farmer that have less chance to get market shares. This needs to be further thought out and is not part of this reflection just yet.