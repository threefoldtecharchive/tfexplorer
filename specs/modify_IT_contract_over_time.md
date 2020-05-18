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
In order to implement this solution a possible approach would be to create some kind of capacity pool entity. User would reservation capacity from different farms. Then once the capacity is reserved, they could provision/decommission workloads on these farms freely without worrying when deleting a workloads cause then the capacity would just be given back to the pool and no token would be "lost".

If a user needs to extend the live of its workloads he just have to make sure the reserved capacity from its pools is always funded enough.

Some questions needs to be answered if we want to go this way:

- Q: A capacity pool apply on a farm or on a list of nodes ?
- A: Reservation on top a farm makes it pretty hard to do capacity planning on the node themself. Cause there is no way to know which nodes will be picked by the user. So there is a possibility a node will not be able to provide capacity even if the capacity pool still has some available capacity. For this reason I think capacity pool should actually be a list of node/capacity pair.

- How does the node actually reserve the capacity ?

- Should the node be involved in the creation of a pool ?
- what about over provisioning ? Can a farmer sell more capacity than he actually has. if yes what happens if the nodes cannot provides the capacity when requested by the client.

#### WIP design
<!-- 
```go

type ResourceAmount struct {
	Cru uint64 
	Mru float64
	Hru float64
	Sru float64
}

type SigningRequest struct {
	Signers   []int64
	QuorumMin int64
}

type SigningSignature struct {
	Tid       int64
	Signature string
	Created     time.Time
}

type CapacityPool struct {
    FarmID int64
    ResourceUnits ResourceAmount
    Expiration time.Time

    SigningRequestProvision SigningRequest
	SigningRequestDelete    SigningRequest

    SignaturesProvision []SigningSignature
	SignaturesFarmer    SigningSignature
	SignaturesDelete    []SigningSignature
}
``` -->