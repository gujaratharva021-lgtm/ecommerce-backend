package services

import (
"fmt"
"time"

"github.com/gujaratharva021-lgtm/ecommerce-backend/internal/database"
"github.com/gujaratharva021-lgtm/ecommerce-backend/internal/models"
"gorm.io/gorm"
)

// PostSalesLedgerEntry records the double-entry ledger lines for one sale,
// triggered at the exact moment revenue is recognized:
//   - Online: right after payment verification (payment_handler.go)
//   - COD:    right after delivery confirmation (ConfirmDelivery)
// It is idempotent per order (checked via reference_type="sale",
// reference_id=orderID) so a retry or duplicate call never double-posts.
//
// Balanced entry:
//   Debit  Bank/Cash (order total actually collected)
//   Debit  Customer Wallet Liability (wallet amount redeemed)
//   Debit  Discount Given (discount amount)
//   Credit Product Sales (gross taxable amount, i.e. before discount)
//   Credit GST Payable (CGST+SGST+IGST, computed on the net/discounted value)
// Debit total  = TotalAmount + WalletUsed + DiscountAmount
// Credit total = (TaxableAmount + DiscountAmount) + totalGST
// These are equal because TaxableAmount + totalGST = TotalAmount + WalletUsed
// (invoice math), so adding DiscountAmount to both sides keeps it balanced.
// Product Sales is credited gross (pre-discount) and Discount Given is
// debited separately - crediting TaxableAmount alone (net/post-discount)
// would leave Debit exceeding Credit by exactly DiscountAmount on every
// discounted order.
// PostWalletTopupLedgerEntry records Debit Bank (1002), Credit Customer
// Wallet Liability (2005) for a Razorpay-verified wallet top-up. Real
// money lands in the bank via the payment gateway, and the corresponding
// increase in what the company owes the customer (their spendable wallet
// balance) must be recorded as a liability at the same moment - without
// this, the credit-wallet side of a top-up was invisible to the general
// ledger entirely, so Account 2005 only ever moved on the debit side
// (when a wallet balance is later spent/refunded), silently drifting
// negative over time (Defect #03).
// Idempotent per top-up via reference_type="wallet_topup", reference_id=topupID.
func PostWalletTopupLedgerEntry(topupID uint) error {
var existing models.LedgerEntry
if err := database.DB.Where("reference_type = ? AND reference_id = ?", "wallet_topup", topupID).First(&existing).Error; err == nil {
return nil
}

var topup models.WalletTopup
if err := database.DB.First(&topup, topupID).Error; err != nil {
return fmt.Errorf("wallet topup not found: %w", err)
}
if topup.Amount <= 0 {
return nil
}

var bank, walletLiability models.Account
if err := database.DB.Where("code = ?", "1002").First(&bank).Error; err != nil {
return fmt.Errorf("chart of accounts missing code 1002 (Bank): %w", err)
}
if err := database.DB.Where("code = ?", "2005").First(&walletLiability).Error; err != nil {
return fmt.Errorf("chart of accounts missing code 2005 (Customer Wallet Liability): %w", err)
}

transactionRef := fmt.Sprintf("WALLETTOPUP-%d", topupID)
now := time.Now()

return database.DB.Transaction(func(tx *gorm.DB) error {
debit := models.LedgerEntry{
TransactionRef: transactionRef,
AccountID:      bank.ID,
Type:           "debit",
Amount:         topup.Amount,
Description:    fmt.Sprintf("Wallet top-up #%d", topupID),
ReferenceType:  "wallet_topup",
ReferenceID:    &topupID,
EntryDate:      now,
}
if err := tx.Create(&debit).Error; err != nil {
return fmt.Errorf("failed to create debit ledger entry: %w", err)
}
credit := models.LedgerEntry{
TransactionRef: transactionRef,
AccountID:      walletLiability.ID,
Type:           "credit",
Amount:         topup.Amount,
Description:    fmt.Sprintf("Wallet top-up #%d", topupID),
ReferenceType:  "wallet_topup",
ReferenceID:    &topupID,
EntryDate:      now,
}
if err := tx.Create(&credit).Error; err != nil {
return fmt.Errorf("failed to create credit ledger entry: %w", err)
}
return nil
})
}

