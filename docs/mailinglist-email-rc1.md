Dear fellow CoreOS users,

We would like to get some feedback on one of the tools we started developing last month: [Inago](https://github.com/giantswarm/inago/).

For some context: We are using fleet to schedule workloads onto our CoreOS clusters. One
of the patterns that emerged here, is what we call [unit chains](https://docs.giantswarm.io/fundamentals/user-services/container-injection/) - a group of units
that must run on the same machine and is chained with Requires/After.

We have developed two internal tools, which interact with fleet directly and operate on 
those chains. Since we are currently [open sourcing](https://giantswarm.io/products/) our tooling, we believe
a general tool to manage groups of units might be useful for others, too. Since our
two existing tools are too specialized, we started working on a more general approach.

With Inago you define a folder with unit files as a group and similar to fleetctl you can
submit, start, stop, and destroy slices of this group. This allows you to define
an app once, and start multiple copies of it for high availability.

Another feature we added is update, which is inspired by Kubernetes' rolling updates.
You provide a max-growth and min-alive parameter, and Inago dynamically creates new
slices and shuts down old ones, to replace all slices with the new version without downtimes.

Would be great to get your feedback and ideas on our v0.1 release candidate. You can find the code on [GitHub](https://github.com/giantswarm/inago/) and a prebuilt binary in the [releases](https://github.com/giantswarm/inago/releases). Feel free to play around
and give us feedback in either this thread, the [GitHub issues](https://github.com/giantswarm/inago/issues), or on IRC (#giantswarm).