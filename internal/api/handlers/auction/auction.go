package auction

import (
	"ads-system/internal/database"
	"encoding/json"
	"net/http"
	"sort"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type AuctionHandler struct {
	Querier database.Querier
}

func NewAuctionHandler(querier database.Querier) *AuctionHandler {
	return &AuctionHandler{
		Querier: querier,
	}
}

type CreateAuctionRequest struct {
	PlacementID    string `json:"placement_id"`
	RequestContext struct {
		Country   string   `json:"country"`
		Device    string   `json:"device"`
		Interests []string `json:"interests"`
		OS        string   `json:"os"`
		Browser   string   `json:"browser"`
		IP        string   `json:"ip"`
		UserAgent string   `json:"user_agent"`
		Referer   string   `json:"referer"`
	} `json:"request_context"`
}

type CreateAuctionResponse struct {
	AdID           string `json:"ad_id"`
	Title          string `json:"title"`
	Description    string `json:"description"`
	ImageURL       string `json:"image_url"`
	DestinationURL string `json:"destination_url"`
	CTALabel       string `json:"cta_label"`
}

func (uh *AuctionHandler) CreateAuction(c *gin.Context) (interface{}, int, error) {
	var request CreateAuctionRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		return nil, http.StatusBadRequest, err
	}

	var placementUUID pgtype.UUID
	err := placementUUID.Scan(request.PlacementID)
	if err != nil {
		return nil, http.StatusBadRequest, err
	}

	// 1. Obtener bids válidas para ese placement
	ads, err := uh.Querier.GetActiveBidsForPlacement(c, placementUUID)
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}

	// 2. Filtrar por targeting (simplificado)
	filteredAds := []database.GetActiveBidsForPlacementRow{}
	for _, ad := range ads {
		if ad.TargetingJson != nil {

			filteredAds = append(filteredAds, ad)
		}
	}
	// 3. Calcular score y elegir ganador (subasta)
	winner, others := RunAuction(filteredAds)

	// 4. Registrar auction y pujas
	requestContextBytes, err := json.Marshal(request.RequestContext)
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}

	var auctionID pgtype.UUID
	_ = auctionID.Scan(uuid.New().String())

	var zeroNumeric pgtype.Numeric
	_ = zeroNumeric.Scan(int64(0))

	var falseBool pgtype.Bool
	_ = falseBool.Scan(false)

	auction, err := uh.Querier.InsertAuction(c, database.InsertAuctionParams{
		ID:             auctionID,
		PlacementID:    placementUUID,
		RequestContext: requestContextBytes,
	})
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}

	for _, bid := range others {
		var bidUUID pgtype.UUID
		_ = bidUUID.Scan(uuid.New().String())

		_, err := uh.Querier.InsertAuctionBid(c, database.InsertAuctionBidParams{
			ID:        bidUUID,
			AuctionID: auction.ID,
			BidID:     bid.ID,
			Price:     bid.BidPrice,
			IsWinner:  falseBool,
		})
		if err != nil {
			return nil, http.StatusInternalServerError, err
		}
	}

	// 5. Retornar anuncio ganador
	response := CreateAuctionResponse{
		AdID:           winner.ID.String(),
		Title:          winner.Title.String,
		Description:    winner.Description.String,
		ImageURL:       winner.ImageUrl.String,
		DestinationURL: winner.DestinationUrl.String,
		CTALabel:       winner.CtaLabel.String,
	}
	return response, http.StatusCreated, nil
}

func RunAuction(bids []database.GetActiveBidsForPlacementRow) (database.GetActiveBidsForPlacementRow, []database.GetActiveBidsForPlacementRow) {
	sort.Slice(bids, func(i, j int) bool {
		iVal, _ := bids[i].BidPrice.Int64Value()
		jVal, _ := bids[j].BidPrice.Int64Value()
		return iVal.Int64 > jVal.Int64
	})
	winner := bids[0]
	others := bids[1:]
	return winner, others
}
