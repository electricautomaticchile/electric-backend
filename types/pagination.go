package types

type PaginationParams struct {
	Page     int
	PageSize int
	SortBy   string
	SortDir  string
}

type PaginatedResponse struct {
	Data       interface{} `json:"data"`
	Page       int         `json:"page"`
	PageSize   int         `json:"pageSize"`
	Total      int64       `json:"total"`
	TotalPages int         `json:"totalPages"`
	HasNext    bool        `json:"hasNext"`
	HasPrev    bool        `json:"hasPrev"`
}

func NewPaginationParams(page, pageSize int, sortBy, sortDir string) PaginationParams {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}
	if sortDir != "asc" && sortDir != "desc" {
		sortDir = "desc"
	}
	
	return PaginationParams{
		Page:     page,
		PageSize: pageSize,
		SortBy:   sortBy,
		SortDir:  sortDir,
	}
}

func (p PaginationParams) GetSkip() int {
	return (p.Page - 1) * p.PageSize
}

func (p PaginationParams) GetLimit() int {
	return p.PageSize
}

func NewPaginatedResponse(data interface{}, page, pageSize int, total int64) PaginatedResponse {
	totalPages := int(total) / pageSize
	if int(total)%pageSize > 0 {
		totalPages++
	}

	return PaginatedResponse{
		Data:       data,
		Page:       page,
		PageSize:   pageSize,
		Total:      total,
		TotalPages: totalPages,
		HasNext:    page < totalPages,
		HasPrev:    page > 1,
	}
}
