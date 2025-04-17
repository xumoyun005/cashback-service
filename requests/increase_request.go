package requests

type CashbackRequest struct {
	CineramaUserID int     `json:"cinerama_user_id" binding:"required"`
	TariffPrice    float64 `json:"tariff_price" binding:"required"`
	Device         string  `json:"device" binding:"required"`
	HostIP         string  `json:"host_ip" binding:"required"`
	TraceCode      string  `json:"-"`
}
