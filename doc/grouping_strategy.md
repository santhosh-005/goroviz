# 🔍 Goroutine Grouping Strategy

This document explains the core grouping strategy used by Goroviz to cluster thousands of concurrent goroutines into a clean, human-readable terminal dashboard.

---

## 🗺️ The Goal of Grouping
When running a high-traffic Go application, there might be **10,000+ active goroutines**. Printing all of them would scroll endlessly and make debugging impossible because most of them are running the same background workers or HTTP connection pools.

Goroviz groups these thousands of goroutines into **unique clusters** (similar to how `htop` groups processes), showing you only unique stack signatures.

---

## 🧩 1. The Strategy Interface
Goroviz utilizes a pluggable **Strategy** design pattern. This is defined by the `Strategy` interface in strategy.go

```go
type Strategy interface {
    // Group takes a slice of parsed goroutines and returns them
    // organized into groups based on the strategy's algorithm.
    Group(goroutines []parser.Goroutine) []Group
}
```

By defining this interface, the codebase adheres directly to the **Open-Closed Principle of SOLID design** (*open for extension, closed for modification*). If we want to implement advanced fuzzy matching or similarity clustering in the future, we can write a new strategy class (extending the system) without modifying any existing parsing or UI rendering code.

---

## 🔍 2. How the `ExactMatchStrategy` Works

Right now, Goroviz uses the [ExactMatchStrategy](file:///home/santhosh/Documents/goroviz/internal/group/strategy.go#L26). Here is the step-by-step logic of how it processes goroutines:

### Step A: Generate a "Signature" for each goroutine
For every parsed goroutine, it creates a unique string signature by stitching all of its function names together with `->` arrows.
* **Crucial Rule:** It completely **ignores** line numbers, parameters, and memory addresses.

#### Why is this necessary?
Consider these two goroutines executing concurrently:
* **Goroutine A:** `main.worker(0xc000100)` on `/app/worker.go:15`
* **Goroutine B:** `main.worker(0xc000250)` on `/app/worker.go:18`

If we didn't ignore parameters and line numbers, these would be treated as completely different groups. By stripping those out, both get the same signature: 
`"main.worker -> main.startWorkers"`.

### Step B: Categorize using a Map (HashMap)
It builds a hash map (`map[string][]parser.Goroutine`) where the key is the signature string, and the value is a list of all matching goroutines:

```go
groupMap := make(map[string][]parser.Goroutine)
```

In JSON / JavaScript, this is equivalent to:
```json
{
  "main.worker -> main.startWorkers": [GoroutineA, GoroutineB],
  "net/http.(*conn).serve -> runtime.goexit": [GoroutineC]
}
```

### Step C: Determine the "Dominant State"
For each group, it counts the state of all of its members (e.g. `running`, `IO wait`, `chan receive`). The state that appears the most is set as the overall group state.

### Step D: Sort by Count (htop style)
Finally, it converts the map back into a list of `Group` structs and sorts them in **descending order** (largest group count first). This ensures that if you have a goroutine leak (e.g., 5,000 goroutines blocked on a channel), that group will immediately pop up at the **very top** of your dashboard.

---

## 🎨 Visual Example

Suppose you feed Goroviz 3 goroutines:
1. **Goroutine #1:** `chan receive` $\rightarrow$ `main.worker` $\rightarrow$ `main.main`
2. **Goroutine #2:** `running` $\rightarrow$ `main.printer` $\rightarrow$ `main.main`
3. **Goroutine #3:** `chan receive` $\rightarrow$ `main.worker` $\rightarrow$ `main.main`

### Process flow:
* Goroutines 1 & 3 get the signature: `"main.worker -> main.main"`
* Goroutine 2 gets the signature: `"main.printer -> main.main"`

### Final Output Groups:
* **Group 1:** Signature `main.worker -> main.main` (Count: **2**, State: `chan receive`)
* **Group 2:** Signature `main.printer -> main.main` (Count: **1**, State: `running`)
