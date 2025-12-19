package param

type Pagination struct {
	PageNum  int `json:"page" form:"page" query:"page"`
	PageSize int `json:"size" form:"size" query:"size"`
}

type Sort struct {
	Fields []string `json:"sort" form:"sort" query:"sort"`
}
