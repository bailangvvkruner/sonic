I will fix the remaining compilation errors by updating context usage in the admin handlers and adding a missing import.

**1. Fix Context Usage in Admin Handlers**
The `*fiber.Ctx` is being passed where `context.Context` is expected, or `UserContext()` is being called on a `context.Context` type (which doesn't have that method).

* **`handler/admin/backup.go`**:

  * Error: `log.CtxErrorf(ctx, ...)` where `ctx` is `*fiber.Ctx`.

  * Fix: Change to `log.CtxErrorf(ctx.UserContext(), ...)`.

* **`handler/admin/comment_journal.go`**:

  * Error: `impl.MustGetAuthorizedUser(ctx)` where `ctx` is `*fiber.Ctx`.

  * Fix: `impl.MustGetAuthorizedUser` likely expects `*fiber.Ctx` if it extracts user info from Fiber context, BUT the error says `cannot use ctx ... as "context".Context`. I need to check `impl.MustGetAuthorizedUser` signature. Wait, the previous errors for this were "ctx ... as context.Context". If `impl.MustGetAuthorizedUser` takes `context.Context`, I should pass `ctx.UserContext()`. I will assume it takes `context.Context` based on the error message.

  * Update: `impl.MustGetAuthorizedUser(ctx.UserContext())`.

* **`handler/admin/comment_post.go`**:

  * Error: `impl.MustGetAuthorizedUser(ctx)`.

  * Fix: `impl.MustGetAuthorizedUser(ctx.UserContext())`.

* **`handler/admin/comment_sheet.go`**:

  * Error 1: `s.ConvertToWithSheet(ctx, comments)` call in `ListSheetComment` and `ListSheetCommentLatest`.

  * Fix: `s.ConvertToWithSheet(ctx.UserContext(), comments)`.

  * Error 2: `impl.MustGetAuthorizedUser(ctx)`.

  * Fix: `impl.MustGetAuthorizedUser(ctx.UserContext())`.

  * Error 3: `ctx.UserContext()` usage inside `ConvertToWithSheet` where `ctx` is already `context.Context`.

  * Fix: Remove `.UserContext()` calls inside `ConvertToWithSheet`. Use `ctx` directly.

**2. Fix Missing Import in Install Handler**

* **`handler/admin/install.go`**:

  * Error: `undefined: util`.

  * Fix: Add `"github.com/go-sonic/sonic/util"` to imports.

**Plan Execution Order**:

1. Update `handler/admin/backup.go`.
2. Update `handler/admin/comment_journal.go`.
3. Update `handler/admin/comment_post.go`.
4. Update `handler/admin/comment_sheet.go`.
5. Update `handler/admin/install.go`.

