Dear fellow CoreOS Users,

We would like to get some feedback on one of the tools we started developing last month: Inago.

At Giant Swarm we are using fleet to schedule our workload onto our cluster. One
of the patterns that emerged here, is what we call unit-chains. A group of units
that must run on the same machine and is chained with Requires/After.

We have developed two internal tools, which interact with fleet directly and operate on 
those chains. Since we are currently [open sourcing](1) a bunch of tools, we believe
a general tool to manage those unit-chains might be useful for others too. Since our
two existing tools are too specialized, we started working on a more general approach:

With Inago you define a folder with unit files and similar to fleetctl you can
submit, start, stop, destroy slices of this folder. This allows you to define
an app once, and start multiple copies for high availability.

Another feature we added is update. Update is inspired by Kubernetes Rolling Update.
You provide a max-growth and min-alive parameter, and Inago dynamically creates new
slices and shutsdown the old ones, to replace all slices with the new version.

You can find the code on GitHub[3] and a prebuild binary at [2]. Feel free to play around
and give us feedback in the github issues.

1: https://giantswarm.io/products/