func PostSalesLedgerEntry(orderID uint) error {
var invoice models.Invoice
if err := database.DB.Where("order_id = ?", orderID).First(&invoice).Error; err != nil {
return fmt.Errorf("no invoice found for order %d: %w", orderID, err)
}

// Idempotency check.
var existing models.LedgerEntry
if err := database.DB.Where("reference_type = ? AND reference_id = ?", "sale", orderID).First(&existing).Error; err == nil {
return nil
}

var order models.Order
if err := database.DB.First(&order, orderID).Error; err != nil {
return fmt.Errorf("order not found: %w", err)
}

cashOrBankCode := "1002" // Bank
if order.PaymentMethod == models.PaymentMethodCOD {
cashOrBankCode = "1001" // Cash
}

totalGST := invoice.CGSTAmount + invoice.SGSTAmount + invoice.IGSTAmount
transactionRef := fmt.Sprintf("SALE-%d", orderID)

type line struct {
code   string
lType  string
amount float64
desc   string
}
var lines []line

lines = append(lines, line{cashOrBankCode, "debit", invoice.TotalAmount, fmt.Sprintf("Sale for order #%d", orderID)})
if invoice.WalletUsed > 0 {
lines = append(lines, line{"2005", "debit", invoice.WalletUsed, fmt.Sprintf("Wallet redeemed on order #%d", orderID)})
}
if invoice.DiscountAmount > 0 {
lines = append(lines, line{"5002", "debit", invoice.DiscountAmount, fmt.Sprintf("Discount on order #%d", orderID)})
}
lines = append(lines, line{"4001", "credit", invoice.TaxableAmount + invoice.DiscountAmount, fmt.Sprintf("Sale for order #%d", orderID)})
if invoice.DeliveryCharge > 0 {
lines = append(lines, line{"4002", "credit", invoice.DeliveryCharge, fmt.Sprintf("Delivery fee for order #%d", orderID)})
}
if invoice.PlatformFee > 0 {
lines = append(lines, line{"4003", "credit", invoice.PlatformFee, fmt.Sprintf("Platform fee for order #%d", orderID)})
}
if totalGST > 0 {
lines = append(lines, line{"2002", "credit", totalGST, fmt.Sprintf("GST on order #%d", orderID)})
}

return database.DB.Transaction(func(tx *gorm.DB) error {
for _, l := range lines {
if l.amount <= 0 {
continue
}
var account models.Account
if err := tx.Where("code = ?", l.code).First(&account).Error; err != nil {
return fmt.Errorf("chart of accounts missing code %s: %w", l.code, err)
}
entry := models.LedgerEntry{
TransactionRef: transactionRef,
AccountID:      account.ID,
Type:           l.lType,
Amount:         l.amount,
Description:    l.desc,
ReferenceType:  "sale",
ReferenceID:    &orderID,
EntryDate:      invoice.GeneratedAt,
CreatedByID:    nil, // system-generated, same convention as invoice audit log
}
if err := tx.Create(&entry).Error; err != nil {
return fmt.Errorf("failed to create ledger entry: %w", err)
}
}
return nil
})
}

// PostVendorPaymentLedgerEntry records the double-entry ledger lines for one
// vendor bill payment: Debit Vendor Payable, Credit Bank. Each call to
// PayVendorBill can record a partial payment, so this posts once per
// payment (not once per bill) - the transaction ref includes a timestamp
// so multiple payments against the same bill each get their own entry.
func PostVendorPaymentLedgerEntry(billID uint, amount float64) error {
if amount <= 0 {
return nil
}

var vendorPayable, bank models.Account
if err := database.DB.Where("code = ?", "2001").First(&vendorPayable).Error; err != nil {
return fmt.Errorf("chart of accounts missing code 2001 (Vendor Payable): %w", err)
}
if err := database.DB.Where("code = ?", "1002").First(&bank).Error; err != nil {
return fmt.Errorf("chart of accounts missing code 1002 (Bank): %w", err)
}

transactionRef := fmt.Sprintf("VBILLPAY-%d-%d", billID, time.Now().UnixNano())
now := time.Now()

return database.DB.Transaction(func(tx *gorm.DB) error {
debit := models.LedgerEntry{
TransactionRef: transactionRef,
AccountID:      vendorPayable.ID,
Type:           "debit",
Amount:         amount,
Description:    fmt.Sprintf("Payment against vendor bill #%d", billID),
ReferenceType:  "vendor_bill_payment",
ReferenceID:    &billID,
EntryDate:      now,
}
if err := tx.Create(&debit).Error; err != nil {
return fmt.Errorf("failed to create debit ledger entry: %w", err)
}
credit := models.LedgerEntry{
TransactionRef: transactionRef,
AccountID:      bank.ID,
Type:           "credit",
Amount:         amount,
Description:    fmt.Sprintf("Payment against vendor bill #%d", billID),
ReferenceType:  "vendor_bill_payment",
ReferenceID:    &billID,
EntryDate:      now,
}
if err := tx.Create(&credit).Error; err != nil {
return fmt.Errorf("failed to create credit ledger entry: %w", err)
}
return nil
})
}


// PostVendorBillLedgerEntry records the double-entry ledger lines for one
// vendor bill at creation time: Debit Inventory for the goods-value portion,
// Debit GST Input Credit (ITC) for the GST portion (claiming the input
// credit on this purchase), and Credit Vendor Payable for the full amount
// owed (amount + GST). This is the mirror image of PostDebitNoteLedgerEntry,
// which reverses these same three legs when goods are returned to a vendor.
//
// Idempotent: if an entry already exists for this bill, it's a no-op, so
// this is safe to call from CreateVendorBill without a separate "posted"
// flag on VendorBill.
func PostVendorBillLedgerEntry(billID uint) error {
var existing models.LedgerEntry
if err := database.DB.Where("reference_type = ? AND reference_id = ?", "vendor_bill", billID).First(&existing).Error; err == nil {
return nil
}

var bill models.VendorBill
if err := database.DB.First(&bill, billID).Error; err != nil {
return fmt.Errorf("vendor bill not found: %w", err)
}

var vendorPayable, inventory models.Account
if err := database.DB.Where("code = ?", "2001").First(&vendorPayable).Error; err != nil {
return fmt.Errorf("chart of accounts missing code 2001 (Vendor Payable): %w", err)
}
if err := database.DB.Where("code = ?", "1004").First(&inventory).Error; err != nil {
return fmt.Errorf("chart of accounts missing code 1004 (Inventory): %w", err)
}

transactionRef := fmt.Sprintf("VENDORBILL-%d", billID)

return database.DB.Transaction(func(tx *gorm.DB) error {
debitInventory := models.LedgerEntry{
TransactionRef: transactionRef,
AccountID:      inventory.ID,
Type:           "debit",
Amount:         bill.Amount,
Description:    fmt.Sprintf("Vendor bill %s", bill.BillNumber),
ReferenceType:  "vendor_bill",
ReferenceID:    &billID,
EntryDate:      bill.BillDate,
}
if err := tx.Create(&debitInventory).Error; err != nil {
return fmt.Errorf("failed to create inventory ledger entry: %w", err)
}

if bill.GSTAmount > 0 {
var gstITC models.Account
if err := tx.Where("code = ?", "1005").First(&gstITC).Error; err != nil {
return fmt.Errorf("chart of accounts missing code 1005 (GST Input Credit): %w", err)
}
debitGST := models.LedgerEntry{
TransactionRef: transactionRef,
AccountID:      gstITC.ID,
Type:           "debit",
Amount:         bill.GSTAmount,
Description:    fmt.Sprintf("Vendor bill %s - GST ITC", bill.BillNumber),
ReferenceType:  "vendor_bill",
ReferenceID:    &billID,
EntryDate:      bill.BillDate,
}
if err := tx.Create(&debitGST).Error; err != nil {
return fmt.Errorf("failed to create GST-ITC ledger entry: %w", err)
}
}

creditVendorPayable := models.LedgerEntry{
TransactionRef: transactionRef,
AccountID:      vendorPayable.ID,
Type:           "credit",
Amount:         bill.Amount + bill.GSTAmount,
Description:    fmt.Sprintf("Vendor bill %s", bill.BillNumber),
ReferenceType:  "vendor_bill",
ReferenceID:    &billID,
EntryDate:      bill.BillDate,
}
if err := tx.Create(&creditVendorPayable).Error; err != nil {
return fmt.Errorf("failed to create vendor-payable ledger entry: %w", err)
}

return nil
})
}

