# 0.3 — Expense Tracker CLI

A small command-line expense tracker. Add, list, delete, and summarize expenses, all persisted to a single `expenses.json` file in the working directory.

This is project **0.3** of my Go learning track — the final project of Phase 0 (fundamentals), focused on getting comfortable with subcommands, JSON serialization, and the read-modify-write pattern.

## How it works

The program is one binary with four subcommands: `add`, `list`, `delete`, `summary`. Each one is a separate `flag.FlagSet` so that each subcommand can declare its own flags without colliding. Expenses live in a JSON array on disk; every mutating command reads the whole file, mutates the slice in memory, and writes the whole file back.

```go
type Expense struct {
    Id    uint64  `json:"id"`
    Name  string  `json:"name"`
    Price float64 `json:"price"`
}
```

### Subcommands with `flag.NewFlagSet`

The standard `flag.Parse()` only handles one set of flags for the whole program. That doesn't fit subcommand-style CLIs (`git commit -m`, `go test -v`) where each command has its own flags. The fix is `flag.NewFlagSet` per subcommand:

```go
addCmd := flag.NewFlagSet("add", flag.ExitOnError)
addName := addCmd.String("name", "", "Name of the expense")
addPrice := addCmd.Float64("price", 0.00, "Cost of the expense")

deleteCmd := flag.NewFlagSet("delete", flag.ExitOnError)
deleteId := deleteCmd.Uint64("id", 0, "The id of the expense")
```

`main` switches on `os.Args[1]` to pick which flag set to parse against `os.Args[2:]`. This keeps `add -name foo -price 5` and `delete -id 3` cleanly separated and gives each subcommand its own `Usage()` output.

### Required-flag validation via `flags.Visit`

The `flag` package doesn't have a "this flag is required" feature — every flag has a default. To require `-name` and `-price` on `add`, the program uses `flags.Visit`, which only iterates over flags that were *actually set on the command line* (as opposed to `VisitAll`, which iterates over every declared flag):

```go
func validateArguments(flags *flag.FlagSet, requirementList map[string]bool) {
    flags.Visit(func(f *flag.Flag) {
        requirementList[f.Name] = true
    })

    for flag, set := range requirementList {
        if !set {
            fmt.Println(flag, "flag not set!")
            flags.Usage()
            os.Exit(1)
        }
    }
}
```

The caller passes in a map of `{"name": false, "price": false}`. `Visit` flips each key to `true` if the user provided that flag. Anything still `false` is missing.

### JSON persistence as whole-file rewrite

For a small CLI with a few hundred entries, streaming JSON or append-only formats are overkill. Every mutating command does the same three-step dance: read the file into a `[]Expense`, modify it, write the slice back. Two helpers factor out the duplication:

```go
func getExpenses() []Expense {
    b, err := os.ReadFile("expenses.json")
    if errors.Is(err, fs.ErrNotExist) {
        return nil
    }
    handleErr(err)

    var expenses []Expense
    if len(b) != 0 {
        err = json.Unmarshal(b, &expenses)
        handleErr(err)
    }
    return expenses
}

func writeToJson(expenses []Expense) {
    bNew, err := json.Marshal(expenses)
    handleErr(err)
    err = os.WriteFile("expenses.json", bNew, 0666)
    handleErr(err)
}
```

Two edge cases get explicit handling:

- **The file doesn't exist yet.** `errors.Is(err, fs.ErrNotExist)` returns a nil slice on first run, so the subsequent `append` allocates fresh. `os.WriteFile` will create the file on the first save.
- **The file exists but is empty (0 bytes).** `json.Unmarshal` on `[]byte{}` returns "unexpected end of JSON input." A `len(b) != 0` guard skips the unmarshal entirely.

`os.ReadFile`/`os.WriteFile` are used instead of `os.OpenFile` here because the entire payload fits in memory and there's no need to manage a long-lived file handle. They handle open+close+truncate in one call.

### Auto-increment IDs

