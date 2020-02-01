package main

import (
	"fmt"
	"testing"
	"time"
)

func TestNotSamePK(t *testing.T) {
	main()

	fmt.Println("---------------------------- Testing ----------------------------")
	fmt.Print("Peers have unique Public Keys...	")
	time.Sleep(10 * time.Millisecond)
	cases := []struct {
		input          string
		expectedOutput string
	}{
		{peer1.PublicKey.String(), peer2.PublicKey.String()},
		{peer1.PublicKey.String(), peer3.PublicKey.String()},
		{peer1.PublicKey.String(), peer4.PublicKey.String()},
		{peer1.PublicKey.String(), peer5.PublicKey.String()},
		{peer1.PublicKey.String(), peer6.PublicKey.String()},
		{peer1.PublicKey.String(), peer7.PublicKey.String()},
		{peer2.PublicKey.String(), peer3.PublicKey.String()},
		{peer2.PublicKey.String(), peer5.PublicKey.String()},
		{peer2.PublicKey.String(), peer4.PublicKey.String()},
		{peer2.PublicKey.String(), peer8.PublicKey.String()},
		{peer3.PublicKey.String(), peer2.PublicKey.String()},
		{peer3.PublicKey.String(), peer4.PublicKey.String()},
		{peer3.PublicKey.String(), peer7.PublicKey.String()},
	}

	passed := true
	for _, c := range cases {
		if output := c.input; output == c.expectedOutput {
			t.Errorf("incorrect output for `%s`: expected `%s` but got `%s`", c.input, c.expectedOutput, output)
			passed = false
		}
	}

	checkPass(passed)
}

func TestNotSameSK(t *testing.T) {
	main()

	fmt.Print("Peers have unique Secret Keys...	")
	time.Sleep(10 * time.Millisecond)
	cases := []struct {
		input          string
		expectedOutput string
	}{
		{peer1.SecretKey.String(), peer2.SecretKey.String()},
		{peer1.SecretKey.String(), peer3.SecretKey.String()},
		{peer1.SecretKey.String(), peer4.SecretKey.String()},
		{peer1.SecretKey.String(), peer5.SecretKey.String()},
		{peer1.SecretKey.String(), peer6.SecretKey.String()},
		{peer1.SecretKey.String(), peer7.SecretKey.String()},
		{peer1.SecretKey.String(), peer8.SecretKey.String()},
		{peer3.SecretKey.String(), peer2.SecretKey.String()},
		{peer3.SecretKey.String(), peer1.SecretKey.String()},
		{peer3.SecretKey.String(), peer4.SecretKey.String()},
		{peer3.SecretKey.String(), peer5.SecretKey.String()},
		{peer2.SecretKey.String(), peer6.SecretKey.String()},
		{peer2.SecretKey.String(), peer7.SecretKey.String()},
		{peer2.SecretKey.String(), peer8.SecretKey.String()},
	}

	passed := true
	for _, c := range cases {
		if output := c.input; output == c.expectedOutput {
			t.Errorf("incorrect output for `%s`: expected `%s` but got `%s`", c.input, c.expectedOutput, output)
			passed = false
		}
	}

	checkPass(passed)
}

func TestRightLengthOfPK(t *testing.T) {
	main()

	fmt.Print("Public Keys have correct length...	")
	time.Sleep(10 * time.Millisecond)
	cases := []struct {
		input          string
		expectedOutput int
	}{
		{peer1.PublicKey.String(), 91},
		{peer3.PublicKey.String(), 91},
		{peer8.PublicKey.String(), 91},
	}

	passed := true
	for _, c := range cases {
		if output := len(c.input); output != c.expectedOutput {
			t.Errorf("incorrect output for `%v`: expected `%v` but got `%v`", c.input, c.expectedOutput, output)
			passed = false
		}
	}
	checkPass(passed)
}

func TestShouldHaveSameAmount(t *testing.T) {
	main()
	fmt.Print("Peers have same starting amount...	")
	time.Sleep(10 * time.Millisecond)
	cases := []struct {
		input          int
		expectedOutput int
	}{
		{peer1.Ledger.Accounts[peer1.PublicKey.String()], peer2.Ledger.Accounts[peer1.PublicKey.String()]},
		{peer3.Ledger.Accounts[peer3.PublicKey.String()], peer4.Ledger.Accounts[peer3.PublicKey.String()]},
		{peer5.Ledger.Accounts[peer1.PublicKey.String()], peer8.Ledger.Accounts[peer1.PublicKey.String()]},
		{peer8.Ledger.Accounts[peer8.PublicKey.String()], peer7.Ledger.Accounts[peer8.PublicKey.String()]},
	}

	passed := true
	for _, c := range cases {
		if output := c.input; output != c.expectedOutput {
			t.Errorf("incorrect output for `%v`: expected `%v` but got `%v`", c.input, c.expectedOutput, output)
			passed = false
		}
	}
	checkPass(passed)
}

