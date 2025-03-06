package main

import (
	"context"
	"fmt"
	"github.com/travis-g/dice"
	"log"
	"os"
	"strconv"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("usage: roll-dice <command>")
	}
	ctx := context.Background()

	command := os.Args[1]

	numDice, err := strconv.Atoi(os.Getenv("NUM_DICE"))
	if err != nil {
		log.Fatal("num_dice was not set to a number")
	}

	numSides, err := strconv.Atoi(os.Getenv("NUM_SIDES"))
	if err != nil {
		log.Fatal("num_sides was not set to a number")
	}

	switch command {
	case "rollDice":
		group, err := dice.NewRollerGroup(
			&dice.RollerProperties{
				Size:  numSides,
				Count: numDice,
			})
		if err != nil {
			log.Fatal(err)
		}
		err = group.FullRoll(ctx)
		if err != nil {
			log.Fatal(err)
		}
		result, err := group.Total(ctx)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("%dd%d rolled, result: %d\n", numDice, numSides, int(result))
	}
}
