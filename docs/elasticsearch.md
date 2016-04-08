# Elasticsearch

With Inago, it's a piece of cake to deploy an Elasticsearch cluster,
as well as updating it without any downtime or data loss.

It should be noted that this is a demo, and there are dangers to using it in a production environment without further testing.
Saying that, if you want to do so, and have questions, the team is always available on IRC (`#giantswarm` on `freenode`) <3

## Setup
- This demo can be run on any standard fleet cluster.
For running it locally, we're using a CoreOS cluster running locally, with Vagrant.
See https://coreos.com/os/docs/latest/booting-on-vagrant.html for further information.
- There are three nodes in the cluster, and each node has 2048MB of memory allocated to it.
This large allocation is to give Elasticsearch more than enough space for heap.
- The `inago` binary, as well as the `elasticsearch` directory from `example`,
have been copied to all the nodes in the cluster.

It should also be noted that any IP addresses or hostnames used in this example are dependent on your setup.

## Inago Groups
Inago introduces the concept of a _group_. A group is a set of unit files.
Groups can be worked with in a similar manner to fleet units -
for example, they can be sliced, creating a group slice.

In this example, we will be using a group called `elasticsearch`,
which consists of one unit, named `elasticsearch@.service`.
We will use this group to create multiple group slices.
Each group slice of the `elasticsearch` group will contain a slice
of the `elasticsearch@.service` unit.

For simplicity in this example, there is only one unit in the group.
It would be possible to place more units in the `elasticsearch` group -
in this example, Logstash and Kibana would make sense,
and deploy them as part of a larger "ELK" group.

As an aside, a rule in Inago dictates that units containing an '@' symbol in their filename
can be sliced - all units in the group need to be sliceable for the group to be valid.

## Starting the Elasticsearch cluster
We're going to use Inago's `up` command to bring up multiple group slices of the `elasticsearch` group.
`up` is equivalent to `submit`, followed by `start`.

Due to the `Conflicts` statement in the unit file for the group,
each group slice will be automatically scheduled onto a separate machine.
```
core@core-01 ~ $ ./inagoctl up elasticsearch 3
2016-04-06 16:57:39.644 | INFO     | context.Background: Succeeded to submit group 'elasticsearch'.
2016-04-06 16:57:46.656 | INFO     | context.Background: Succeeded to start 3 slices for group 'elasticsearch': [1f4 8f1 9e9].
```
As you can see, the `elasticsearch` group has been submitted and started.
Each group slice has an ID, which can be seen in the output.

We can use the `status` command to inspect the state of the group.
The _verbose_ flag is used so that the hash of the unit file is also printed.
This will be interesting later, after we upgrade Elasticsearch.
You can also see that the group slices have been scheduled onto multiple machines.
```
core@core-01 ~ $ ./inagoctl status -v elasticsearch
Group              Units                      FDState   FCState   SAState  Hash                                      IP            Machine

elasticsearch@2cf  elasticsearch@2cf.service  launched  launched  active   07969bf8caa1a7f8c73b35bb826229425d3264f6  172.17.8.103  1564e0d3c4ab4aadbe76b24cea90c62f
elasticsearch@445  elasticsearch@445.service  launched  launched  active   07969bf8caa1a7f8c73b35bb826229425d3264f6  172.17.8.102  2dbbe125e237410bb946e7b0f6b95e4c
elasticsearch@891  elasticsearch@891.service  launched  launched  active   07969bf8caa1a7f8c73b35bb826229425d3264f6  172.17.8.101  b89277d82f6a425eb83abb39c00485bc
```

We can also check Elasticsearch via its own API.
```
core@core-02 ~ $ curl -XGET 'http://172.17.8.101:9200/_cluster/health?pretty=true'
{
  "cluster_name" : "elk",
  "status" : "green",
  "timed_out" : false,
  "number_of_nodes" : 3,
  "number_of_data_nodes" : 3,
  "active_primary_shards" : 0,
  "active_shards" : 0,
  "relocating_shards" : 0,
  "initializing_shards" : 0,
  "unassigned_shards" : 0,
  "delayed_unassigned_shards" : 0,
  "number_of_pending_tasks" : 0,
  "number_of_in_flight_fetch" : 0,
  "task_max_waiting_in_queue_millis" : 0,
  "active_shards_percent_as_number" : 100.0
}
```
The cluster status is `green`, and we have three nodes running. Hurrah!

We'll add an index, and a document to that index, too.
```
core@core-02 ~ $ curl -XPUT '172.17.8.102:9200/inago-example-test-index?pretty'
{
  "acknowledged" : true
}
core@core-02 ~ $ curl -XPUT 'http://172.17.8.103:9200/inago-example-test-index/external/1?pretty' -d '{"nanana": "batman!"}'
{
  "_index" : "inago-example-test-index",
  "_type" : "external",
  "_id" : "1",
  "_version" : 1,
  "_shards" : {
    "total" : 2,
    "successful" : 2,
    "failed" : 0
  },
  "created" : true
}
core@core-02 ~ $ curl -XGET '172.17.8.101:9200/inago-example-test-index/external/1?pretty'
{
  "_index" : "inago-example-test-index",
  "_type" : "external",
  "_id" : "1",
  "_version" : 1,
  "found" : true,
  "_source" : {
    "nanana" : "batman!"
  }
}
```

