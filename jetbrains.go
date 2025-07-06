package main

type SSEData struct {
	Type      string       `json:"type"`
	EventType string       `json:"event_type"`
	Content   string       `json:"content,omitempty"`
	Reason    string       `json:"reason,omitempty"`
	Updated   *UpdatedData `json:"updated,omitempty"`
	Spent     *SpentData   `json:"spent,omitempty"`
}

type UpdatedData struct {
	License string     `json:"license"`
	Current AmountData `json:"current"`
	Maximum AmountData `json:"maximum"`
	Until   int64      `json:"until"`
	QuotaID QuotaInfo  `json:"quotaID"`
}

type AmountData struct {
	Amount string `json:"amount"`
}

type QuotaInfo struct {
	QuotaId string `json:"quotaId"`
}

type SpentData struct {
	Amount string `json:"amount"`
}
