# Code Review

The primary goal of this comprehensive code review task is to ensure the code is **correct, maintainable, efficient, and simple**.

Workflow:

- Read all the following documents:
  - README.md
- Read all the \*.go files
- Analyse the code using the topics below

Outcome:

- Once the review is complete, document all of your findings in a new document named PROJECT-code-review-rectification.md
- This document is a stand-alone rectification project to be carried out at a later date
- Put everything you need into the project document to complete a rectification process

---

## 1. Correctness and Functionality

This is the most fundamental part: does the code do what it's supposed to do?

- **Requirements:** Does the code correctly implement the feature or fix the bug described in the ticket/issue?
- **Logic:** Are there any flaws in the algorithm or logic?
- **Edge Cases:** Are edge cases handled?
  - Nil inputs (pointers, slices, maps, channels).
  - Empty inputs (empty slices, maps, strings).
  - Zero values (e.g., `0`, `""`).
  - Handling of upper/lower bounds, off-by-one errors.
- **Race Conditions:** (See Concurrency section) Is the code free of data races? Run tests with the `-race` flag.

---

## 2. Design and Architecture

Does the code fit well into the larger system and follow good design principles?

- **Simplicity:** Is the code unnecessarily complex? Go values **simplicity over cleverness**. Avoid overly abstract or "magic" code.
- **Cohesion:** Are packages and structs well-defined and focused on a single responsibility?
- **Coupling:** Is the code loosely coupled?
- **Interfaces:**
  - Are interfaces used effectively to decouple components?
  - Follow the "accept interfaces, return structs" principle.
  - Are interfaces small and focused (often just one or two methods)?
  - Avoid using `interface{}` (or `any`) unless absolutely necessary.
- **Package Structure:**
  - Is the package name short, lowercase, and representative?
  - Does the new code belong in this package, or should it be in a new or different one?
  - Are there any circular dependencies being introduced?

---

## 3. Idiomatic Go (The "Go Way")

This is about writing code that other Go developers will find familiar and easy to read.

- **Formatting:** Has `gofmt` (or `goimports`) been run? This is non-negotiable and should be automated.
- **Linting:** Does the code pass a linter like `golangci-lint` or `staticcheck`? These tools automatically catch a vast number of common issues.
- **Zero Value:** Is the zero value for a struct useful? (e.g., `sync.Mutex` works correctly as its zero value). Avoid "constructor" functions (`New...`) that just set default values if the zero value is sufficient.
- **Pointers vs. Values:**
  - Are parameters passed as pointers when they need to be mutated?
  - Are large structs passed as pointers to avoid expensive copies?
  - Conversely, are pointers used when a simple value copy would be safer and clearer?
- **Composition:** Does the code favor struct embedding (composition) over inheritance?

---

## 4. Error Handling

Error handling is a first-class concern in Go.

- **Check Every Error:** Is _every_ function that returns an `error` being checked? Never ignore errors with `_`.
- **Error Wrapping:** Are errors from downstream calls wrapped with context using `fmt.Errorf("...: %w", err)`? This preserves the error chain.
- **Error Checking:** Is code using `errors.Is()` for checking specific sentinel errors (like `io.EOF`) and `errors.As()` for checking specific error types?
- **Error Naming:** Are error variables named `err`? Are new error types named `ErrSomething` (e.g., `ErrNotFound`)?
- **No `panic`:** Is `panic` avoided for normal error conditions? `panic` should only be for truly unrecoverable situations (e.g., programming error during initialization).

---

## 5. Concurrency

Concurrency is a core feature of Go, but it's easy to get wrong.

- **Race Conditions:** Is data shared between goroutines? If so, is it protected by a mutex (`sync.Mutex` or `sync.RWMutex`) or handled via channels?
- **Goroutine Leaks:** Is every goroutine launched guaranteed to exit?
  - If using channels, are senders/receivers correctly handling closed channels?
  - If using `select`, is there a `default` or `context.Done()` case to prevent blocking forever?
- **`context.Context`:**
  - Is `ctx` passed as the **first argument** to any function that performs I/O, makes network calls, or could be long-running?
  - Is the context respected (e.g., checking `ctx.Done()`) to enable cancellation and deadlines?
- **`sync.WaitGroup`:** If using a `WaitGroup`, is `Add()` called _before_ the goroutine is launched, and `Done()` called _inside_ the goroutine (often with `defer`)?
- **Channel Usage:** Are channels being used correctly? Are buffered or unbuffered channels appropriate for the use case?

---

## 6. Testing

No code is complete without tests.

- **Coverage:** Are there meaningful tests? (High coverage is good, but 100% coverage of trivial code is not as valuable as good tests for complex logic).
- **Test Cases:** Do tests cover the "happy path," error conditions, and edge cases?
- **Table-Driven Tests:** Are table-driven tests used for functions that have multiple inputs and expected outputs? This is a very common and readable Go idiom.
- **Test Naming:** Do tests follow the `TestMyFunction` convention?
- **Race Detector:** Are tests run with the `-race` flag?

---

## 7. Performance and Resource Management

Is the code efficient and well-behaved?

- **`defer` for Cleanup:** Are resources like files (`os.File`), network connections, and HTTP response bodies (`resp.Body`) closed? Is `defer` used to ensure they are _always_ closed, even if an error occurs?
- **Allocations:** Is the code avoiding unnecessary memory allocations in hot paths (e.g., in a tight loop)?
  - Is `make` used with a size hint to pre-allocate slices or maps if the size is known?
- **String Concatenation:** If building strings in a loop, is `strings.Builder` or `bytes.Buffer` used instead of repeated `+` concatenation?

---

## 8. Naming and Documentation

Is the code understandable at a glance?

- **Naming:**
  - Are names clear, concise, and descriptive? (e.g., `i` for a loop index, but `userIndex` for a more complex index).
  - Is `CamelCase` used for exported identifiers (public) and `camelCase` for unexported (private)?
  - Are package names short, lowercase, and singular (e.g., `http`, `user`, `parser`)?
- **Comments:**
  - Is all exported code (functions, types, constants, variables) documented with a proper Go doc-comment (e.g., `// MyFunction does...`)?
  - Are comments used to explain _why_ something is done, not _what_ is being done (the code should explain the _what_)?
  - Are `// TODO:` or `// FIXME:` comments used appropriately for follow-up work?
