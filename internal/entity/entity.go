package entity

type Account struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type Balance struct {
	Current   float64 `json:"current"`
	Withdrawn float64 `json:"withdrawn"`
}

type Order struct {
	Number     string  `json:"number"`
	Status     string  `json:"status"`
	Accrual    float64 `json:"accrual,omitempty"`
	UploadedAt string  `json:"uploaded_at"`
}

type Withdraw struct {
	Order       string  `json:"order"`
	Sum         float64 `json:"sum"`
	UserID      uint64  `json:"-"`
	ProcessedAt string  `json:"processed_at"`
}
