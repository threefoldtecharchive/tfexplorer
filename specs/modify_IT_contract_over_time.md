# The problem: update IT contract over time

At the moment, there is no way for a grid user to modify anything in the workloads that are live on the grid. The IT contract is something that is agreed by the user and all involved farmers before hands.

Once a IT contract is signed by all parties, it is then provisioned by the 3Nodes and the workloads comes to life.

If the reservation has been made for a duration of 6 month. There is no way for the user to cancel it after 3 month and not loose the tokens for the last 3months. A manual discussion between the user and all the farmers can maybe be initiated but there is no grantee the farmer will accept to refund the remaining 3 months and this process is impossible to scale.

## Possible solutions

### Solution 1: Allow modification of the IT contract over time

One possible approach to solve this would be to allow to "update/modify" an IT contract over time.
Depending which field of the reservation is modified different things would need to happen:

- [No change in overall resource unit account of the reservation](#No-change-in-overall-resource-unit-account-of-the-reservation)
- [Change in overall resource unit account of the reservation](#Change-in-overall-resource-unit-account-of-the-reservation)
- [Change of the reservation expiration timestamp](#Change-of-the-reservation-expiration-timestamp)

It should be possible to follow the evolution of a single reservation over time. Which means we need to be smart about the data structure we use and how we include updates inside.

A simple approach would just be to update an existing reservation object, then send it to a new API endpoint of the explorer, `reservation_update`, this endpoint would then just validate that the reservation actually already exists, then redo the same kind of validation of all the fields, eventual payment flow. This would eventually lead to new workloads being sent to the nodes.

While this seems easy, it prevents any tracking of update over time cause we just simply overwrite the content of the reservation. So this solution not valid.

Another idea could be to define new object that can be attached to a reservation. This would create some kind of chain that represent the change of a reservation over time. This idea is fine if we only consider expiration extension. But for actual workloads definition changes, this is a bit hard to put in place. That would requires to build the chain in memory in the explorer anytime someone want to get a reservation or a node ask for its workloads. I don't think this can be scaled properly.

Third idea would be to modify the reservation object we currently use to add some field or convert some into a list. Each item into the list would have a timestamp attached to it. That would allow to follow the evolution of a reservation over time. This seems the be the better approach but this also means we completely break any compatibility with what we have today.

#### No change in overall resource unit account of the reservation

Only the user(s) need(s) to sign in this case, as from the farmer's point of view, nothing changes. However, it can be that the workload needs to be countersigned by other users. The same quorum applies as with original provisioning of the workload.

Once agreed the 3Node needs to do the necessary change to make the workloads update. Usually this means a `stop` / `start`  of the workloads. Although depending on the workloads type and the modified field this can be different. Note that it can happen that some workloads would not be possible to update.

#### Change in overall resource unit account of the reservation

This change is often combined with the previous one: if a workload requires more capacity, then this workload needs to be reserved additionally, and then this capacity needs to be addressed in the workload.  
A new round of payment needs to happen to cover the new total amount of resource reserved.

#### Change of the reservation expiration timestamp

The user wants to extend his reservation without changing anything in the workloads definition.  
Requires new round of payment needs to happen to cover the new total amount of resource reserved.  

This raises the question of how does potential discount on longer reservation would apply.

Same rules apply as with provisioning, like the `quorum_min` must be respected to indicate minimum amount of signatures to be respected. For this type of change, the reservation.data object doesn't change.

Proposal is not to allow shortening the reservation time by the user, as this can only be initiated by the sender of the tokens, in this case the farmer. The capacity user does not have the keys to trigger this action.

The workflow for this change looks like : ![extend_reservation_flow](workflow_extend_reservation.png)

##### Updatable field per workloads types

Here we list of the field that could be updated at runtime for each workloads types. Any field not present in this list is considered not "updatable" and any attempt to change it would result in a BadRequest error from the explorer.

###### 0-OS workloads

volume:

- size: can be increased but never reduced

zdb:

- size: can be increased but never reduced
- password: can be changed
- public: can be changed

container:

any change in a container definition will involved a stop/start of the container.

- Flist: can be changed
- HubUrl: can be changed
- Environment: can be changed
- SecretEnvironment: can be changed
- Entrypoint: can be changed
- Interactive: can be changed
- Volumes: volumes can be added, but not removed
- NetworkConnection: can be changed
- StatsAggregator: can be changed
- Logs: can be changed
- Capacity: can be changed

K8S:

At this point nothing can be updated at runtime for a k8s VM. If someone needs more resource it must add more workers.

WebGateway workloads:

None of the workloads provided by the webgateway can be updated.

### Solution 2: Decouple the reservation of capacity from the actual workloads description

This solution splits the reservation of capacity from the workloads consuming this capacity.
In order to implement this solution a possible approach would be to create some kind of capacity pool entity. User would reserve capacity from different farms. Then once the capacity is reserved, they could provision/decommission workloads on these farms freely without worrying when deleting a workloads cause then the capacity would just be given back to the pool and no token would be "lost".

If a user needs to extend the live of its workloads he just have to make sure the reserved capacity from its pools is always funded enough.

Here are the 2 new flows that would need to be implemented:

#### Creation of the capacity pool

![reserve_capacity](reserve_capacity.png)

This flow is very similar to the current flow to deploy workloads. The main difference is only the amount of resource units if defined instead of a full workloads definition.

Both the farmer and the user still needs to sign the reservation to mark their agreement on the deal. Multi signature is still possible using the same signing request construct that exists today.

The main difference is during the call numbered 3 in the diagram. When the user send the capacity reservation to the explorer, the explorer will block. During this time, the explorer will sends the capacity reservation to the nodes and wait for them to answer (in reality the node poll the explorer but this is an implementation detail, the idea is that the node is made aware of capacity reservation).
There are 2 possibility, or the node managed to lock the amount of capacity asked, or not. In both cases the node answer to the explorer.
If all nodes managed to lock the capacity, the explorer will mark the capacity reservation as to be paid and send the payment detail to the user.  
If any of the node failed to lock the capacity, the explorer will return an error to the user. The error will give detail about which node failed to lock capacity. The user can then modify it's request and retry.

This flow is made in such a way to avoid having to refund any token back to the client. When the node received the request to lock capacity, it will create an capacity pool object kept on cache the define the amount of capacity reserved by the pool. This object will be used later on during the life of the node to decide if a workload can be deployed or not.  
The pool object also contains the expiration date of the pool. This will also be used by the node over time to periodically check back in the explorer and make sure the pool has been extended. If a node has some workloads that are part of a pool that is about to expire, a call is make to the explorer to ask what is the expiration date of the pool. If the pool has not been extended, all the workloads linked to this pool will be decommissioned.

With this system in place, expiration are not needed on the workload definition anymore. It also greatly reduce the amount of time the token are locked in the explorer, cause now as soon as they token are received on the escrow account, they can be forwarded directly to the farmer.

If for some reason the client fails to pay the capacity reservation in time. Only the capacity needs to be unlock on the node but there are no token transfer involved at all (unless the client pay only a part of the amount, then refund needs to happens)

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