Every expense gets a unique numeric ID so `delete` can target one specific record (deleting by name would fail on duplicates — two coffees in the same week shouldn't collapse). The new ID is computed by scanning the existing slice for the maximum and adding one:

```go
var maxId uint64 = 0
for _, expense := range expenses {
    if expense.Id > maxId {
        maxId = expense.Id
    }
}

newExpense := Expense{
    Id:    maxId + 1,
    Name:  *expenseName,
    Price: *expensePrice,
}
```

This is O(n) per add, which is fine for a personal tracker. For a system with millions of records I'd use a UUID or store the next-id separately.

### Delete is a filter, not a splice

`deleteExpense` builds a new slice containing every entry whose ID doesn't match, plus a flag tracking whether the target was actually found:

```go
for _, expense := range expenses {
    if expense.Id != *expenseIdToDelete {
        newExpenses = append(newExpenses, expense)
    } else {
        hasOccurred = true
    }
}

if !hasOccurred {
    fmt.Println("expense with ID", *expenseIdToDelete, "does not exist!")
    return
}

writeToJson(newExpenses)
```

The early `return` on a missing ID matters: without it, the program would happily rewrite `expenses.json` with the same content, wasting a disk write and giving the user no feedback that their command didn't do what they expected.

## How to run it

```sh
go run . add -name coffee -price 3.50
go run . add -name lunch -price 12
go run . list
go run . summary
go run . delete -id 1
```

### Subcommands

| Command   | Flags                            | What it does                                  |
| --------- | -------------------------------- | --------------------------------------------- |
| `add`     | `-name <string>`, `-price <num>` | Append a new expense with an auto-assigned ID |
| `list`    | —                                | Print every expense to stdout                 |
| `delete`  | `-id <uint>`                     | Remove the expense with the given ID          |
| `summary` | —                                | Print the total spend across all expenses     |

### Example session

```
$ go run . add -name coffee -price 3.50
$ go run . add -name lunch -price 12
$ go run . list
Id: 1
Name: coffee
Price €3.50
Id: 2
Name: lunch
Price €12.00
$ go run . summary
Total price of expenses: €15.50
$ go run . delete -id 1
$ go run . list
Id: 2
Name: lunch
Price €12.00
```

## What I learned

- **The `flag` package is per-`FlagSet`, not global.** `flag.Parse()` operates on a single hidden default set. For subcommands, each command needs its own `flag.NewFlagSet` so flag names like `-id` can mean different things in different contexts. The same `Usage()`, `Parse(args)`, and `Visit(...)` methods exist on every set.
- **`Visit` vs `VisitAll`.** `VisitAll` walks every declared flag whether or not it was set; `Visit` only walks the ones the user actually passed on the command line. Required-flag detection lives in the gap between those two.
- **`json.Unmarshal` needs a pointer; `json.Marshal` doesn't.** Unmarshal *mutates* the destination, so it needs `&expenses` to write into the caller's variable. Marshal just *reads* the data and returns bytes, so a plain `expenses` is fine. The asymmetry maps onto Go's pass-by-value rule: when a function needs to change something, give it the address.
- **Struct tags drive JSON field names.** Without `` `json:"id"` `` on a field, the JSON output uses the Go field name verbatim — so `Id` becomes `"Id"` instead of the conventional `"id"`. Tags are read by `encoding/json` at runtime via reflection, not by the compiler.
- **An empty file is a real edge case.** `os.ReadFile` on a 0-byte file succeeds and returns `[]byte{}`. `json.Unmarshal` on that returns an error. Either guard against empty input before unmarshaling, or check for a missing file with `errors.Is(err, fs.ErrNotExist)` and short-circuit before reading.
- **`errors.Is` over string matching.** Go's `errors.Is(err, fs.ErrNotExist)` works against the error's *identity*, including any wrapped errors. Comparing against `err.Error()` string fragments is fragile — error messages can change between Go versions, and they don't survive wrapping.
- **`fmt.Println` always adds spaces and a newline; `fmt.Printf` adds nothing it isn't told to.** Mixing them in `listExpenses` produced a visible bug: the price line lacked a trailing `\n`, so the next iteration's `Id:` glued onto it. Either commit to `Println` everywhere, or use `Printf` with explicit `\n`.
- **DRY by helper, not by abstraction.** The first version had the read-modify-write block inlined in every command. Extracting `getExpenses()` returning `[]Expense` (not `[]byte`) collapsed all the duplication because every command really did want the same thing — the parsed slice, with empty-file handling done. The right level of abstraction was the one that matched what callers actually needed.
- **A no-op write is worse than no write.** When `delete` is called with a non-existent ID, returning before the `writeToJson` call avoids both the wasted disk write *and* the silent user experience of "I asked to delete something and nothing said it failed."
- **Pass-by-value almost always, except…** Maps and slices are reference-ish under the hood — the value is a header pointing to the backing data, so a function that mutates a map's contents (like `validateArguments` flipping `requirementList[f.Name] = true`) affects the caller without needing `*map`. One of Go's subtler design choices, and a gap I want to firm up before Phase 1.
