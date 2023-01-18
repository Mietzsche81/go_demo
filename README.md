# go_demo
Demo of the Go programming language

## Concurrency
A quick and dirty implementation of multithreading. 
- Creates a series of `Job`s that will pretend to "run" by sleeping for a random amount of time between 1-10s
- Distribute theses jobs to `n` batches.
- All batches execute simultaneously, with each batch executing their segment of `Job`s in series.
- The batches use `chan` comms to signal when a job starts, completes, or fails.
- A `Monitor` checks these `chan`s through blocking `select` statements to get signals as they arrive.
- The `Monitor` logs these signals and periodically prints a status report.