## Updating the Elasticsearch cluster
As mentioned previously, Inago operates on _groups_ of unit files.
If we look at the `elasticsearch@.service` file in the `elasticsearch` directory,
you can see a unit file of the group we're using in this example.

The Docker image for this unit is set via the line:
```
Environment="IMAGE=elasticsearch:2.2"
```

The currently running version of Elasticsearch can be verified via the Elasticsearch API.
```
core@core-02 ~ $ curl 172.17.8.101:9200
{
  "name" : "Crime-Buster",
  "cluster_name" : "inago-example",
  "version" : {
    "number" : "2.2.2",
    "build_hash" : "fcc01dd81f4de6b2852888450ce5a56436fd5852",
    "build_timestamp" : "2016-03-29T08:49:35Z",
    "build_snapshot" : false,
    "lucene_version" : "5.4.1"
  },
  "tagline" : "You Know, for Search"
}
```

However, this isn't the latest version of Elasticsearch! We want to upgrade it to 2.3.

We need to modify the line to read:
```
Environment="IMAGE=elasticsearch:2.3"
```
This updates the unit to use the later version of the Elasticsearch Docker image.

Next, we're going to use the `update` command from Inago to perform the update of the Elasticsearch cluster.
This will replace all the currently running group slices with new group slices.
```
core@core-01 ~ $ ./inagoctl update elasticsearch --max-growth=1 --min-alive=2 --ready-secs=60
2016-04-06 17:11:54.026 | INFO     | context.Background: Succeeded to update 3 slices for group 'elasticsearch': [184 372 88f].
```
The arguments used here mean that Inago is allowed to create one additional slice
during the update (meaning there will be no more than four instances of Elasticsearch running at any time),
and that at least two instances of Elasticsearch have to be running at any time.

The `ready-secs` flag determines how long to wait between rounds of updating.
This value gives Elasticsearch enough time to replicate any necessary data.
The amount of time required is dependent on your setup,
and is affected by the amount of data.

The IDs of the new group slices created during the update are printed in the output.

We can check the Elasticsearch cluster health via the Elasticsearch API.
```
core@core-02 ~ $ curl -XGET 'http://172.17.8.101:9200/_cluster/health?pretty=true'
{
  "cluster_name" : "inago-example",
  "status" : "green",
  "timed_out" : false,
  "number_of_nodes" : 3,
  "number_of_data_nodes" : 3,
  "active_primary_shards" : 5,
  "active_shards" : 10,
  "relocating_shards" : 0,
  "initializing_shards" : 0,
  "unassigned_shards" : 0,
  "delayed_unassigned_shards" : 0,
  "number_of_pending_tasks" : 0,
  "number_of_in_flight_fetch" : 0,
  "task_max_waiting_in_queue_millis" : 0,
  "active_shards_percent_as_number" : 100.0
}
```

If we had been watching this during the update, we would have seen the Elasticsearch
cluster status change between green and yellow, as nodes were added and removed.
The status would never go to red.

We can also check the current version of Elasticsearch running again, via Elasticsearch itself.
```
core@core-02 ~ $ curl 172.17.8.103:9200
{
  "name" : "Magilla",
  "cluster_name" : "inago-example",
  "version" : {
    "number" : "2.3.1",
    "build_hash" : "bd980929010aef404e7cb0843e61d0665269fc39",
    "build_timestamp" : "2016-04-04T12:25:05Z",
    "build_snapshot" : false,
    "lucene_version" : "5.5.0"
  },
  "tagline" : "You Know, for Search"
}
```

We can also check that we have not lost any data during the update, by fetching the document we created earlier.
```
core@core-02 ~ $ curl -XGET '172.17.8.101:9200/inago-example-test-index/external/1?pretty'
{
  "_index" : "inago-example-test-index",
  "_type" : "external",
  "_id" : "1",
  "_version" : 1,
  "found" : true,
  "_source" : {
    "nanana" : "batman!"
  }
}
```

By having a look at the verbose output of the `status` command, we can see that the hash of the groups has changed.
```
core@core-01 ~ $ ./inagoctl status -v elasticsearch
Group              Units                      FDState   FCState   SAState  Hash                                      IP            Machine

elasticsearch@184  elasticsearch@184.service  launched  launched  active   32e7209e49f5ef1c1116fa33c6481aafed8ea46b  172.17.8.103  1564e0d3c4ab4aadbe76b24cea90c62f
elasticsearch@372  elasticsearch@372.service  launched  launched  active   32e7209e49f5ef1c1116fa33c6481aafed8ea46b  172.17.8.101  b89277d82f6a425eb83abb39c00485bc
elasticsearch@88f  elasticsearch@88f.service  launched  launched  active   32e7209e49f5ef1c1116fa33c6481aafed8ea46b  172.17.8.102  2dbbe125e237410bb946e7b0f6b95e4c
```

That's all, folks!
We have managed to start up an Elasticsearch cluster,
write data to it, and perform a rolling update on it, all thanks to Inago.

The techniques used here are in no way specific to Elasticsearch.
Due to Inago building on top of fleet and systemd unit files,
all kinds of distributed systems can be orchestrated with Inago.