---
title: Concurrency in GO - O'Reilly - Kathrine Cox Buday
date: Tuesday, October 21, 2025
keywords: [go, golang, books, programming]
---

## Concurrency in GO - Kathrine Cox Buday

### Deadlocks, Livelocks, Starvation

- Starvation is situation where on process cannot get resources it needs because
there is other process or other processes that are keeping to much of available
resources for themselfs resulting the one process (that starves) to be in a situation
where he spends most of his time waiting for resources (acquiring a lock or some
other resource)

### sync package

- sync.Cond

```go
func SyncCondExample() {
m := &sync.Mutex{}
cond := sync.Cond{L: m}

status := false

go func() {
  cond.L.Lock()

  for !status {
    fmt.Println("I will execute this once and go to sleep")
    cond.Wait()
  }
  cond.L.Unlock()
  }()

go func() {
  time.Sleep(1 * time.Second)

  cond.L.Lock()

  fmt.Println("Here we go")
  status = true

  cond.L.Unlock()

  //unlock only one waiting thread
  cond.Signal()

  //cond.Broadcast() -> unlock all threads waiting

}()
}
```

***Important note***

When using sync.Cond and when in first goroutine we call cond.Wait(), what this
does is it makes the thread go to sleep waiting for the signal and calls *L.Unlock*,
and then once it recieves a signal it cals *L.Lock* and that is why we have one
more *L.Unlock()* after for loop ends

- sync.Once
- sync.Pool

### SELECT

Select statement can read from a closed channel but cannot from nil channel, and
when it encounters a nil channel in select statement it ignores it which makes nil
channel useful for some algorithms where we have to stop reading from one path
in select statement at some point

### Confinement

Lexical confinement is using lexixal sxope to expose only the correct data and
concurrency primitives for multiple concurrent processes to use.

e.g.

1) Function that creates channel is responsible to close that channel
2) Function returns channel on which it plans to communicate results

### Prevent goroutine leaks

Goroutine should have few possible paths:

- Complete its work
- Stop due to error
- Be told to stop (Done channel)

Goroutine that is responsible for creating other goroutine should also be responsible
for ensuring it can stop it. Usually be using read only channel named "done" or
by using Context.

### Interesting channel implemetation

* or channel
* Fan out; Fan in
* or-done channel
* tee channel
* bridge channel -> for reading values from sequence of channels

### Queuing -> Littles Law -> L = lambda * W

### Context

Why is context immutable?
The idea is that child goroutine cannot change values of context which would then
cancel parent execution, it can only create new context from original to pass to
its children.

Storing data in context

```go
type foo int
type bar int

m := make(map[int]int)

m[foo(1)] = 1
m[bar(1)] = 2 // will not overwrite previous imput
```

This is why context keys should be used in following way:

```go
type ctxKey int

const (
  ctxUserID = iota
  ctxAuthToken
)

func UserID(c context.Context) string {
  return c.Value(ctxUserID).(string)
}

```

This way you cannot overwrite value in other package.

Be aware of existance of Timeouts and Cancellations anth the way they are part
of concurrent processes and Context.

### Heartbeats

Interesting concept that can be scaled to distributed systems and is useful in
using goroutines and making sure they are alive but I did not find use for them
so far.

### Rate limiting

Token bucket implementation

### Work stealing

***Stealing continuations*** - worht revisiting in future

Investigate follwoing to complete this page:

- Goroutine vs OS Thread vs A context (processor)
- race dection tools
- pprof