func TestShouldHaveSameAmountAfterSend(t *testing.T) {
	main()
	fmt.Print("Ledgers update correctly...			")
	time.Sleep(10 * time.Millisecond)
	peer1.Sender(peer3, 500)
	time.Sleep(10 * time.Millisecond)
	peer3.Sender(peer1, 499)
	time.Sleep(10 * time.Millisecond)
	peer8.Sender(peer2, 1337)
	time.Sleep(10 * time.Millisecond)
	peer2.Sender(peer5, 999)
	time.Sleep(2000 * time.Millisecond)
	cases := []struct {
		input          int
		expectedOutput int
	}{
		{peer1.Ledger.Accounts[peer1.PublicKey.String()], peer2.Ledger.Accounts[peer1.PublicKey.String()]},
		{peer3.Ledger.Accounts[peer3.PublicKey.String()], peer4.Ledger.Accounts[peer3.PublicKey.String()]},
		{peer5.Ledger.Accounts[peer1.PublicKey.String()], peer8.Ledger.Accounts[peer1.PublicKey.String()]},
		{peer8.Ledger.Accounts[peer8.PublicKey.String()], peer1.Ledger.Accounts[peer8.PublicKey.String()]},
		{peer7.Ledger.Accounts[peer5.PublicKey.String()], peer2.Ledger.Accounts[peer5.PublicKey.String()]},
		{peer6.Ledger.Accounts[peer5.PublicKey.String()], peer3.Ledger.Accounts[peer5.PublicKey.String()]},
		{peer5.Ledger.Accounts[peer5.PublicKey.String()], peer4.Ledger.Accounts[peer5.PublicKey.String()]},
		{peer4.Ledger.Accounts[peer5.PublicKey.String()], peer5.Ledger.Accounts[peer5.PublicKey.String()]},
		{peer3.Ledger.Accounts[peer5.PublicKey.String()], peer6.Ledger.Accounts[peer5.PublicKey.String()]},
		{peer2.Ledger.Accounts[peer5.PublicKey.String()], peer7.Ledger.Accounts[peer5.PublicKey.String()]},
		{peer1.Ledger.Accounts[peer5.PublicKey.String()], peer8.Ledger.Accounts[peer5.PublicKey.String()]},
		{peer7.Ledger.Accounts[peer2.PublicKey.String()], peer2.Ledger.Accounts[peer2.PublicKey.String()]},
		{peer6.Ledger.Accounts[peer2.PublicKey.String()], peer3.Ledger.Accounts[peer2.PublicKey.String()]},
		{peer5.Ledger.Accounts[peer2.PublicKey.String()], peer4.Ledger.Accounts[peer2.PublicKey.String()]},
		{peer4.Ledger.Accounts[peer8.PublicKey.String()], peer5.Ledger.Accounts[peer8.PublicKey.String()]},
		{peer3.Ledger.Accounts[peer8.PublicKey.String()], peer6.Ledger.Accounts[peer8.PublicKey.String()]},
		{peer2.Ledger.Accounts[peer8.PublicKey.String()], peer7.Ledger.Accounts[peer8.PublicKey.String()]},
		{peer1.Ledger.Accounts[peer8.PublicKey.String()], peer8.Ledger.Accounts[peer8.PublicKey.String()]},
	}

	passed := true
	for _, c := range cases {
		if output := c.input; output != c.expectedOutput {
			t.Errorf("incorrect output for `%v`: expected `%v` but got `%v`", c.input, c.expectedOutput, output)
			passed = false
		}
	}
	checkPass(passed)
}

func TestShouldNotTransferMoreThanHave(t *testing.T) {
	main()
	fmt.Print("Cannot make negative transactions...")
	time.Sleep(10 * time.Millisecond)
	peer1.Sender(peer3, 1000001)
	time.Sleep(10 * time.Millisecond)
	peer4.Sender(peer5, 1)
	time.Sleep(10 * time.Millisecond)
	peer5.Sender(peer4, -10000)
	time.Sleep(10 * time.Millisecond)
	cases := []struct {
		input          int
		expectedOutput int
	}{
		{peer1.Ledger.Accounts[peer1.PublicKey.String()], 1000000},
		{peer3.Ledger.Accounts[peer3.PublicKey.String()], 1000000},
		{peer8.Ledger.Accounts[peer5.PublicKey.String()], 1000000},
		{peer4.Ledger.Accounts[peer4.PublicKey.String()], 1000000},
	}

	passed := true
	for _, c := range cases {
		if output := c.input; output != c.expectedOutput {
			t.Errorf("incorrect output for `%v`: expected `%v` but got `%v`", c.input, c.expectedOutput, output)
			passed = false
		}
	}
	checkPass(passed)
}

func checkPass(passed bool) {
	if passed {
		fmt.Println(" 	Test Passed")
	} else {
		fmt.Println(" 	Test Failed!")
	}
}
