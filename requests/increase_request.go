package requests

type CashbackRequest struct {
	CineramaUserID int     `json:"cinerama_user_id" binding:"required"`
	CashbackAmount    float64 `json:"cashback_amount" binding:"required"`
	Device         string  `json:"device" binding:"required"`
	HostIP         string  `json:"host_ip" binding:"required"`
	TraceCode      string  `json:"-"`
}
