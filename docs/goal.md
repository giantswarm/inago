# What is the goal of Inago?
Inago is a deployment tool that manages groups of unit files to deploy them to
a fleet cluster similar to `fleetctl`. Since `fleetctl` is quite limited, Inago
aims to abstract units away and provide more sugar on top like update
strategies. That way the user can manage unit files more easily. All actions one
can apply to a unit using `fleetctl`, can also be applied towards a group using
`inagoctl`.
