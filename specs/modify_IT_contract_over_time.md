# The problem: update IT contract over time

At the moment, there is no way for a grid user to modify anything in the workloads that are live on the grid. The IT contract is something that is agreed by the user and all involved farmers before hands.

Once a IT contract is signed by all parties, it is then provisioned by the 3Nodes and the workloads comes to life.

If the reservation has been made for a duration of 6 month. There is no way for the user to cancel it after 3 month and not loose the tokens for the last 3months. A manual discussion between the user and all the farmers can maybe be initiated but there is no grantee the farmer will accept to refund the remaining 3 months and this process is impossible to scale.

## Possible solutions

### Allow modification of the IT contract over time

One possible approach to solve this would be to allow to "update/modify" an IT contract over time.
Depending which field of the reservation is modified different things would need to happen:

### No change in overall resource unit account of the reservation

This can happen if a bug or a change needs to happen in the flist.  
Only the user(s) need(s) to sign in this case, as from the farmer's point of view, nothing changes. However, it can be that the workload needs to be countersigned by other users. The same quorum applies as with original provisioning of the workload.

Once agreed the 3Node needs to do the necessary change to make the workloads update. Usually this means a `stop` / `start`  of the workloads. Although depending on the workloads type and the modified field this can be different. Note that it can happen that some workloads would not be possible to update.

### Change in overall resource unit account of the reservation

This change is often combined with the previous one: if a workload requires more capacity, then this workload needs to be reserved additionally, and then this capacity needs to be addressed in the workload.  
A new round of payment needs to happen to cover the new total amount of resource reserved.

### Change of the reservation expiration timestamp

The user wants to extend his reservation without changing anything in the workloads definition.  
Requires new round of payment needs to happen to cover the new total amount of resource reserved. 
To achieve this, a `signing_request_extend` field needs to be filled with valid signatures.  

This raises the question of how does potential discount on longer reservation would apply.

After signing `signature_extend`is filled with a valid signature. Same rules apply as with provisioning, like the `quorum_min` must be respected to indicate minimum amount of signatures to be respected. For this type of change, the reservation.data object doesn't change.

Proposal is not to allow shortening the reservation time by the user, as this can only be initiated by the sender of the tokens, in this case the farmer. The capacity user does not have the keys to trigger this action.

The workflow for this change looks like : ![extend_reservation_flow](workflow_extend_reservation.png)

### Decouple the reservation of capacity from the actual workloads description

This solution splits the reservation of capacity from the workloads consuming this capacity.
In order to implement this solution a possible approach would be to create some kind of capacity pool entity. User would reservation capacity from different farms. Then once the capacity is reserved, they could provision/decommission workloads on these farms freely without worrying when deleting a workloads cause then the capacity would just be given back to the pool and no token would be "lost".

If a user needs to extend the live of its workloads he just have to make sure the reserved capacity from its pools is always funded enough.

While this solution seems to solve of the problems is also requires nearly a complete rewrite of how capacity is reserved on the grid.
How a reservation will be linked to a capacity pool is not yet defined in this spec.
