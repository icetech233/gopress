---
title: Guide
---
# GoPress Guide Page

## Quick Start with Go

Below is a simple "Hello, World!" example in Go — a great starting point for beginners:

```
package main

import "fmt"

func main() {
    fmt.Println("Hello Go")
}
```

### Key Notes:
- The `package main` declaration indicates this is an executable program.
- The `import "fmt"` statement brings in Go’s standard formatting package.
- `fmt.Println` outputs text to the console.

## Common Development Tasks

### 1. Setting Up Your Environment
Make sure you have [Go installed](https://golang.org/dl/) (version 1.20 or higher recommended). Verify your installation with:

```
go version
```

### 2. Creating a New Project
Initialize a new module for your GoPress project:

```
mkdir my-gopress-site
cd my-gopress-site
go mod init my-gopress-site
```

### 3. Running Your Code
Save the example above in a file named `main.go`, then run it:

```
go run main.go
```

You should see:
```
Hello Go
```

## Integration with GoPress

GoPress leverages Go’s simplicity to enable fast static site generation, API integrations, and custom tooling. You can extend GoPress by:

- Writing custom handlers in Go
- Using Go templates for dynamic content
- Deploying compiled binaries for high performance

> 💡 **Tip**: Explore the [IceTest](/icetest/index.html) module to validate your Go components within the GoPress environment.

---

[← Back Home](/index.html)