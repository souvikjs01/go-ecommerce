package request

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type SignupRequest struct {
	Username     string  `json:"username" binding:"required"`
	FirstName    string  `json:"firstName" binding:"required"`
	LastName     string  `json:"lastName" binding:"required"`
	Email        string  `json:"email" binding:"required"`
	Password     string  `json:"password" binding:"required,min=4"`
	Gender       string  `json:"gender" binding:"required,oneof=male female other"`
	ProfileImage *string `json:"profileImage"`
}

type UpdateRequest struct {
	Username     *string `json:"username"`
	FirstName    *string `json:"firstName"`
	LastName     *string `json:"lastName"`
	ProfileImage *string `json:"profileImage"`
}
