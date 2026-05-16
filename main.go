package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
)

type Expense struct {
	Id    uint64  `json:"id"`
	Name  string  `json:"name"`
	Price float64 `json:"price"`
}

func main() {
	// Create file if it doesnt exist yet
	f, err := os.OpenFile("expenses.json", os.O_CREATE|os.O_WRONLY, 0666)

	handleErr(err)

	defer f.Close()

	addCmd := flag.NewFlagSet("add", flag.ExitOnError)
	addName := addCmd.String("name", "", "Name of the expense")
	addPrice := addCmd.Float64("price", 0.00, "Cost of the expense")

	listCmd := flag.NewFlagSet("list", flag.ExitOnError)

	deleteCmd := flag.NewFlagSet("delete", flag.ExitOnError)
	deleteId := deleteCmd.Uint64("id", 0, "The id of the expense")

	summaryCmd := flag.NewFlagSet("summary", flag.ExitOnError)

	if len(os.Args) < 2 {
		fmt.Println("No arguments given.")
		os.Exit(1)
	}

	switch os.Args[1] {
	case "add":
		addCmd.Parse(os.Args[2:])
		validateArguments(addCmd, map[string]bool{"name": false, "price": false})

		addExpense(addName, addPrice)
	case "list":
		listCmd.Parse(os.Args[2:])
		listExpenses()
	case "delete":
		deleteCmd.Parse(os.Args[2:])
		validateArguments(deleteCmd, map[string]bool{"id": false})

		deleteExpense(deleteId)
	case "summary":
		summaryCmd.Parse(os.Args[2:])
		summarize()
	}
}

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

func addExpense(expenseName *string, expensePrice *float64) {
	expenses := getExpenses()

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

	expenses = append(expenses, newExpense)

	writeToJson(expenses)
}

func listExpenses() {
	expenses := getExpenses()

	for _, expense := range expenses {
		fmt.Println("Id:", expense.Id)
		fmt.Println("Name:", expense.Name)
		fmt.Printf("Price €%.2f", expense.Price)
	}
}

func deleteExpense(expenseIdToDelete *uint64) {
	expenses := getExpenses()

	var newExpenses []Expense

	hasOccured := false

	for _, expense := range expenses {
		if expense.Id != *expenseIdToDelete {
			newExpenses = append(newExpenses, expense)
		} else {
			hasOccured = true
		}
	}

	if !hasOccured {
		fmt.Println("expense with ID", *expenseIdToDelete, "does not exist!")

		return
	}

	writeToJson(newExpenses)
}

func summarize() {
	expenses := getExpenses()

	totalPrice := 0.0

	for _, expense := range expenses {
		totalPrice += expense.Price
	}

	fmt.Printf("Total price of expenses: €%.2f", totalPrice)
}

func handleErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func getExpenses() []Expense {
	b, err := os.ReadFile("expenses.json")

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
