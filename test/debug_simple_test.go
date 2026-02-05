package test

import (
	"testing"
	"time"

	"github.com/Crescent617/json-repair-go/jsonrepair"
)

func TestBracketDirect(t *testing.T) {
	input := "["

	// Test direct parser
	parser := jsonrepair.NewParser(input)

	done := make(chan bool, 1)
	var result interface{}
	var err error

	go func() {
		result, err = parser.Parse()
		done <- true
	}()

	select {
	case <-done:
		t.Logf("Parse completed: result=%v, err=%v\n", result, err)
	case <-time.After(2 * time.Second):
		t.Fatal("Parser.Parse() timed out after 2 seconds")
	}
}

func TestBracketAPI(t *testing.T) {
	input := "["

	done := make(chan bool, 1)
	var result string
	var err error

	go func() {
		result, err = jsonrepair.RepairJSON(input)
		done <- true
	}()

	select {
	case <-done:
		t.Logf("RepairJSON completed: result=%s, err=%v\n", result, err)
	case <-time.After(2 * time.Second):
		t.Fatal("RepairJSON timed out after 2 seconds")
	}
}
