package role

type ProfileUser struct {
	FullName   string `json:"full_name"`
	Gender     string `json:"gender"`
	Avatar     string `json:"avatar"`
	RoleID     int    `json:"role_id"`
	RoleName   string `json:"role_name"`
	IsBlocked  bool   `json:"is_blocked"`
	IsSendWise bool   `json:"is_send_wise"`
}
