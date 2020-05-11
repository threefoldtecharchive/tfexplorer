## The problem

At the moment, there is no way for a grid user to modify anything in the workloads that are live on the grid. The IT contract is something that is agreed by the user and all involved farmers before hands.

Once a IT contract is signed by all parties, it is then provisioned by the 3Nodes and the workloads comes to life.

If the reservation has been made for a duration of 6 month. There is no way for the user to cancel it after 3 month and not loose the tokens for the last 3months. A manual discussion between the user and all the farmers can maybe be initiated but there is no grantee the farmer will accept to refund the remaining 3 months and this process is impossible to scale.

## Possible solutions

### Allow modification of the IT contract over time

One possible approach to solve this would be to allow to "update/modify" an IT contract over time.
Depending which field of the reservation is modified different things would need to happen:

- **change of workloads definition that doesn't change the overall resource unit count of the reservation**: The farmer and user just needs to sign the new workloads definition. Once agreed the 3Node needs to do the necessary change to make the workloads update. Usually this means a `stop` / `start` of the workloads. Although depending on the workloads type and the modified field this can be different. note that it can happen that some workloads would not be possible to update.
- **change of workloads definition that  does change the overall resource unit count of the reservation**: Same as the previous one + a new round of payment needs to happen to cover the new total amount of resource reserved.
- **change of the reservation expiration**: The user wants to extend his reservation without changing anything in the workloads definition. Requires new round of payment needs to happen to cover the new total amount of resource reserved. This raises the question of how does potential discount on longer reservation would apply.

To implement this solution we need to go over all the field from each possible workloads definition and decide if it is possible to modify it or not.

### Decouple the reservation of capacity from the actual workloads description

This solution splits the reservation of capacity from the workloads consuming this capacity.
In order to implement this solution a possible approach would be to create some kind of capacity pool entity. User would reservation capacity from different farms. Then once the capacity is reserved, they could provision/decommission workloads on these farms freely without worrying when deleting a workloads cause then the capacity would just be given back to the pool and no token would be "lost".

If a user needs to extend the live of its workloads he just have to make sure the reserved capacity from its pools is always funded enough.

While this solution seems to solve of the problems is also requires nearly a complete rewrite of how capacity is reserved on the grid.
How a reservation will be linked to a capacity pool is not yet defined in this spec.