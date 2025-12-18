I will fix the compilation errors by addressing the missing imports, undefined methods, and context mismatches.

### Plan:

1.  **Fix `handler/admin/option.go`**:
    *   Add missing import `github.com/go-sonic/sonic/util`.

2.  **Fix `handler/admin/post.go`**:
    *   Update `PreviewPost` signature to return `error` (required for Fiber handlers).
    *   Replace undefined `ctx.Error(err)` with `return err`.

3.  **Fix `handler/admin/sheet.go`**:
    *   Update `PreviewSheet` signature to return `error`.
    *   Replace undefined `ctx.Error(err)` with `return err`.

4.  **Fix `handler/middleware/auth.go`**:
    *   Import `context`.
    *   Update middleware to propagate the authenticated user to `ctx.UserContext()` (standard `context.Context`) in addition to `ctx.Locals`. This ensures service-level functions like `GetAuthorizedUser` work correctly.

5.  **Fix `handler/admin/user.go`**:
    *   Update `impl.GetAuthorizedUser(ctx)` to `impl.GetAuthorizedUser(ctx.UserContext())` to match the function signature.

After these changes, I will run the build to verify the fixes.