// ReverseVendorBillLedgerEntry reverses the ledger entries created by
// PostVendorBillLedgerEntry when a vendor bill is voided: Credit Inventory,
// Credit GST Input Credit (if any), Debit Vendor Payable - the exact mirror
// of the original posting. Idempotent per bill via reference_type="vendor_bill_void".
// No-op if the original bill was never posted to the ledger in the first place.
func ReverseVendorBillLedgerEntry(billID uint) error {
var existing models.LedgerEntry
if err := database.DB.Where("reference_type = ? AND reference_id = ?", "vendor_bill_void", billID).First(&existing).Error; err == nil {
return nil
}

var original models.LedgerEntry
if err := database.DB.Where("reference_type = ? AND reference_id = ?", "vendor_bill", billID).First(&original).Error; err != nil {
// Bill was never posted to the ledger (e.g. void before posting ran) - nothing to reverse.
return nil
}

var bill models.VendorBill
if err := database.DB.First(&bill, billID).Error; err != nil {
return fmt.Errorf("vendor bill not found: %w", err)
}

var vendorPayable, inventory models.Account
if err := database.DB.Where("code = ?", "2001").First(&vendorPayable).Error; err != nil {
return fmt.Errorf("chart of accounts missing code 2001 (Vendor Payable): %w", err)
}
if err := database.DB.Where("code = ?", "1004").First(&inventory).Error; err != nil {
return fmt.Errorf("chart of accounts missing code 1004 (Inventory): %w", err)
}

transactionRef := fmt.Sprintf("VENDORBILLVOID-%d", billID)
now := time.Now()

return database.DB.Transaction(func(tx *gorm.DB) error {
creditInventory := models.LedgerEntry{
TransactionRef: transactionRef,
AccountID:      inventory.ID,
Type:           "credit",
Amount:         bill.Amount,
Description:    fmt.Sprintf("Void vendor bill %s", bill.BillNumber),
ReferenceType:  "vendor_bill_void",
ReferenceID:    &billID,
EntryDate:      now,
}
if err := tx.Create(&creditInventory).Error; err != nil {
return fmt.Errorf("failed to create inventory reversal ledger entry: %w", err)
}

if bill.GSTAmount > 0 {
var gstITC models.Account
if err := tx.Where("code = ?", "1005").First(&gstITC).Error; err != nil {
return fmt.Errorf("chart of accounts missing code 1005 (GST Input Credit): %w", err)
}
creditGST := models.LedgerEntry{
TransactionRef: transactionRef,
AccountID:      gstITC.ID,
Type:           "credit",
Amount:         bill.GSTAmount,
Description:    fmt.Sprintf("Void vendor bill %s - GST ITC reversal", bill.BillNumber),
ReferenceType:  "vendor_bill_void",
ReferenceID:    &billID,
EntryDate:      now,
}
if err := tx.Create(&creditGST).Error; err != nil {
return fmt.Errorf("failed to create GST-ITC reversal ledger entry: %w", err)
}
}

debitVendorPayable := models.LedgerEntry{
TransactionRef: transactionRef,
AccountID:      vendorPayable.ID,
Type:           "debit",
Amount:         bill.Amount + bill.GSTAmount,
Description:    fmt.Sprintf("Void vendor bill %s", bill.BillNumber),
ReferenceType:  "vendor_bill_void",
ReferenceID:    &billID,
EntryDate:      now,
}
if err := tx.Create(&debitVendorPayable).Error; err != nil {
return fmt.Errorf("failed to create vendor-payable reversal ledger entry: %w", err)
}

return nil
})
}
// PostExpenseLedgerEntry records the double-entry ledger lines for one
// expense at creation time: Debit Operating Expenses, Credit Bank. This
// posts once, at CreateExpense - if the expense amount is later edited via
// UpdateExpense, the ledger is NOT automatically corrected (that requires a
// proper reversal/adjustment entry, which is out of scope for this phase;
// tracked separately). Idempotent per expense via reference_type="expense".
// PostPayrollLedgerEntry records Debit Salary Expense (5006), Credit
// Cash/Bank (1001/1002 depending on payroll.PaymentMethod - "cash" maps
// to Cash, "bank"/"upi" both settle through the bank account) for a
// staff salary payment. Idempotency-checked the same way as
// PostExpenseLedgerEntry: a payroll record can only ever post once,
// keyed on reference_type="payroll", reference_id=payrollID (Bug
// FINANCE-07: payroll payments previously never touched the ledger at
// all, silently understating total costs on the P&L and trial balance).
// PostPayrollLedgerEntry takes an existing transaction rather than opening
// its own (Defect #12) - the caller (UpdatePayroll/CreatePayroll) locks the
// payroll row for the duration of its own transaction, and the idempotency
// check + insert here must run inside that same lock. Calling this with
// database.DB directly (no active lock held) would reopen the exact race
// this fix closes: two concurrent requests could both pass the idempotency
// check before either commits, producing duplicate salary debits and bank
// credits in the general ledger.
func PostPayrollLedgerEntry(tx *gorm.DB, payrollID uint) error {
var existing models.LedgerEntry
if err := tx.Where("reference_type = ? AND reference_id = ?", "payroll", payrollID).First(&existing).Error; err == nil {
return nil
}

var payroll models.Payroll
if err := tx.Preload("Staff").First(&payroll, payrollID).Error; err != nil {
return fmt.Errorf("payroll record not found: %w", err)
}
if payroll.Amount <= 0 {
return nil
}

cashOrBankCode := "1002" // Bank (also used for "upi", which settles via bank)
if payroll.PaymentMethod == "cash" {
cashOrBankCode = "1001"
}

var salaryExpense, cashOrBank models.Account
if err := tx.Where("code = ?", "5006").First(&salaryExpense).Error; err != nil {
return fmt.Errorf("chart of accounts missing code 5006 (Salary Expense): %w", err)
}
if err := tx.Where("code = ?", cashOrBankCode).First(&cashOrBank).Error; err != nil {
return fmt.Errorf("chart of accounts missing code %s: %w", cashOrBankCode, err)
}

transactionRef := fmt.Sprintf("PAYROLL-%d", payrollID)
staffName := payroll.Staff.Name
now := time.Now()

debit := models.LedgerEntry{
TransactionRef: transactionRef,
AccountID:      salaryExpense.ID,
Type:           "debit",
Amount:         payroll.Amount,
Description:    fmt.Sprintf("Salary paid to %s for %d/%d", staffName, payroll.Month, payroll.Year),
ReferenceType:  "payroll",
ReferenceID:    &payrollID,
EntryDate:      now,
}
if err := tx.Create(&debit).Error; err != nil {
return fmt.Errorf("failed to create debit ledger entry: %w", err)
}
credit := models.LedgerEntry{
TransactionRef: transactionRef,
AccountID:      cashOrBank.ID,
Type:           "credit",
Amount:         payroll.Amount,
Description:    fmt.Sprintf("Salary paid to %s for %d/%d", staffName, payroll.Month, payroll.Year),
ReferenceType:  "payroll",
ReferenceID:    &payrollID,
EntryDate:      now,
}
if err := tx.Create(&credit).Error; err != nil {
return fmt.Errorf("failed to create credit ledger entry: %w", err)
}
return nil
}

