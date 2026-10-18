package dashboard

// Response is the dashboard payload.
type Resposta struct {
	TotalRaisedCents  int64                `json:"total_raised_cents"`
	TotalExpenseCents int64                `json:"total_expense_cents"`
	BalanceCents      int64                `json:"balance_cents"`
	TotalDonors       int                  `json:"total_donors"`
	PendingPayments   int                  `json:"pending_payments"`
	FailedPayments    int                  `json:"failed_payments"`
	CanceledRecurring int                  `json:"canceled_recurring"`
	ActiveRecurring   int                  `json:"active_recurring"`
	CommunityRanking  []TotalPorComunidade `json:"community_ranking"`
}

// CommunityTotal is one row of the ranking.
type TotalPorComunidade struct {
	CommunityID   string `json:"community_id"`
	CommunityName string `json:"community_name"`
	AmountCents   int64  `json:"amount_cents"`
	Contributions int    `json:"contributions"`
}
