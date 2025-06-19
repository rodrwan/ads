package auction

import (
	"ads-system/internal/database"
	"bytes"
	"context"
	"encoding/json"
	"io"
	"math/big"
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockQuerier es un mock de la interfaz database.Querier
type MockQuerier struct {
	mock.Mock
}

func (m *MockQuerier) GetActiveBidsForPlacement(ctx context.Context, placementID pgtype.UUID) ([]database.GetActiveBidsForPlacementRow, error) {
	args := m.Called(ctx, placementID)
	return args.Get(0).([]database.GetActiveBidsForPlacementRow), args.Error(1)
}

func (m *MockQuerier) InsertAuction(ctx context.Context, arg database.InsertAuctionParams) (database.Auction, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).(database.Auction), args.Error(1)
}

func (m *MockQuerier) InsertAuctionBid(ctx context.Context, arg database.InsertAuctionBidParams) (database.AuctionBid, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).(database.AuctionBid), args.Error(1)
}

// Implementación de otros métodos requeridos por la interfaz
func (m *MockQuerier) AverageCTR(ctx context.Context) (pgtype.Numeric, error) {
	return pgtype.Numeric{}, nil
}

func (m *MockQuerier) CheckCampaignBelongsToAdvertiser(ctx context.Context, arg database.CheckCampaignBelongsToAdvertiserParams) (bool, error) {
	return false, nil
}

func (m *MockQuerier) CountActiveAds(ctx context.Context) (int64, error) {
	return 0, nil
}

func (m *MockQuerier) CountClicks(ctx context.Context) (int64, error) {
	return 0, nil
}

func (m *MockQuerier) CountImpressions(ctx context.Context) (int64, error) {
	return 0, nil
}

func (m *MockQuerier) CreateAd(ctx context.Context, arg database.CreateAdParams) (database.Ad, error) {
	return database.Ad{}, nil
}

func (m *MockQuerier) CreateCampaign(ctx context.Context, arg database.CreateCampaignParams) (database.AdCampaign, error) {
	return database.AdCampaign{}, nil
}

func (m *MockQuerier) CreateClick(ctx context.Context, arg database.CreateClickParams) (database.Click, error) {
	return database.Click{}, nil
}

func (m *MockQuerier) CreateImpression(ctx context.Context, arg database.CreateImpressionParams) (database.Impression, error) {
	return database.Impression{}, nil
}

func (m *MockQuerier) CreateSession(ctx context.Context, arg database.CreateSessionParams) (database.Session, error) {
	return database.Session{}, nil
}

func (m *MockQuerier) DeleteSession(ctx context.Context, token pgtype.UUID) error {
	return nil
}

func (m *MockQuerier) GetAdPerformanceAnalysis(ctx context.Context, advertiserID pgtype.UUID) ([]database.GetAdPerformanceAnalysisRow, error) {
	return nil, nil
}

func (m *MockQuerier) GetAds(ctx context.Context) ([]database.Ad, error) {
	return nil, nil
}

func (m *MockQuerier) GetAdsByCampaign(ctx context.Context, campaignID pgtype.UUID) ([]database.GetAdsByCampaignRow, error) {
	return nil, nil
}

func (m *MockQuerier) GetAdvertiserByEmail(ctx context.Context, email string) (database.Advertiser, error) {
	return database.Advertiser{}, nil
}

func (m *MockQuerier) GetBudgetAlerts(ctx context.Context, advertiserID pgtype.UUID) ([]database.GetBudgetAlertsRow, error) {
	return nil, nil
}

func (m *MockQuerier) GetCampaignByID(ctx context.Context, id pgtype.UUID) (database.AdCampaign, error) {
	return database.AdCampaign{}, nil
}

func (m *MockQuerier) GetCampaignMetrics(ctx context.Context, id pgtype.UUID) ([]database.GetCampaignMetricsRow, error) {
	return nil, nil
}

func (m *MockQuerier) GetCampaignPerformanceAnalysis(ctx context.Context, advertiserID pgtype.UUID) ([]database.GetCampaignPerformanceAnalysisRow, error) {
	return nil, nil
}

func (m *MockQuerier) GetCampaignSpend(ctx context.Context, id pgtype.UUID) (interface{}, error) {
	return nil, nil
}

func (m *MockQuerier) GetCampaignsByAdvertiser(ctx context.Context, advertiserID pgtype.UUID) ([]database.GetCampaignsByAdvertiserRow, error) {
	return nil, nil
}

func (m *MockQuerier) GetLastImpressionByAdID(ctx context.Context, adID pgtype.UUID) (pgtype.UUID, error) {
	return pgtype.UUID{}, nil
}

func (m *MockQuerier) GetOptimizationRecommendations(ctx context.Context, advertiserID pgtype.UUID) ([]database.GetOptimizationRecommendationsRow, error) {
	return nil, nil
}

func (m *MockQuerier) GetSessionByToken(ctx context.Context, token pgtype.UUID) (pgtype.UUID, error) {
	return pgtype.UUID{}, nil
}

func (m *MockQuerier) UpdateCampaign(ctx context.Context, arg database.UpdateCampaignParams) (database.AdCampaign, error) {
	return database.AdCampaign{}, nil
}

func (m *MockQuerier) UpdateCampaignStatus(ctx context.Context, arg database.UpdateCampaignStatusParams) (database.AdCampaign, error) {
	return database.AdCampaign{}, nil
}

