package templates

type DashboardMetrics struct {
	TotalImpressions int
	TotalClicks      int
	CTR              float64
	ActiveAds        int
}

type Ad struct {
	ID             string
	Title          string
	Description    string
	ImageURL       string
	DestinationURL string
	CtaLabel       string
	CampaignID     string
}