func PostExpenseLedgerEntry(expenseID uint) error {
var existing models.LedgerEntry
if err := database.DB.Where("reference_type = ? AND reference_id = ?", "expense", expenseID).First(&existing).Error; err == nil {
return nil
}

var expense models.Expense
if err := database.DB.First(&expense, expenseID).Error; err != nil {
return fmt.Errorf("expense not found: %w", err)
}
if expense.Amount <= 0 {
return nil
}

var opex, bank models.Account
if err := database.DB.Where("code = ?", "5003").First(&opex).Error; err != nil {
return fmt.Errorf("chart of accounts missing code 5003 (Operating Expenses): %w", err)
}
if err := database.DB.Where("code = ?", "1002").First(&bank).Error; err != nil {
return fmt.Errorf("chart of accounts missing code 1002 (Bank): %w", err)
}

transactionRef := fmt.Sprintf("EXPENSE-%d", expenseID)

return database.DB.Transaction(func(tx *gorm.DB) error {
debit := models.LedgerEntry{
TransactionRef: transactionRef,
AccountID:      opex.ID,
Type:           "debit",
Amount:         expense.Amount,
Description:    fmt.Sprintf("Expense: %s", expense.Category),
ReferenceType:  "expense",
ReferenceID:    &expenseID,
EntryDate:      expense.ExpenseDate,
}
if err := tx.Create(&debit).Error; err != nil {
return fmt.Errorf("failed to create debit ledger entry: %w", err)
}
credit := models.LedgerEntry{
TransactionRef: transactionRef,
AccountID:      bank.ID,
Type:           "credit",
Amount:         expense.Amount,
Description:    fmt.Sprintf("Expense: %s", expense.Category),
ReferenceType:  "expense",
ReferenceID:    &expenseID,
EntryDate:      expense.ExpenseDate,
}
if err := tx.Create(&credit).Error; err != nil {
return fmt.Errorf("failed to create credit ledger entry: %w", err)
}
return nil
})
}

