package handlers

import (
"net/http"

"github.com/gin-gonic/gin"
"github.com/gujaratharva021-lgtm/ecommerce-backend/internal/database"
)

// RunCartReservationMigration godoc
// GET /api/v1/run-cart-reservation-migration
//
// TEMPORARY, ONE-TIME USE ONLY. Creates the cart_reservations table in
// production (where AutoMigrate doesn't run). Safe to call more than once
// - every statement uses IF NOT EXISTS / matching guards. Delete this file
// and its route in routes.go right after running it once successfully.
func RunCartReservationMigration(c *gin.Context) {
statements := []string{
`CREATE TABLE IF NOT EXISTS public.cart_reservations (
id BIGSERIAL PRIMARY KEY,
user_id bigint NOT NULL,
product_id bigint NOT NULL,
warehouse_id bigint NOT NULL,
quantity integer NOT NULL,
expires_at timestamp with time zone NOT NULL,
created_at timestamp with time zone,
updated_at timestamp with time zone
)`,
`CREATE UNIQUE INDEX IF NOT EXISTS idx_reservation_user_product_wh
ON public.cart_reservations (user_id, product_id, warehouse_id)`,
`CREATE INDEX IF NOT EXISTS idx_reservation_product_wh
ON public.cart_reservations (product_id, warehouse_id)`,
`CREATE INDEX IF NOT EXISTS idx_cart_reservations_expires_at
ON public.cart_reservations (expires_at)`,
}

for _, stmt := range statements {
if err := database.DB.Exec(stmt).Error; err != nil {
c.JSON(http.StatusInternalServerError, gin.H{
"error":     "Migration step failed",
"statement": stmt,
"detail":    err.Error(),
})
return
}
}

fkStatements := []string{
`ALTER TABLE public.cart_reservations ADD CONSTRAINT fk_cart_reservations_user FOREIGN KEY (user_id) REFERENCES public.users(id)`,
`ALTER TABLE public.cart_reservations ADD CONSTRAINT fk_cart_reservations_product FOREIGN KEY (product_id) REFERENCES public.products(id)`,
`ALTER TABLE public.cart_reservations ADD CONSTRAINT fk_cart_reservations_warehouse FOREIGN KEY (warehouse_id) REFERENCES public.warehouses(id)`,
}
for _, stmt := range fkStatements {
database.DB.Exec(stmt)
}

c.JSON(http.StatusOK, gin.H{"status": "cart_reservations table ready"})
}
