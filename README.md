# VIZIER
Vizier is a tool for concurrency abstraction using a state design architecture for Go (golang).

![thread-drop](./img/gopher-thread-drop.png "thread drop gopher")

image credit: edited version of [ashleymcnamara's mic drop gopher](https://twitter.com/ashleymcnamara/status/860462810819702784) + [thread clipart](http://clipart-library.com/clipart/8TzrpBdRc.htm)

### TODO

- Invoke States Inline
- Pin GoRoutines To States
- Create [Benchmark Comparison](https://benchmarksgame-team.pages.debian.net/benchmarksgame/fastest/go-node.html)
- Look into using Dynamic Select Statement reflect.SelectCase vs GoSelect
- Observer For The Worker Pool To Record Statistics On States/Edges
- Detect & Fail On Edge Cycle Within Manager
- Unit Test Coverage
- Examples
    - DynamoDB
  -