// PostExpenseAdjustmentLedgerEntry posts an adjusting double-entry line
// when a paid expense's amount is edited after PostExpenseLedgerEntry has
// already run for it, so the general ledger stays in sync with the
// Expense record instead of going stale. deltaAmount is
// newAmount-oldAmount: positive means the expense grew (post an extra
// Debit Opex / Credit Bank for the increase), negative means it shrank
// (post the reverse - Debit Bank / Credit Opex - for the decrease).
// No-op if deltaAmount is zero.
// PostPayrollAdjustmentLedgerEntry posts a compensating ledger entry
// when a PAID payroll record's amount is corrected after the fact -
// mirrors PostExpenseAdjustmentLedgerEntry's increase/decrease-direction
// swap exactly. Debit Salary Expense (5006) / Credit Cash-or-Bank for an
// increase; reversed for a decrease. Without this, editing a paid
// payroll's amount left the general ledger permanently reflecting the
// old, now-incorrect figure (Bug #09).
// PostPayrollAdjustmentLedgerEntry takes an existing transaction rather
// than opening its own (Defect #12), for the same reason as
// PostPayrollLedgerEntry above - this runs inside UpdatePayroll's locked
// transaction so amount-correction adjustments can't race either.
func PostPayrollAdjustmentLedgerEntry(tx *gorm.DB, payrollID uint, deltaAmount float64) error {
if deltaAmount == 0 {
return nil
}

var payroll models.Payroll
if err := tx.First(&payroll, payrollID).Error; err != nil {
return fmt.Errorf("payroll record not found: %w", err)
}

cashOrBankCode := "1002" // Bank (also used for "upi", which settles via bank)
if payroll.PaymentMethod == "cash" {
cashOrBankCode = "1001"
}

var salaryExpense, cashOrBank models.Account
if err := tx.Where("code = ?", "5006").First(&salaryExpense).Error; err != nil {
return fmt.Errorf("chart of accounts missing code 5006 (Salary Expense): %w", err)
}
if err := tx.Where("code = ?", cashOrBankCode).First(&cashOrBank).Error; err != nil {
return fmt.Errorf("chart of accounts missing code %s: %w", cashOrBankCode, err)
}

amount := deltaAmount
increased := true
if amount < 0 {
amount = -amount
increased = false
}

transactionRef := fmt.Sprintf("PAYROLLADJ-%d-%d", payrollID, time.Now().UnixNano())
now := time.Now()

salaryEntry := models.LedgerEntry{
TransactionRef: transactionRef,
AccountID:      salaryExpense.ID,
Type:           "debit",
Amount:         amount,
Description:    fmt.Sprintf("Salary adjustment for payroll #%d", payrollID),
ReferenceType:  "payroll_adjustment",
ReferenceID:    &payrollID,
EntryDate:      now,
}
cashEntry := models.LedgerEntry{
TransactionRef: transactionRef,
AccountID:      cashOrBank.ID,
Type:           "credit",
Amount:         amount,
Description:    fmt.Sprintf("Salary adjustment for payroll #%d", payrollID),
ReferenceType:  "payroll_adjustment",
ReferenceID:    &payrollID,
EntryDate:      now,
}
if !increased {
// Payroll amount decreased - reverse the direction: Debit Cash/Bank, Credit Salary Expense.
salaryEntry.Type = "credit"
cashEntry.Type = "debit"
}
if err := tx.Create(&salaryEntry).Error; err != nil {
return fmt.Errorf("failed to create salary adjustment ledger entry: %w", err)
}
if err := tx.Create(&cashEntry).Error; err != nil {
return fmt.Errorf("failed to create cash/bank adjustment ledger entry: %w", err)
}
return nil
}

func PostExpenseAdjustmentLedgerEntry(expenseID uint, deltaAmount float64) error {
if deltaAmount == 0 {
return nil
}

var expense models.Expense
if err := database.DB.First(&expense, expenseID).Error; err != nil {
return fmt.Errorf("expense not found: %w", err)
}

var opex, bank models.Account
if err := database.DB.Where("code = ?", "5003").First(&opex).Error; err != nil {
return fmt.Errorf("chart of accounts missing code 5003 (Operating Expenses): %w", err)
}
if err := database.DB.Where("code = ?", "1002").First(&bank).Error; err != nil {
return fmt.Errorf("chart of accounts missing code 1002 (Bank): %w", err)
}

amount := deltaAmount
increased := true
if amount < 0 {
amount = -amount
increased = false
}

transactionRef := fmt.Sprintf("EXPENSEADJ-%d-%d", expenseID, time.Now().UnixNano())
now := time.Now()

return database.DB.Transaction(func(tx *gorm.DB) error {
opexEntry := models.LedgerEntry{
TransactionRef: transactionRef,
AccountID:      opex.ID,
Type:           "debit",
Amount:         amount,
Description:    fmt.Sprintf("Adjustment for expense: %s", expense.Category),
ReferenceType:  "expense_adjustment",
ReferenceID:    &expenseID,
EntryDate:      now,
}
bankEntry := models.LedgerEntry{
TransactionRef: transactionRef,
AccountID:      bank.ID,
Type:           "credit",
Amount:         amount,
Description:    fmt.Sprintf("Adjustment for expense: %s", expense.Category),
ReferenceType:  "expense_adjustment",
ReferenceID:    &expenseID,
EntryDate:      now,
}
if !increased {
// Expense amount decreased - reverse the direction: Debit Bank, Credit Opex.
opexEntry.Type = "credit"
bankEntry.Type = "debit"
}
if err := tx.Create(&opexEntry).Error; err != nil {
return fmt.Errorf("failed to create opex adjustment ledger entry: %w", err)
}
if err := tx.Create(&bankEntry).Error; err != nil {
return fmt.Errorf("failed to create bank adjustment ledger entry: %w", err)
}
return nil
})
}

