package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	addCmd := flag.NewFlagSet("add", flag.ExitOnError)
	addName := addCmd.String("name", "", "Name of the expense")
	addPrice := addCmd.Float64("price", 0.00, "Cost of the expense")

	listCmd := flag.NewFlagSet("list", flag.ExitOnError)

	deleteCmd := flag.NewFlagSet("delete", flag.ExitOnError)
	deleteId := deleteCmd.Uint("id", 0, "The id of the expense")

	summaryCmd := flag.NewFlagSet("summary", flag.ExitOnError)

	if len(os.Args) < 2 {
		fmt.Println("No arguments given.")
		os.Exit(1)
	}

	switch os.Args[1] {
	case "add":
		addCmd.Parse(os.Args[2:])
		required := map[string]bool{"name": false, "price": false}
		addCmd.Visit(func(f *flag.Flag) {
			required[f.Name] = true
		})

		for flag, set := range required {
			if !set {
				fmt.Println(flag, "flag not set!")
				addCmd.Usage()
				os.Exit(1)
			}
		}

		fmt.Println("Expense name:", *addName)
		fmt.Printf("Expense price: €%.2f\n", *addPrice)
	case "list":
		listCmd.Parse(os.Args[2:])
		fmt.Println("Printing the expenses")
	case "delete":
		deleteCmd.Parse(os.Args[2:])
		required := map[string]bool{"id": false}
		deleteCmd.Visit(func(f *flag.Flag) {
			required[f.Name] = true
		})

		for flag, set := range required {
			if !set {
				fmt.Println(flag, "flag not set!")
				deleteCmd.Usage()
				os.Exit(1)
			}
		}

		fmt.Println("Deleting expense with id", *deleteId)
	case "summary":
		summaryCmd.Parse(os.Args[2:])
		fmt.Println("Printing the summary")
	}
}