func TestCreateAuction(t *testing.T) {
	// Configurar el modo de prueba de Gin
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		request        CreateAuctionRequest
		setupMock      func(*MockQuerier)
		expectedStatus int
		expectedError  bool
	}{
		{
			name: "successful auction",
			request: CreateAuctionRequest{
				PlacementID: uuid.New().String(),
				RequestContext: struct {
					Country   string   `json:"country"`
					Device    string   `json:"device"`
					Interests []string `json:"interests"`
					OS        string   `json:"os"`
					Browser   string   `json:"browser"`
					IP        string   `json:"ip"`
					UserAgent string   `json:"user_agent"`
					Referer   string   `json:"referer"`
				}{
					Country: "US",
					Device:  "mobile",
				},
			},
			setupMock: func(m *MockQuerier) {
				// Mock GetActiveBidsForPlacement
				m.On("GetActiveBidsForPlacement", mock.Anything, mock.Anything).Return(
					[]database.GetActiveBidsForPlacementRow{
						{
							ID: pgtype.UUID{},
							BidPrice: pgtype.Numeric{
								Int:   big.NewInt(200),
								Valid: true,
							},
							Title: pgtype.Text{
								String: "Test Ad",
								Valid:  true,
							},
							TargetingJson: []byte(`{"country": "US"}`),
						},
					},
					nil,
				)

				// Mock InsertAuction
				m.On("InsertAuction", mock.Anything, mock.Anything).Return(
					database.Auction{
						ID: pgtype.UUID{},
					},
					nil,
				)

				// Mock InsertAuctionBid
				m.On("InsertAuctionBid", mock.Anything, mock.Anything).Return(
					database.AuctionBid{
						ID: pgtype.UUID{},
					},
					nil,
				).Maybe()
			},
			expectedStatus: 201,
			expectedError:  false,
		},
		{
			name: "invalid placement ID",
			request: CreateAuctionRequest{
				PlacementID: "invalid-uuid",
			},
			setupMock:      func(m *MockQuerier) {},
			expectedStatus: 400,
			expectedError:  true,
		},
		{
			name: "no bids available",
			request: CreateAuctionRequest{
				PlacementID: uuid.New().String(),
			},
			setupMock: func(m *MockQuerier) {
				m.On("GetActiveBidsForPlacement", mock.Anything, mock.Anything).Return(
					[]database.GetActiveBidsForPlacementRow{},
					nil,
				)
			},
			expectedStatus: 500,
			expectedError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Crear mock y handler
			mockQuerier := new(MockQuerier)
			tt.setupMock(mockQuerier)
			handler := NewAuctionHandler(mockQuerier)

			// Crear contexto de Gin
			c, _ := gin.CreateTestContext(nil)
			c.Request = &http.Request{
				Header: make(http.Header),
			}

			// Configurar el cuerpo de la solicitud
			jsonBytes, _ := json.Marshal(tt.request)
			c.Request.Body = io.NopCloser(bytes.NewBuffer(jsonBytes))

			// Ejecutar el handler
			response, status, err := handler.CreateAuction(c)

			// Verificar resultados
			if tt.expectedError {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedStatus, status)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedStatus, status)
				assert.NotNil(t, response)
			}

			// Verificar que se llamaron todos los métodos mock esperados
			mockQuerier.AssertExpectations(t)
		})
	}
}

func TestRunAuction(t *testing.T) {
	tests := []struct {
		name       string
		bids       []database.GetActiveBidsForPlacementRow
		wantWinner *database.GetActiveBidsForPlacementRow
		wantOthers []database.GetActiveBidsForPlacementRow
		wantPanic  bool
	}{
		{
			name: "successful auction with multiple bids",
			bids: []database.GetActiveBidsForPlacementRow{
				{
					ID: pgtype.UUID{},
					BidPrice: pgtype.Numeric{
						Int:   big.NewInt(100),
						Valid: true,
					},
				},
				{
					ID: pgtype.UUID{},
					BidPrice: pgtype.Numeric{
						Int:   big.NewInt(200),
						Valid: true,
					},
				},
				{
					ID: pgtype.UUID{},
					BidPrice: pgtype.Numeric{
						Int:   big.NewInt(150),
						Valid: true,
					},
				},
			},
			wantWinner: &database.GetActiveBidsForPlacementRow{
				ID: pgtype.UUID{},
				BidPrice: pgtype.Numeric{
					Int:   big.NewInt(200),
					Valid: true,
				},
			},
		},
		{
			name: "auction with single bid",
			bids: []database.GetActiveBidsForPlacementRow{
				{
					ID: pgtype.UUID{},
					BidPrice: pgtype.Numeric{
						Int:   big.NewInt(100),
						Valid: true,
					},
				},
			},
			wantWinner: &database.GetActiveBidsForPlacementRow{
				ID: pgtype.UUID{},
				BidPrice: pgtype.Numeric{
					Int:   big.NewInt(100),
					Valid: true,
				},
			},
		},
		{
			name:      "auction with no bids",
			bids:      []database.GetActiveBidsForPlacementRow{},
			wantPanic: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.wantPanic {
				assert.Panics(t, func() {
					RunAuction(tt.bids)
				})
				return
			}

			winner, others := RunAuction(tt.bids)

			// Verificar el precio del ganador
			winnerPrice, _ := winner.BidPrice.Int64Value()
			expectedWinnerPrice, _ := tt.wantWinner.BidPrice.Int64Value()
			assert.Equal(t, expectedWinnerPrice.Int64, winnerPrice.Int64)

			// Verificar que el ganador tiene el precio más alto
			for _, other := range others {
				otherPrice, _ := other.BidPrice.Int64Value()
				assert.Greater(t, winnerPrice.Int64, otherPrice.Int64)
			}

			// Verificar que el número de otros es correcto
			assert.Equal(t, len(tt.bids)-1, len(others))
		})
	}
}
