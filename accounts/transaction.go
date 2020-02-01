package accounts

type Transaction struct {
	ID        string
	From      string
	To        string
	Amount    int
	Signature string
}

func (l *Ledger) Transaction(t *Transaction) {
	l.lock.Lock()
	defer l.lock.Unlock()
	l.Accounts[t.From] -= t.Amount
	l.Accounts[t.To] += t.Amount
}
