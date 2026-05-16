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
		validate_arguments(addCmd, map[string]bool{"name": false, "price": false})

		add_expense(addName, addPrice)
	case "list":
		listCmd.Parse(os.Args[2:])
		list_expenses()
	case "delete":
		deleteCmd.Parse(os.Args[2:])
		validate_arguments(deleteCmd, map[string]bool{"id": false})

		delete_expense(deleteId)
	case "summary":
		summaryCmd.Parse(os.Args[2:])
		summarize()
	}
}

func validate_arguments(flags *flag.FlagSet, requirement_list map[string]bool) {
	flags.Visit(func(f *flag.Flag) {
		requirement_list[f.Name] = true
	})

	for flag, set := range requirement_list {
		if !set {
			fmt.Println(flag, "flag not set!")
			flags.Usage()
			os.Exit(1)
		}
	}
}

func add_expense(expense_name *string, expense_price *float64) {
	fmt.Println("Expense name:", expense_name)
	fmt.Printf("Expense price: €%.2f\n", *expense_price)
}

func list_expenses() {

}

func delete_expense(expense_id *uint) {

}

func summarize() {

}
