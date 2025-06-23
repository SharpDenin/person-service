package domain

type CreatePersonRequest struct {
	Name       string `json:"name" binding:"required"`
	Surname    string `json:"surname" binding:"required"`
	Patronymic string `json:"patronymic,omitempty"`
}

type UpdatePersonRequest struct {
	Name        string `json:"name,omitempty"`
	Surname     string `json:"surname,omitempty"`
	Patronymic  string `json:"patronymic,omitempty"`
	Age         int    `json:"age,omitempty"`
	Gender      string `json:"gender,omitempty"`
	Nationality string `json:"nationality,omitempty"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

type PersonListResponse struct {
	Data []*Person      `json:"data"`
	Meta PaginationMeta `json:"meta"`
}

type PaginationMeta struct {
	Page       int `json:"page"`
	PageSize   int `json:"page_size"`
	TotalItems int `json:"total_items"`
}
