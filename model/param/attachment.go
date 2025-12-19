package param

import "github.com/go-sonic/sonic/consts"

type AttachmentQuery struct {
	Pagination
	Keyword        string                 `json:"keyword" form:"keyword" query:"keyword"`
	MediaType      string                 `json:"mediaType" form:"mediaType" query:"mediaType"`
	AttachmentType *consts.AttachmentType `json:"attachmentType" form:"attachmentType" query:"attachmentType"`
}

type AttachmentUpdate struct {
	Name string `json:"name" binding:"gte=1,lte=255"`
}
