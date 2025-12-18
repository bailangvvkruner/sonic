I will fix the syntax errors in `handler/admin/comment_journal.go`, `handler/admin/comment_post.go`, and `handler/admin/comment_sheet.go`. The previous search and replace operation introduced malformed `if` statements.

**1. Fix** **`handler/admin/comment_journal.go`**

* **Error**: `syntax error: unexpected keyword if at end of statement` (line 141) and `syntax error: unexpected { at end of statement` (line 169 - adjusted for new lines).

* **Cause**: The previous edit resulted in `}+ if err != nil {` which is syntactically incorrect. It should be `}` followed by `if err != nil {`.

* **Action**: Correct the `CreateJournalComment` function to properly close the `if` block for validation error handling.

**2. Fix** **`handler/admin/comment_post.go`**

* **Error**: `syntax error: unexpected keyword if at end of statement` (line 147) and `syntax error: unexpected { at end of statement` (line 175).

* **Cause**: Similar to above, `}+ if err != nil {`.

* **Action**: Correct the `CreatePostComment` function.

**3. Fix** **`handler/admin/comment_sheet.go`**

* **Error**: `syntax error: unexpected { at end of statement` (line 182 - roughly).

* **Cause**: Similar to above, `}+ if err != nil {` (likely around line 154 based on previous file read).

* **Action**: Correct the `CreateSheetComment` function.

**Plan Execution Order**:

1. Read the files to confirm the exact location of the syntax errors (already done via previous tool outputs, but good to be precise with search/replace).
2. Apply fixes to all three files.

