# Expense Tracker CLI

A command-line expense tracker. Add, list, delete, and summarize expenses, all persisted to a single `expenses.json` file in the working directory.

## How to use

```sh
go run . add -name coffee -price 3.50
go run . add -name lunch -price 12
go run . list
go run . summary
go run . delete -id 1
```

| Command   | Flags                            | What it does                                  |
| --------- | -------------------------------- | --------------------------------------------- |
| `add`     | `-name <string>`, `-price <num>` | Append a new expense with an auto-assigned ID |
| `list`    | —                                | Print every expense                           |
| `delete`  | `-id <uint>`                     | Remove the expense with the given ID          |
| `summary` | —                                | Print the total spend                         |

## What I learned

- **The `flag` package is per-`FlagSet`, not global.** For subcommands, each command needs its own `flag.NewFlagSet` so `-id` can mean different things in different contexts.
- **`Visit` vs `VisitAll`.** `VisitAll` walks every declared flag; `Visit` only walks the ones the user actually passed. Required-flag detection lives in the gap between them.
- **`json.Unmarshal` needs a pointer; `json.Marshal` doesn't.** Unmarshal mutates the destination, so it needs `&expenses`. Marshal reads and returns bytes. When a function needs to change something, give it the address.
- **Struct tags drive JSON field names.** Without `` `json:"id"` ``, the JSON output uses the Go field name verbatim (`Id`). Tags are read at runtime via reflection — typos in the tag compile fine and silently misbehave.
- **An empty file is a real edge case.** `os.ReadFile` on a 0-byte file succeeds and returns `[]byte{}`; `json.Unmarshal` on that errors. Guard with `len(b) != 0`, or check for `fs.ErrNotExist` first.
- **`errors.Is` over string matching.** `errors.Is(err, fs.ErrNotExist)` works against the error's identity, including wrapped errors. Comparing `err.Error()` strings is fragile.
- **`fmt.Println` always adds spaces and a newline; `fmt.Printf` adds nothing it isn't told to.** Mixing them silently glues output lines together.
- **A no-op write is worse than no write.** When `delete` is called with a non-existent ID, returning early avoids both the wasted disk write *and* the silent "nothing happened" UX.
- **Maps and slices are reference-ish.** The value is a header pointing to backing data, so a function can mutate map contents without needing `*map`. One of Go's subtler design choices.