// PostRefundLedgerEntry records the double-entry ledger lines for one
// customer refund: Debit Customer Refund Payable, Credit Bank/Cash. Called
// with only the newly-added refund delta (not the cumulative refunded
// total), so a payment that moves partially_refunded -> refunded twice
// posts two separate entries, one per incremental amount actually paid
// out - never a single entry for the full cumulative refund.
// PostGatewayCaptureRefundLedgerEntry records Debit Bank (1002), Credit
// Customer Wallet Liability (2005) for the specific case where a Razorpay
// payment was captured for an order that had already been cancelled/returned
// while the customer was mid-checkout on the gateway page (VerifyPayment).
// Real money landed in the bank via the gateway, and it was immediately
// credited back to the customer as wallet balance rather than kept as
// revenue - both sides of that must be recorded, or the bank inflow is
// invisible to the ledger while the wallet liability increase has no
// matching debit anywhere (Defect #04). Distinct from PostRefundLedgerEntry
// (which reverses an EXISTING sale entry via Refund Payable/Bank outflow) -
// this case never became a sale in the first place, so there is nothing to
// reverse; it is closer in shape to PostWalletTopupLedgerEntry, just keyed
// off the order/payment instead of a top-up row.
// Idempotent per payment via reference_type="payment_capture_refund", reference_id=paymentID.
func PostGatewayCaptureRefundLedgerEntry(paymentID, orderID uint, amount float64) error {
if amount <= 0 {
return nil
}

var existing models.LedgerEntry
if err := database.DB.Where("reference_type = ? AND reference_id = ?", "payment_capture_refund", paymentID).First(&existing).Error; err == nil {
return nil
}

var bank, walletLiability models.Account
if err := database.DB.Where("code = ?", "1002").First(&bank).Error; err != nil {
return fmt.Errorf("chart of accounts missing code 1002 (Bank): %w", err)
}
if err := database.DB.Where("code = ?", "2005").First(&walletLiability).Error; err != nil {
return fmt.Errorf("chart of accounts missing code 2005 (Customer Wallet Liability): %w", err)
}

transactionRef := fmt.Sprintf("PAYCAPTUREREFUND-%d", paymentID)
now := time.Now()

return database.DB.Transaction(func(tx *gorm.DB) error {
debit := models.LedgerEntry{
TransactionRef: transactionRef,
AccountID:      bank.ID,
Type:           "debit",
Amount:         amount,
Description:    fmt.Sprintf("Payment captured after order #%d was cancelled - refunded to wallet", orderID),
ReferenceType:  "payment_capture_refund",
ReferenceID:    &paymentID,
EntryDate:      now,
}
if err := tx.Create(&debit).Error; err != nil {
return fmt.Errorf("failed to create debit ledger entry: %w", err)
}
credit := models.LedgerEntry{
TransactionRef: transactionRef,
AccountID:      walletLiability.ID,
Type:           "credit",
Amount:         amount,
Description:    fmt.Sprintf("Payment captured after order #%d was cancelled - refunded to wallet", orderID),
ReferenceType:  "payment_capture_refund",
ReferenceID:    &paymentID,
EntryDate:      now,
}
if err := tx.Create(&credit).Error; err != nil {
return fmt.Errorf("failed to create credit ledger entry: %w", err)
}
return nil
})
}

func PostRefundLedgerEntry(orderID uint, deltaAmount float64, paymentMethod string) error {
if deltaAmount <= 0 {
return nil
}

bankCode := "1002" // Bank
if paymentMethod == models.PaymentMethodCOD {
bankCode = "1001" // Cash
}

var refundPayable, bank models.Account
if err := database.DB.Where("code = ?", "2004").First(&refundPayable).Error; err != nil {
return fmt.Errorf("chart of accounts missing code 2004 (Customer Refund Payable): %w", err)
}
if err := database.DB.Where("code = ?", bankCode).First(&bank).Error; err != nil {
return fmt.Errorf("chart of accounts missing code %s: %w", bankCode, err)
}

transactionRef := fmt.Sprintf("REFUND-%d-%d", orderID, time.Now().UnixNano())
now := time.Now()

return database.DB.Transaction(func(tx *gorm.DB) error {
debit := models.LedgerEntry{
TransactionRef: transactionRef,
AccountID:      refundPayable.ID,
Type:           "debit",
Amount:         deltaAmount,
Description:    fmt.Sprintf("Refund for order #%d", orderID),
ReferenceType:  "refund",
ReferenceID:    &orderID,
EntryDate:      now,
}
if err := tx.Create(&debit).Error; err != nil {
return fmt.Errorf("failed to create debit ledger entry: %w", err)
}
credit := models.LedgerEntry{
TransactionRef: transactionRef,
AccountID:      bank.ID,
Type:           "credit",
Amount:         deltaAmount,
Description:    fmt.Sprintf("Refund for order #%d", orderID),
ReferenceType:  "refund",
ReferenceID:    &orderID,
EntryDate:      now,
}
if err := tx.Create(&credit).Error; err != nil {
return fmt.Errorf("failed to create credit ledger entry: %w", err)
}
return nil
})
}

