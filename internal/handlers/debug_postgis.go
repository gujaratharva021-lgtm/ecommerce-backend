package handlers

import (
"net/http"

"github.com/gin-gonic/gin"
"github.com/gujaratharva021-lgtm/ecommerce-backend/internal/database"
)

func DebugCheckPostGIS(c *gin.Context) {
if err := database.DB.Exec("CREATE EXTENSION IF NOT EXISTS postgis").Error; err != nil {
c.JSON(http.StatusInternalServerError, gin.H{"postgis_available": false, "error": err.Error()})
return
}
var version string
database.DB.Raw("SELECT PostGIS_Version()").Scan(&version)
c.JSON(http.StatusOK, gin.H{"postgis_available": true, "version": version})
}
