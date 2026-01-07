package user_dto

type ParamGetUserByID struct {
	ID string `params:"id" validate:"required,uuid"`
}

type UpdateSelfProfileRequest struct {
	Username string `json:"username,omitempty" validate:"omitempty,min=3,max=30"`
	Name     string `json:"name,omitempty" validate:"omitempty,min=3"`
	Email    string `json:"email,omitempty" validate:"omitempty,email"`
}

type DeactivateSelfUserRequest struct {
	Password string `json:"password" validate:"min=3"`
}