// PostWalletRefundLedgerEntry books a refund that was credited to the
// customer's in-app wallet rather than paid out via bank/cash. No money
// actually leaves the business, so unlike PostRefundLedgerEntry (which
// credits Bank/Cash), this credits Customer Wallet Liability (2005) -
// the business now owes the customer that amount as store credit.
func PostWalletRefundLedgerEntry(orderID uint, deltaAmount float64) error {
if deltaAmount <= 0 {
return nil
}

var refundPayable, walletLiability models.Account
if err := database.DB.Where("code = ?", "2004").First(&refundPayable).Error; err != nil {
return fmt.Errorf("chart of accounts missing code 2004 (Customer Refund Payable): %w", err)
}
if err := database.DB.Where("code = ?", "2005").First(&walletLiability).Error; err != nil {
return fmt.Errorf("chart of accounts missing code 2005 (Customer Wallet Liability): %w", err)
}

transactionRef := fmt.Sprintf("WALLET-REFUND-%d-%d", orderID, time.Now().UnixNano())
now := time.Now()

return database.DB.Transaction(func(tx *gorm.DB) error {
debit := models.LedgerEntry{
TransactionRef: transactionRef,
AccountID:      refundPayable.ID,
Type:           "debit",
Amount:         deltaAmount,
Description:    fmt.Sprintf("Refund to wallet for order #%d", orderID),
ReferenceType:  "refund",
ReferenceID:    &orderID,
EntryDate:      now,
}
if err := tx.Create(&debit).Error; err != nil {
return fmt.Errorf("failed to create debit ledger entry: %w", err)
}
credit := models.LedgerEntry{
TransactionRef: transactionRef,
AccountID:      walletLiability.ID,
Type:           "credit",
Amount:         deltaAmount,
Description:    fmt.Sprintf("Refund to wallet for order #%d", orderID),
ReferenceType:  "refund",
ReferenceID:    &orderID,
EntryDate:      now,
}
if err := tx.Create(&credit).Error; err != nil {
return fmt.Errorf("failed to create credit ledger entry: %w", err)
}
return nil
})
}

// PostGatewaySettlementLedgerEntry records the gateway fee deduction for
// one online payment (SRS 12.10): Debit Gateway Fees (expense), Credit
// Bank - the fee is deducted from what the gateway actually pays out, so
// this reduces the Bank balance already recorded at sale time rather than
// re-booking the whole gross amount.
func PostGatewaySettlementLedgerEntry(paymentID uint, feeAmount float64) error {
if feeAmount <= 0 {
return nil
}
var existing models.LedgerEntry
if err := database.DB.Where("reference_type = ? AND reference_id = ?", "gateway_settlement", paymentID).First(&existing).Error; err == nil {
return nil
}

var gatewayFees, bank models.Account
if err := database.DB.Where("code = ?", "5005").First(&gatewayFees).Error; err != nil {
return fmt.Errorf("chart of accounts missing code 5005 (Gateway Fees): %w", err)
}
if err := database.DB.Where("code = ?", "1002").First(&bank).Error; err != nil {
return fmt.Errorf("chart of accounts missing code 1002 (Bank): %w", err)
}

transactionRef := fmt.Sprintf("GWSETTLE-%d", paymentID)
now := time.Now()

return database.DB.Transaction(func(tx *gorm.DB) error {
debit := models.LedgerEntry{
TransactionRef: transactionRef, AccountID: gatewayFees.ID, Type: "debit", Amount: feeAmount,
Description: fmt.Sprintf("Gateway fee for payment #%d", paymentID),
ReferenceType: "gateway_settlement", ReferenceID: &paymentID, EntryDate: now,
}
if err := tx.Create(&debit).Error; err != nil {
return fmt.Errorf("failed to create gateway fee ledger entry: %w", err)
}
credit := models.LedgerEntry{
TransactionRef: transactionRef, AccountID: bank.ID, Type: "credit", Amount: feeAmount,
Description: fmt.Sprintf("Gateway fee for payment #%d", paymentID),
ReferenceType: "gateway_settlement", ReferenceID: &paymentID, EntryDate: now,
}
if err := tx.Create(&credit).Error; err != nil {
return fmt.Errorf("failed to create bank ledger entry: %w", err)
}
return nil
})
}

// PostRiderPayoutAccrualLedgerEntry recognizes a rider payout as owed
// (SRS 12.11), the moment it's approved: Debit Rider Delivery Expense,
// Credit Rider Payable.
// PostRiderPayoutAccrualLedgerEntry takes an existing transaction rather
// than opening its own (Defect #12) - the caller (ApproveRiderPayout) locks
// the payout row for the duration of its own transaction, and the
// idempotency check + insert here must run inside that same lock. Without
// this, two concurrent approve requests on the same payout could both pass
// the "pending" status check before either commits, both flip it to
// "approved", and both post a separate accrual entry to the ledger -
// double-counting the rider's owed amount.
func PostRiderPayoutAccrualLedgerEntry(tx *gorm.DB, payoutID uint) error {
var existing models.LedgerEntry
if err := tx.Where("reference_type = ? AND reference_id = ?", "rider_payout_accrual", payoutID).First(&existing).Error; err == nil {
return nil
}

var payout models.RiderPayout
if err := tx.First(&payout, payoutID).Error; err != nil {
return fmt.Errorf("rider payout not found: %w", err)
}
if payout.Amount <= 0 {
return nil
}

var expense, payable models.Account
if err := tx.Where("code = ?", "5004").First(&expense).Error; err != nil {
return fmt.Errorf("chart of accounts missing code 5004 (Rider Delivery Expense): %w", err)
}
if err := tx.Where("code = ?", "2003").First(&payable).Error; err != nil {
return fmt.Errorf("chart of accounts missing code 2003 (Rider Payable): %w", err)
}

transactionRef := fmt.Sprintf("RIDERACCRUAL-%d", payoutID)
now := time.Now()

debit := models.LedgerEntry{
TransactionRef: transactionRef, AccountID: expense.ID, Type: "debit", Amount: payout.Amount,
Description: fmt.Sprintf("Rider payout #%d accrual", payoutID),
ReferenceType: "rider_payout_accrual", ReferenceID: &payoutID, EntryDate: now,
}
if err := tx.Create(&debit).Error; err != nil {
return fmt.Errorf("failed to create rider expense ledger entry: %w", err)
}
credit := models.LedgerEntry{
TransactionRef: transactionRef, AccountID: payable.ID, Type: "credit", Amount: payout.Amount,
Description: fmt.Sprintf("Rider payout #%d accrual", payoutID),
ReferenceType: "rider_payout_accrual", ReferenceID: &payoutID, EntryDate: now,
}
if err := tx.Create(&credit).Error; err != nil {
return fmt.Errorf("failed to create rider payable ledger entry: %w", err)
}
return nil
}

