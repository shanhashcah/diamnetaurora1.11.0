---
title: Aurora Administration Guide
replacement: https://developers.diamnet.org/docs/run-api-server/
---
## Aurora Administration Guide

Aurora is responsible for providing an HTTP API to data in the Diamnet network. It ingests and re-serves the data produced by the diamnet network in a form that is easier to consume than the performance-oriented data representations used by diamnet-core.

This document describes how to administer a **production** Aurora instance. If you are just starting with Aurora and want to try it out, consider the [Quickstart Guide](quickstart.md) instead. For information about developing on the Aurora codebase, check out the [Development Guide](developing.md).

## Why run Aurora?

The Diamnet Development Foundation runs two Aurora servers, one for the public network and one for the test network, free for anyone's use at https://aurora.diamnet.org and https://aurora-testnet.diamnet.org.  These servers should be fine for development and small scale projects, but it is not recommended that you use them for production services that need strong reliability.  By running Aurora within your own infrastructure provides a number of benefits:

  - Multiple instances can be run for redundancy and scalability.
  - Request rate limiting can be disabled.
  - Full operational control without dependency on the Diamnet Development Foundations operations.

## Prerequisites

Aurora is dependent upon a diamnet-core server.  Aurora needs access to both the SQL database and the HTTP API that is published by diamnet-core. See [the administration guide](https://www.diamnet.org/developers/diamnet-core/software/admin.html) to learn how to set up and administer a diamnet-core server.  Secondly, Aurora is dependent upon a postgres server, which it uses to store processed core data for ease of use. Aurora requires postgres version >= 9.5.

## Installing

To install Aurora, you have a choice: either downloading a [prebuilt release for your target architecture](https://github.com/diamnet/go/releases) and operation system, or [building Aurora yourself](#Building).  When either approach is complete, you will find yourself with a directory containing a file named `aurora`.  This file is a native binary.

After building or unpacking Aurora, you simply need to copy the native binary into a directory that is part of your PATH.  Most unix-like systems have `/usr/local/bin` in PATH by default, so unless you have a preference or know better, we recommend you copy the binary there.

To test the installation, simply run `aurora --help` from a terminal.  If the help for Aurora is displayed, your installation was successful. Note: some shells, such as zsh, cache PATH lookups.  You may need to clear your cache  (by using `rehash` in zsh, for example) before trying to run `aurora --help`.


## Building

Should you decide not to use one of our prebuilt releases, you may instead build Aurora from source.  To do so, you need to install some developer tools:

- A unix-like operating system with the common core commands (cp, tar, mkdir, bash, etc.)
- A compatible distribution of Go (Go 1.14 or later)
- [git](https://git-scm.com/)
- [mercurial](https://www.mercurial-scm.org/)

1. See the details in [README.md](../../../../README.md#dependencies) for installing dependencies.
2. Compile the Aurora binary: `go install github.com/diamnet/go/services/aurora`. You should see the `aurora` binary in `$GOPATH/bin`.
3. Add Go binaries to your PATH in your `bashrc` or equivalent, for easy access: `export PATH=${GOPATH//://bin:}/bin:$PATH`

Open a new terminal. Confirm everything worked by running `aurora --help` successfully.

Note:  Building directly on windows is not supported.


## Configuring

Aurora is configured using command line flags or environment variables.  To see the list of command line flags that are available (and their default values) for your version of Aurora, run:

`aurora --help`

As you will see if you run the command above, Aurora defines a large number of flags, however only three are required:

| flag                    | envvar                      | example                              |
|-------------------------|-----------------------------|--------------------------------------|
| `--db-url`              | `DATABASE_URL`              | postgres://localhost/aurora_testnet |
| `--diamnet-core-db-url` | `DIAMNET_CORE_DATABASE_URL` | postgres://localhost/core_testnet    |
| `--diamnet-core-url`    | `DIAMNET_CORE_URL`          | http://localhost:11626               |

`--db-url` specifies the Aurora database, and its value should be a valid [PostgreSQL Connection URI](http://www.postgresql.org/docs/9.2/static/libpq-connect.html#AEN38419).  `--diamnet-core-db-url` specifies a diamnet-core database which will be used to load data about the diamnet ledger.  Finally, `--diamnet-core-url` specifies the HTTP control port for an instance of diamnet-core.  This URL should be associated with the diamnet-core that is writing to the database at `--diamnet-core-db-url`.

Specifying command line flags every time you invoke Aurora can be cumbersome, and so we recommend using environment variables.  There are many tools you can use to manage environment variables:  we recommend either [direnv](http://direnv.net/) or [dotenv](https://github.com/bkeepers/dotenv).  A template configuration that is compatible with dotenv can be found in the [Aurora git repo](https://github.com/diamnet/go/blob/master/services/aurora/.env.template).



## Preparing the database

Before the Aurora server can be run, we must first prepare the Aurora database.  This database will be used for all of the information produced by Aurora, notably historical information about successful transactions that have occurred on the diamnet network.

To prepare a database for Aurora's use, first you must ensure the database is blank.  It's easiest to simply create a new database on your postgres server specifically for Aurora's use.  Next you must install the schema by running `aurora db init`.  Remember to use the appropriate command line flags or environment variables to configure Aurora as explained in [Configuring ](#Configuring).  This command will log any errors that occur.

### Postgres configuration

It is recommended to set `random_page_cost=1` in Postgres configuration if you are using SSD storage. With this setting Query Planner will make a better use of indexes, especially for `JOIN` queries. We have noticed a huge speed improvement for some queries.

## Running

Once your Aurora database is configured, you're ready to run Aurora.  To run Aurora you simply run `aurora` or `aurora serve`, both of which start the HTTP server and start logging to standard out.  When run, you should see some output that similar to:

```
INFO[0000] Starting aurora on :8000                     pid=29013
```

The log line above announces that Aurora is ready to serve client requests. Note: the numbers shown above may be different for your installation.  Next we can confirm that Aurora is responding correctly by loading the root resource.  In the example above, that URL would be [http://127.0.0.1:8000/] and simply running `curl http://127.0.0.1:8000/` shows you that the root resource can be loaded correctly.

If you didn't set up a diamnet-core yet, you may see an error like this:
```
ERRO[2019-05-06T16:21:14.126+08:00] Error getting core latest ledger err="get failed: pq: relation \"ledgerheaders\" does not exist"
```
Aurora requires a functional diamnet-core. Go back and set up diamnet-core as described in the admin guide. In particular, you need to initialise the database as [described here](https://www.diamnet.org/developers/diamnet-core/software/admin.html#database-and-local-state).

## Ingesting live diamnet-core data

Aurora provides most of its utility through ingested data.  Your Aurora server can be configured to listen for and ingest transaction results from the connected diamnet-core.

To enable ingestion, you must either pass `--ingest=true` on the command line or set the `INGEST`
environment variable to "true". Since version 1.0.0 you can start multiple ingesting machines in your cluster.

### Ingesting historical data and reingesting Ledgers

To reingest older ledgers (due to a version upgrade) or to ingest ledgers closed by the network before you
started Aurora. This is done through the `aurora db range [START_LEDGER] [END_LEDGER]` command, which could
be run as follows:

```
aurora1> aurora db reingest range 1 10000
aurora2> aurora db reingest range 10001 20000
aurora3> aurora db reingest range 20001 30000
# ... etc.
```

This allows reingestion to be split up and done in parallel by multiple Aurora processes.

### Managing storage for historical data

Over time, the recorded network history will grow unbounded, increasing storage used by the database. Aurora expands the data ingested from diamnet-core and needs sufficient disk space. Unless you need to maintain a history archive you may configure Aurora to only retain a certain number of ledgers in the database. This is done using the `--history-retention-count` flag or the `HISTORY_RETENTION_COUNT` environment variable. Set the value to the number of recent ledgers you wish to keep around, and every hour the Aurora subsystem will reap expired data.  Alternatively, you may execute the command `aurora db reap` to force a collection.

### Surviving diamnet-core downtime

Aurora tries to maintain a gap-free window into the history of the diamnet-network.  This reduces the number of edge cases that Aurora-dependent software must deal with, aiming to make the integration process simpler.  To maintain a gap-free history, Aurora needs access to all of the metadata produced by diamnet-core in the process of closing a ledger, and there are instances when this metadata can be lost.  Usually, this loss of metadata occurs because the diamnet-core node went offline and performed a catchup operation when restarted.

To ensure that the metadata required by Aurora is maintained, you have several options: You may either set the `CATCHUP_COMPLETE` diamnet-core configuration option to `true` or configure `CATCHUP_RECENT` to determine the amount of time your diamnet-core can be offline without having to rebuild your Aurora database.

Unless your node is a full validator and archive publisher we _do not_ recommend using the `CATCHUP_COMPLETE` method, as this will force diamnet-core to apply every transaction from the beginning of the ledger, which will take an ever increasing amount of time. Instead, we recommend you set the `CATCHUP_RECENT` config value. To do this, determine how long of a downtime you would like to survive (expressed in seconds) and divide by ten.  This roughly equates to the number of ledgers that occur within your desired grace period (ledgers roughly close at a rate of one every ten seconds).  With this value set, diamnet-core will replay transactions for ledgers that are recent enough, ensuring that the metadata needed by Aurora is present.

### Correcting gaps in historical data

In the section above, we mentioned that Aurora _tries_ to maintain a gap-free window.  Unfortunately, it cannot directly control the state of diamnet-core and [so gaps may form](https://www.diamnet.org/developers/software/known-issues.html#gaps-detected) due to extended down time.  When a gap is encountered, Aurora will stop ingesting historical data and complain loudly in the log with error messages (log lines will include "ledger gap detected").  To resolve this situation, you must re-establish the expected state of the diamnet-core database and purge historical data from Aurora's database.  We leave the details of this process up to the reader as it is dependent upon your operating needs and configuration, but we offer one potential solution:

We recommend you configure the HISTORY_RETENTION_COUNT in Aurora to a value less than or equal to the configured value for CATCHUP_RECENT in diamnet-core.  Given this situation any downtime that would cause a ledger gap will require a downtime greater than the amount of historical data retained by Aurora.  To re-establish continuity:

1.  Stop Aurora.
2.  Run `aurora db reap` to clear the historical database.
3.  Clear the cursor for Aurora by running `diamnet-core -c "dropcursor?id=HORIZON"` (ensure capitilization is maintained).
4.  Clear ledger metadata from before the gap by running `diamnet-core -c "maintenance?queue=true"`.
5.  Restart Aurora.

### Some endpoints are not available during state ingestion

Endpoints that display state information are not available during initial state ingestion and will return a `503 Service Unavailable`/`Still Ingesting` error.  An example is the `/paths` endpoint (built using offers). Such endpoints will become available after state ingestion is done (usually within a couple of minutes).

### State ingestion is taking a lot of time

State ingestion shouldn't take more than a couple of minutes on an AWS `c5.xlarge` instance, or equivalent.

It's possible that the progress logs (see below) will not show anything new for a longer period of time or print a lot of progress entries every few seconds. This happens because of the way history archives are designed. The ingestion is still working but it's processing entries of type `DEADENTRY`'. If there is a lot of them in the bucket, there are no _active_ entries to process. We plan to improve the progress logs to display actual percentage progress so it's easier to estimate ETA.

If you see that ingestion is not proceeding for a very long period of time:
1. Check the RAM usage on the machine. It's possible that system run out of RAM and it using swap memory that is extremely slow.
2. If above is not the case, file a new issue in this repository.

### CPU usage goes high every few minutes

This is _by design_. Aurora runs a state verifier routine that compares state in local storage to history archives every 64 ledgers to ensure data changes are applied correctly. If data corruption is detected Aurora will block access to endpoints serving invalid data.

We recommend to keep this security feature turned on however if it's causing problems (due to CPU usage) this can be disabled by `--ingest-disable-state-verification` CLI param or `INGEST-DISABLE-STATE-VERIFICATION` env variable.

### I see `Waiting for the next checkpoint...` messages

If you were running the new system in the past during experimental stage (`ENABLE_EXPERIMENTAL_INGESTION` flag) it's possible that the old and new systems are not in sync. In such case, the upgrade code will activate and will make sure the data is in sync. When this happens you may see `Waiting for the next checkpoint...` messages for up to 5 minutes.

## Reading the logs

In order to check the progress and the status of experimental ingestion you should check the logs. All logs connected to experimental ingestion are tagged with `service=ingest`.

It starts with informing you about state ingestion:
```
INFO[2019-08-29T13:04:13.473+02:00] Starting ingestion system from empty state...  pid=5965 service=ingest temp_set="*io.MemoryTempSet"
INFO[2019-08-29T13:04:15.263+02:00] Reading from History Archive Snapshot         ledger=25565887 pid=5965 service=ingest
```
During state ingestion, Aurora will log number of processed entries every 100,000 entries (there are currently around 7M entries in the public network):
```
INFO[2019-08-29T13:04:34.652+02:00] Processing entries from History Archive Snapshot  ledger=25565887 numEntries=100000 pid=5965 service=ingest
INFO[2019-08-29T13:04:38.487+02:00] Processing entries from History Archive Snapshot  ledger=25565887 numEntries=200000 pid=5965 service=ingest
INFO[2019-08-29T13:04:41.322+02:00] Processing entries from History Archive Snapshot  ledger=25565887 numEntries=300000 pid=5965 service=ingest
INFO[2019-08-29T13:04:48.429+02:00] Processing entries from History Archive Snapshot  ledger=25565887 numEntries=400000 pid=5965 service=ingest
INFO[2019-08-29T13:05:00.306+02:00] Processing entries from History Archive Snapshot  ledger=25565887 numEntries=500000 pid=5965 service=ingest
```
When state ingestion is finished it will proceed to ledger ingestion starting from the next ledger after checkpoint ledger (25565887+1 in this example) to update the state using transaction meta:
```
INFO[2019-08-29T13:39:41.590+02:00] Processing entries from History Archive Snapshot  ledger=25565887 numEntries=5300000 pid=5965 service=ingest
INFO[2019-08-29T13:39:44.518+02:00] Processing entries from History Archive Snapshot  ledger=25565887 numEntries=5400000 pid=5965 service=ingest
INFO[2019-08-29T13:39:47.488+02:00] Processing entries from History Archive Snapshot  ledger=25565887 numEntries=5500000 pid=5965 service=ingest
INFO[2019-08-29T13:40:00.670+02:00] Processed ledger                              ledger=25565887 pid=5965 service=ingest type=state_pipeline
INFO[2019-08-29T13:40:00.670+02:00] Finished processing History Archive Snapshot  duration=2145.337575904 ledger=25565887 numEntries=5529931 pid=5965 service=ingest shutdown=false
INFO[2019-08-29T13:40:00.693+02:00] Reading new ledger                            ledger=25565888 pid=5965 service=ingest
INFO[2019-08-29T13:40:00.694+02:00] Processing ledger                             ledger=25565888 pid=5965 service=ingest type=ledger_pipeline updating_database=true
INFO[2019-08-29T13:40:00.779+02:00] Processed ledger                              ledger=25565888 pid=5965 service=ingest type=ledger_pipeline
INFO[2019-08-29T13:40:00.779+02:00] Finished processing ledger                    duration=0.086024492 ledger=25565888 pid=5965 service=ingest shutdown=false transactions=14
INFO[2019-08-29T13:40:00.815+02:00] Reading new ledger                            ledger=25565889 pid=5965 service=ingest
INFO[2019-08-29T13:40:00.816+02:00] Processing ledger                             ledger=25565889 pid=5965 service=ingest type=ledger_pipeline updating_database=true
INFO[2019-08-29T13:40:00.881+02:00] Processed ledger                              ledger=25565889 pid=5965 service=ingest type=ledger_pipeline
INFO[2019-08-29T13:40:00.881+02:00] Finished processing ledger                    duration=0.06619956 ledger=25565889 pid=5965 service=ingest shutdown=false transactions=29
INFO[2019-08-29T13:40:00.901+02:00] Reading new ledger                            ledger=25565890 pid=5965 service=ingest
INFO[2019-08-29T13:40:00.902+02:00] Processing ledger                             ledger=25565890 pid=5965 service=ingest type=ledger_pipeline updating_database=true
INFO[2019-08-29T13:40:00.972+02:00] Processed ledger                              ledger=25565890 pid=5965 service=ingest type=ledger_pipeline
INFO[2019-08-29T13:40:00.972+02:00] Finished processing ledger                    duration=0.071039012 ledger=25565890 pid=5965 service=ingest shutdown=false transactions=20
```


## Managing Stale Historical Data

Aurora ingests ledger data from a connected instance of diamnet-core.  In the event that diamnet-core stops running (or if Aurora stops ingesting data for any other reason), the view provided by Aurora will start to lag behind reality.  For simpler applications, this may be fine, but in many cases this lag is unacceptable and the application should not continue operating until the lag is resolved.

To help applications that cannot tolerate lag, Aurora provides a configurable "staleness" threshold.  Given that enough lag has accumulated to surpass this threshold (expressed in number of ledgers), Aurora will only respond with an error: [`stale_history`](./reference/errors/stale-history.md).  To configure this option, use either the `--history-stale-threshold` command line flag or the `HISTORY_STALE_THRESHOLD` environment variable.  NOTE:  non-historical requests (such as submitting transactions or finding payment paths) will not error out when the staleness threshold is surpassed.

## Monitoring

To ensure that your instance of Aurora is performing correctly we encourage you to monitor it, and provide both logs and metrics to do so.

Aurora will output logs to standard out.  Information about what requests are coming in will be reported, but more importantly, warnings or errors will also be emitted by default.  A correctly running Aurora instance will not output any warning or error log entries.

Metrics are collected while a Aurora process is running and they are exposed at the `/metrics` path.  You can see an example at (https://aurora-testnet.diamnet.org/metrics).

Below we present a few standard log entries with associated fields. You can use them to build metrics and alerts. We present below some examples. Please note that this represents Aurora app metrics only. You should also monitor your hardware metrics like CPU or RAM Utilization.

### Starting HTTP request

| Key              | Value                                                                                          |
|------------------|------------------------------------------------------------------------------------------------|
| **`msg`**        | **`Starting request`**                                                                         |
| `client_name`    | Value of `X-Client-Name` HTTP header representing client name                                  |
| `client_version` | Value of `X-Client-Version` HTTP header representing client version                            |
| `app_name`       | Value of `X-App-Name` HTTP header representing app name                                        |
| `app_version`    | Value of `X-App-Version` HTTP header representing app version                                  |
| `forwarded_ip`   | First value of `X-Forwarded-For` header                                                        |
| `host`           | Value of `Host` header                                                                         |
| `ip`             | IP of a client sending HTTP request                                                            |
| `ip_port`        | IP and port of a client sending HTTP request                                                   |
| `method`         | HTTP method (`GET`, `POST`, ...)                                                               |
| `path`           | Full request path, including query string (ex. `/transactions?order=desc`)                     |
| `streaming`      | Boolean, `true` if request is a streaming request                                              |
| `referer`        | Value of `Referer` header                                                                      |
| `req`            | Random value that uniquely identifies a request, attached to all logs within this HTTP request |

### Finished HTTP request

| Key              | Value                                                                                          |
|------------------|------------------------------------------------------------------------------------------------|
| **`msg`**        | **`Finished request`**                                                                         |
| `bytes`          | Number of response bytes sent                                                                  |
| `client_name`    | Value of `X-Client-Name` HTTP header representing client name                                  |
| `client_version` | Value of `X-Client-Version` HTTP header representing client version                            |
| `app_name`       | Value of `X-App-Name` HTTP header representing app name                                        |
| `app_version`    | Value of `X-App-Version` HTTP header representing app version                                  |
| `duration`       | Duration of request in seconds                                                                 |
| `forwarded_ip`   | First value of `X-Forwarded-For` header                                                        |
| `host`           | Value of `Host` header                                                                         |
| `ip`             | IP of a client sending HTTP request                                                            |
| `ip_port`        | IP and port of a client sending HTTP request                                                   |
| `method`         | HTTP method (`GET`, `POST`, ...)                                                               |
| `path`           | Full request path, including query string (ex. `/transactions?order=desc`)                     |
| `route`          | Route pattern without query string (ex. `/accounts/{id}`)                                      |
| `status`         | HTTP status code (ex. `200`)                                                                   |
| `streaming`      | Boolean, `true` if request is a streaming request                                              |
| `referer`        | Value of `Referer` header                                                                      |
| `req`            | Random value that uniquely identifies a request, attached to all logs within this HTTP request |

### Metrics

Using the entries above you can build metrics that will help understand performance of a given Aurora node, some examples below:
* Number of requests per minute.
* Number of requests per route (the most popular routes).
* Average response time per route.
* Maximum response time for non-streaming requests.
* Number of streaming vs. non-streaming requests.
* Number of rate-limited requests.
* List of rate-limited IPs.
* Unique IPs.
* The most popular SDKs/apps sending requests to a given Aurora node.
* Average ingestion time of a ledger.
* Average ingestion time of a transaction.

### Alerts

Below we present example alerts with potential cause and solution. Feel free to add more alerts using your metrics.

Alert | Cause | Solution
-|-|-
Spike in number of requests | Potential DoS attack | Lower rate-limiting threshold
Large number of rate-limited requests | Rate-limiting threshold too low | Increase rate-limiting threshold
Ingestion is slow | Aurora server spec too low | Increase hardware spec
Spike in average response time of a single route | Possible bug in a code responsible for rendering a route | Report an issue in Aurora repository.

## I'm Stuck! Help!

If any of the above steps don't work or you are otherwise prevented from correctly setting up
Aurora, please come to our community and tell us. Either
[post a question at our Stack Exchange](https://diamnet.stackexchange.com/) or
[chat with us on Keybase in #dev_discussion](https://keybase.io/team/diamnet.public) to ask for
help.
