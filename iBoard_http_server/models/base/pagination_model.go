package base_models

type PaginationQuery struct {
	PageNum  int  `form:"pageNum" json:"pageNum"`
	PageSize int  `form:"pageSize" json:"pageSize"`
	Desc     bool `form:"desc" json:"desc"`
}

type PaginationResult struct {
	Total    int `form:"total" json:"total"`
	PageNum  int `form:"pageNum" json:"pageNum"`
	PageSize int `form:"pageSize" json:"pageSize"`
}