// PostRiderPayoutSettlementLedgerEntry records actually paying out a rider
// (SRS 12.11): Debit Rider Payable, Credit Bank. Only meaningful after
// PostRiderPayoutAccrualLedgerEntry has already run for the same payout.
// PostRiderPayoutSettlementLedgerEntry takes an existing transaction rather
// than opening its own (Defect #12), for the same reason as
// PostRiderPayoutAccrualLedgerEntry above - runs inside PayRiderPayout's
// locked transaction so a concurrent double-settlement can't slip past the
// "approved" status check either.
func PostRiderPayoutSettlementLedgerEntry(tx *gorm.DB, payoutID uint) error {
var existing models.LedgerEntry
if err := tx.Where("reference_type = ? AND reference_id = ?", "rider_payout_settlement", payoutID).First(&existing).Error; err == nil {
return nil
}

var payout models.RiderPayout
if err := tx.First(&payout, payoutID).Error; err != nil {
return fmt.Errorf("rider payout not found: %w", err)
}
if payout.Amount <= 0 {
return nil
}

var payable, bank models.Account
if err := tx.Where("code = ?", "2003").First(&payable).Error; err != nil {
return fmt.Errorf("chart of accounts missing code 2003 (Rider Payable): %w", err)
}
if err := tx.Where("code = ?", "1002").First(&bank).Error; err != nil {
return fmt.Errorf("chart of accounts missing code 1002 (Bank): %w", err)
}

transactionRef := fmt.Sprintf("RIDERPAY-%d", payoutID)
now := time.Now()

debit := models.LedgerEntry{
TransactionRef: transactionRef, AccountID: payable.ID, Type: "debit", Amount: payout.Amount,
Description: fmt.Sprintf("Rider payout #%d settlement", payoutID),
ReferenceType: "rider_payout_settlement", ReferenceID: &payoutID, EntryDate: now,
}
if err := tx.Create(&debit).Error; err != nil {
return fmt.Errorf("failed to create rider payable ledger entry: %w", err)
}
credit := models.LedgerEntry{
TransactionRef: transactionRef, AccountID: bank.ID, Type: "credit", Amount: payout.Amount,
Description: fmt.Sprintf("Rider payout #%d settlement", payoutID),
ReferenceType: "rider_payout_settlement", ReferenceID: &payoutID, EntryDate: now,
}
if err := tx.Create(&credit).Error; err != nil {
return fmt.Errorf("failed to create bank ledger entry: %w", err)
}
return nil
}

// PostRiderCODDepositLedgerEntry moves the recorded balance of a verified
// COD deposit from Cash (where COD sales are already booked, since the
// rider physically holds the cash) into Bank (SRS 12.9).
//
// Takes tx rather than opening its own transaction against database.DB
// (Bug #14): the caller (VerifyRiderCODDeposit) must be able to run the
// deposit's status update and this ledger posting as one atomic unit, so
// that if the ledger post fails, the status change rolls back too instead
// of leaving the deposit permanently marked "verified" with no matching
// ledger entry - a cash-vs-GL reconciliation gap with no way to detect or
// recover from it after the fact.
func PostRiderCODDepositLedgerEntry(tx *gorm.DB, depositID uint) error {
var existing models.LedgerEntry
if err := tx.Where("reference_type = ? AND reference_id = ?", "rider_cod_deposit", depositID).First(&existing).Error; err == nil {
return nil
}

var deposit models.RiderCODDeposit
if err := tx.First(&deposit, depositID).Error; err != nil {
return fmt.Errorf("rider COD deposit not found: %w", err)
}

var cash, bank models.Account
if err := tx.Where("code = ?", "1001").First(&cash).Error; err != nil {
return fmt.Errorf("chart of accounts missing code 1001 (Cash): %w", err)
}
if err := tx.Where("code = ?", "1002").First(&bank).Error; err != nil {
return fmt.Errorf("chart of accounts missing code 1002 (Bank): %w", err)
}

transactionRef := fmt.Sprintf("CODDEPOSIT-%d", depositID)
now := time.Now()

debit := models.LedgerEntry{
TransactionRef: transactionRef, AccountID: bank.ID, Type: "debit", Amount: deposit.Amount,
Description: fmt.Sprintf("Rider COD deposit #%d", depositID),
ReferenceType: "rider_cod_deposit", ReferenceID: &depositID, EntryDate: now,
}
if err := tx.Create(&debit).Error; err != nil {
return fmt.Errorf("failed to create bank ledger entry: %w", err)
}
credit := models.LedgerEntry{
TransactionRef: transactionRef, AccountID: cash.ID, Type: "credit", Amount: deposit.Amount,
Description: fmt.Sprintf("Rider COD deposit #%d", depositID),
ReferenceType: "rider_cod_deposit", ReferenceID: &depositID, EntryDate: now,
}
if err := tx.Create(&credit).Error; err != nil {
return fmt.Errorf("failed to create cash ledger entry: %w", err)
}
return nil
}
