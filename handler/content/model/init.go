package model

import (
	"github.com/go-sonic/sonic/injection"
	"github.com/go-sonic/sonic/service/impl"
)

func init() {
	// 注册缓存服务
	injection.Provide(impl.NewCachedPostService)
	injection.Provide(impl.NewCachedCategoryService)
	injection.Provide(impl.NewCachedTagService)
	// 使用带缓存的 PostModel 替换原有的 PostModel
	// 注意：这里 CachedPostModel 和 PostModel 有相同的方法集
	// 但 fx 不支持直接替换类型，需要在 handler 中修改依赖
	injection.Provide(NewPostModel)
	injection.Provide(NewCategoryModel)
	injection.Provide(NewSheetModel)
	injection.Provide(NewTagModel)
	injection.Provide(NewLinkModel)
	injection.Provide(NewPhotoModel)
	injection.Provide(NewJournalModel)
}
