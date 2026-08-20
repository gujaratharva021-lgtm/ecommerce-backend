package handlers

import (
"net/http"

"github.com/gin-gonic/gin"
"github.com/gujaratharva021-lgtm/ecommerce-backend/internal/database"
)

// TEMPORARY DEBUG ENDPOINT - remove after use.
// POST /api/v1/admin/debug/cleanup-test-credit-note
func DebugCleanupTestCreditNote(c *gin.Context) {
r1 := database.DB.Exec("DELETE FROM ledger_entries WHERE transaction_ref = ?", "CREDITNOTE-1")
r2 := database.DB.Exec("DELETE FROM credit_note_items WHERE credit_note_id = ?", 1)
r3 := database.DB.Exec("DELETE FROM credit_notes WHERE id = ?", 1)
c.JSON(http.StatusOK, gin.H{
"ledger_entries_deleted": r1.RowsAffected,
"credit_note_items_deleted": r2.RowsAffected,
"credit_notes_deleted": r3.RowsAffected,
"errors": gin.H{
"r1": r1.Error,
"r2": r2.Error,
"r3": r3.Error,
},
})
}
