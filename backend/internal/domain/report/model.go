package report

type PageResult struct {
	Rows     [][]string `json:"rows"`
	Headers  []string   `json:"headers"`
	Total    int        `json:"total"`
	Page     int        `json:"page"`
	PageSize int        `json:"page_size"`
}

type StockOnHandFilters struct {
	ItemType   string
	LotStatus  string
	NearExpiry bool
	ItemCode   string
	Page       int
	PageSize   int
}

type MovementLedgerFilters struct {
	FromDate        string
	ToDate          string
	TransactionType string
	ItemCode        string
	LotInternalID   string
	Page            int
	PageSize        int
}

type WOSummaryFilters struct {
	Status       string
	FromDate     string
	ToDate       string
	TargetFGCode string
	Page         int
	PageSize     int
}

type ShortageDamageFilters struct {
	FromDate   string
	ToDate     string
	ReasonCode string
	Page       int
	PageSize   int
}

type AuditTrailFilters struct {
	WONumber  string
	Action    string
	UserID    string
	FromDate  string
	ToDate    string
	Page      int
	PageSize  int
}

type PartialCompletionFilters struct {
	FromDate       string
	ToDate         string
	ApprovalStatus string
	Page           int
	PageSize       int
}
