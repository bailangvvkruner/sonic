I will fix the compilation errors by updating the context usage to be compatible with Fiber and removing leftover Gin references.

**1. Fix Context Compatibility in Content Handlers**
The `fiber.Ctx` is not compatible with `context.Context` directly. I will use `ctx.UserContext()` which returns a `context.Context`.

- **Files to modify**:
  - `handler/content/archive.go`
  - `handler/content/category.go`
  - `handler/content/index.go`
  - `handler/content/journal.go`
- **Action**: Replace `ctx` with `ctx.UserContext()` in calls to model methods (e.g., `a.PostModel.Archives(ctx.UserContext(), ...)`).

**2. Fix Context Usage in Admin Handler**
`ctx.UserContext()` already returns a `context.Context`, so calling `.UserContext()` on it again is incorrect.

- **File to modify**: `handler/admin/admin.go`
- **Action**: Replace `ctx.UserContext().UserContext()` with `ctx.UserContext()`.

**3. Fix Gin References in Backup Handler**
The `wrapHandler` function signature still refers to `gin.HandlerFunc`.

- **File to modify**: `handler/admin/backup.go`
- **Action**:
  - Change return type from `gin.HandlerFunc` to `fiber.Handler`.
  - Update the returned function signature to return `error` (Fiber style) instead of `void`.
  - Update implementation to return `ctx.JSON(...)` or `ctx.Status(...)...`.

**4. Fix Binding in Attachment Handler**
`ctx.ShouldBindWith` is a Gin method. Fiber uses `BodyParser` or `QueryParser`.

- **File to modify**: `handler/admin/attachment.go`
- **Action**: Replace `ctx.ShouldBindWith(queryParam, binding.CustomFormBinding)` with `ctx.QueryParser(queryParam)` since it handles query parameters for GET requests.
