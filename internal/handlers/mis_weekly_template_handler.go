package handlers

import (
"fmt"
"net/http"

"github.com/gin-gonic/gin"
"github.com/xuri/excelize/v2"
)

func GetWeeklyTemplateMIS(c *gin.Context) {
start, end := weekBounds(c.Query("week_start"))
prevStart := start.AddDate(0, 0, -7)
prevEnd := end.AddDate(0, 0, -7)
ytdStart := monthBoundsYearStart(start)
periodKey := start.Format("2006-01-02")

c.JSON(http.StatusOK, gin.H{
"week_start": start.Format("2006-01-02"),
"week_end":   end.Format("2006-01-02"),
"dashboard":  computeDashboardMonthlyMIS(periodKey, start, end, prevStart, prevEnd, ytdStart),
"grocery":    computeGroceryMonthlyMIS(periodKey, start, end, prevStart, prevEnd, ytdStart),
})
}

func ExportWeeklyTemplateMIS(c *gin.Context) {
start, end := weekBounds(c.Query("week_start"))
prevStart := start.AddDate(0, 0, -7)
prevEnd := end.AddDate(0, 0, -7)
ytdStart := monthBoundsYearStart(start)
periodKey := start.Format("2006-01-02")

xf := excelize.NewFile()
defer xf.Close()

dashSheet := "Dashboard"
xf.SetSheetName("Sheet1", dashSheet)
writeMonthlySheet(xf, dashSheet, computeDashboardMonthlyMIS(periodKey, start, end, prevStart, prevEnd, ytdStart))

grocerySheet := "Grocery"
xf.NewSheet(grocerySheet)
writeMonthlySheet(xf, grocerySheet, computeGroceryMonthlyMIS(periodKey, start, end, prevStart, prevEnd, ytdStart))

xf.SetActiveSheet(0)
filename := fmt.Sprintf("weekly-dashboard-mis-%s.xlsx", periodKey)
c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
if err := xf.Write(c.Writer); err != nil {
c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate Excel file"})
}
}
