package requests

type CashbackDecreaseQueue struct {
	CineramaUserID         int     `json:"cinerama_user_id" binding:"required"`
	DecreaseCashbackAmount float64 `json:"decrease_cashback_amount" binding:"required"`
	Device                 string  `json:"device" binding:"required"`
	HostIP                 string  `json:"host_ip" binding:"required"`
	TraceCode              string  `json:"-"`
}
