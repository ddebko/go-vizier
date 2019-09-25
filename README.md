# VIZIER
Vizier is a tool for concurrency abstraction using a state design architecture for Go (golang).

### TODO

- Look into logging, Run ~50x Slower When Tracing
- Look into avoiding channel buffer overflow & deadlocks
- Look into using Dynamic Select Statement reflect.SelectCase vs GoSelect
- Observer For The Worker Pool To Record Statistics On States/Edges
- Detect & Fail On Edge Cycle Within Manager
- Examples
    - DynamoDB