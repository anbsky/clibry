package client

type ClaimSearchParams struct {
	PageSize      int           `json:"page_size"`
	Page          int           `json:"page"`
	NoTotals      bool          `json:"no_totals"`
	AnyTags       []string      `json:"any_tags"`
	ChannelIds    []interface{} `json:"channel_ids"`
	NotChannelIds []interface{} `json:"not_channel_ids"`
	NotTags       []string      `json:"not_tags"`
	OrderBy       []string      `json:"order_by"`
}