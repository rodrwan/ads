package campaigns

import (
	"context"
	"math/big"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"ads-system/internal/database"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// --- Mock de database.Querier ---
type MockDB struct {
	mock.Mock
}

// Métodos que se usan en los tests - con implementaciones de mock
func (m *MockDB) CheckCampaignBelongsToAdvertiser(ctx context.Context, arg database.CheckCampaignBelongsToAdvertiserParams) (bool, error) {
	args := m.Called(ctx, arg)
	return args.Bool(0), args.Error(1)
}

func (m *MockDB) GetCampaignByID(ctx context.Context, id pgtype.UUID) (database.AdCampaign, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(database.AdCampaign), args.Error(1)
}

func (m *MockDB) GetCampaignMetrics(ctx context.Context, id pgtype.UUID) ([]database.GetCampaignMetricsRow, error) {
	args := m.Called(ctx, id)
	return args.Get(0).([]database.GetCampaignMetricsRow), args.Error(1)
}

func (m *MockDB) GetCampaignSpend(ctx context.Context, id pgtype.UUID) (interface{}, error) {
	args := m.Called(ctx, id)
	return args.Get(0), args.Error(1)
}

func (m *MockDB) GetAdsByCampaign(ctx context.Context, campaignID pgtype.UUID) ([]database.GetAdsByCampaignRow, error) {
	args := m.Called(ctx, campaignID)
	return args.Get(0).([]database.GetAdsByCampaignRow), args.Error(1)
}

func (m *MockDB) UpdateCampaign(ctx context.Context, arg database.UpdateCampaignParams) (database.AdCampaign, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).(database.AdCampaign), args.Error(1)
}

func (m *MockDB) UpdateCampaignStatus(ctx context.Context, arg database.UpdateCampaignStatusParams) (database.AdCampaign, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).(database.AdCampaign), args.Error(1)
}

func (m *MockDB) GetCampaignPerformanceAnalysis(ctx context.Context, advertiserID pgtype.UUID) ([]database.GetCampaignPerformanceAnalysisRow, error) {
	args := m.Called(ctx, advertiserID)
	return args.Get(0).([]database.GetCampaignPerformanceAnalysisRow), args.Error(1)
}

func (m *MockDB) GetAdPerformanceAnalysis(ctx context.Context, advertiserID pgtype.UUID) ([]database.GetAdPerformanceAnalysisRow, error) {
	args := m.Called(ctx, advertiserID)
	return args.Get(0).([]database.GetAdPerformanceAnalysisRow), args.Error(1)
}

func (m *MockDB) GetBudgetAlerts(ctx context.Context, advertiserID pgtype.UUID) ([]database.GetBudgetAlertsRow, error) {
	args := m.Called(ctx, advertiserID)
	return args.Get(0).([]database.GetBudgetAlertsRow), args.Error(1)
}

func (m *MockDB) GetOptimizationRecommendations(ctx context.Context, advertiserID pgtype.UUID) ([]database.GetOptimizationRecommendationsRow, error) {
	args := m.Called(ctx, advertiserID)
	return args.Get(0).([]database.GetOptimizationRecommendationsRow), args.Error(1)
}

// Stubs para métodos no usados en estos tests
func (m *MockDB) AverageCTR(ctx context.Context) (pgtype.Numeric, error) {
	return pgtype.Numeric{}, nil
}

func (m *MockDB) CountActiveAds(ctx context.Context) (int64, error) {
	return 0, nil
}

func (m *MockDB) CountClicks(ctx context.Context) (int64, error) {
	return 0, nil
}

func (m *MockDB) CountImpressions(ctx context.Context) (int64, error) {
	return 0, nil
}

func (m *MockDB) CreateAd(ctx context.Context, arg database.CreateAdParams) (database.Ad, error) {
	return database.Ad{}, nil
}

func (m *MockDB) CreateCampaign(ctx context.Context, arg database.CreateCampaignParams) (database.AdCampaign, error) {
	return database.AdCampaign{}, nil
}

func (m *MockDB) CreateClick(ctx context.Context, arg database.CreateClickParams) (database.Click, error) {
	return database.Click{}, nil
}

func (m *MockDB) CreateImpression(ctx context.Context, arg database.CreateImpressionParams) (database.Impression, error) {
	return database.Impression{}, nil
}

func (m *MockDB) CreateSession(ctx context.Context, arg database.CreateSessionParams) (database.Session, error) {
	return database.Session{}, nil
}

func (m *MockDB) DeleteSession(ctx context.Context, token pgtype.UUID) error {
	return nil
}

func (m *MockDB) GetActiveBidsForPlacement(ctx context.Context, placementID pgtype.UUID) ([]database.GetActiveBidsForPlacementRow, error) {
	return []database.GetActiveBidsForPlacementRow{}, nil
}

func (m *MockDB) GetAds(ctx context.Context) ([]database.Ad, error) {
	return []database.Ad{}, nil
}

func (m *MockDB) GetAdvertiserByEmail(ctx context.Context, email string) (database.Advertiser, error) {
	return database.Advertiser{}, nil
}

func (m *MockDB) GetCampaignsByAdvertiser(ctx context.Context, advertiserID pgtype.UUID) ([]database.GetCampaignsByAdvertiserRow, error) {
	return []database.GetCampaignsByAdvertiserRow{}, nil
}

func (m *MockDB) GetLastImpressionByAdID(ctx context.Context, adID pgtype.UUID) (pgtype.UUID, error) {
	return pgtype.UUID{}, nil
}

func (m *MockDB) GetSessionByToken(ctx context.Context, token pgtype.UUID) (pgtype.UUID, error) {
	return pgtype.UUID{}, nil
}

func (m *MockDB) InsertAuction(ctx context.Context, arg database.InsertAuctionParams) (database.Auction, error) {
	return database.Auction{}, nil
}

func (m *MockDB) InsertAuctionBid(ctx context.Context, arg database.InsertAuctionBidParams) (database.AuctionBid, error) {
	return database.AuctionBid{}, nil
}

// --- Helpers para tests ---
func testUUID() uuid.UUID {
	return uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
}

func testPgUUID() pgtype.UUID {
	return pgtype.UUID{Bytes: testUUID(), Valid: true}
}

func testCampaign() database.AdCampaign {
	return database.AdCampaign{
		ID:           testPgUUID(),
		AdvertiserID: testPgUUID(),
		Name:         "Test Campaign",
		Status:       pgtype.Text{String: "active", Valid: true},
		Budget:       pgtype.Numeric{Int: big.NewInt(10000), Valid: true},
		DailyBudget:  pgtype.Numeric{Int: big.NewInt(1000), Valid: true},
		StartDate:    pgtype.Date{Time: time.Now(), Valid: true},
		EndDate:      pgtype.Date{Time: time.Now().AddDate(0, 1, 0), Valid: true},
		CreatedAt:    pgtype.Timestamp{Time: time.Now(), Valid: true},
	}
}

func testMetrics() []database.GetCampaignMetricsRow {
	return []database.GetCampaignMetricsRow{{Impressions: 100, Clicks: 10, Ctr: 10}}
}

func testAds() []database.GetAdsByCampaignRow {
	return []database.GetAdsByCampaignRow{{ID: testPgUUID(), Title: pgtype.Text{String: "Ad", Valid: true}}}
}

func addSessionCookie(req *http.Request) {
	cookie := &http.Cookie{Name: "session_token", Value: testUUID().String()}
	req.AddCookie(cookie)
}

// --- TESTS ShowCampaign ---
func TestShowCampaign_Unauthorized(t *testing.T) {
	h := NewHandler(new(MockDB))
	req := httptest.NewRequest("GET", "/campaigns/123", nil)
	w := httptest.NewRecorder()
	h.ShowCampaign(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestShowCampaign_InvalidID(t *testing.T) {
	h := NewHandler(new(MockDB))
	req := httptest.NewRequest("GET", "/campaigns/invalid-uuid", nil)
	addSessionCookie(req)
	w := httptest.NewRecorder()
	h.ShowCampaign(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestShowCampaign_Success(t *testing.T) {
	mockDB := new(MockDB)
	h := NewHandler(mockDB)
	cid := testUUID()
	addPath := "/campaigns/" + cid.String()
	req := httptest.NewRequest("GET", addPath, nil)
	addSessionCookie(req)
	w := httptest.NewRecorder()
	mockDB.On("CheckCampaignBelongsToAdvertiser", mock.Anything, mock.Anything).Return(true, nil)
	mockDB.On("GetCampaignByID", mock.Anything, mock.Anything).Return(testCampaign(), nil)
	mockDB.On("GetCampaignMetrics", mock.Anything, mock.Anything).Return(testMetrics(), nil)
	mockDB.On("GetCampaignSpend", mock.Anything, mock.Anything).Return(float64(100), nil)
	mockDB.On("GetAdsByCampaign", mock.Anything, mock.Anything).Return(testAds(), nil)
	h.ShowCampaign(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

// --- TESTS UpdateCampaign ---
func TestUpdateCampaign_Unauthorized(t *testing.T) {
	h := NewHandler(new(MockDB))
	req := httptest.NewRequest("POST", "/campaigns/123", nil)
	w := httptest.NewRecorder()
	h.UpdateCampaign(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestUpdateCampaign_InvalidID(t *testing.T) {
	h := NewHandler(new(MockDB))
	req := httptest.NewRequest("POST", "/campaigns/invalid-uuid", nil)
	addSessionCookie(req)
	w := httptest.NewRecorder()
	h.UpdateCampaign(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUpdateCampaign_Success(t *testing.T) {
	mockDB := new(MockDB)
	h := NewHandler(mockDB)
	cid := testUUID()
	form := make(url.Values)
	form.Set("name", "Test")
	form.Set("budget", "100.0")
	form.Set("daily_budget", "10.0")
	form.Set("start_date", "2024-01-01")
	form.Set("end_date", "2024-12-31")
	addPath := "/campaigns/" + cid.String()
	req := httptest.NewRequest("POST", addPath, strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	addSessionCookie(req)
	w := httptest.NewRecorder()
	mockDB.On("UpdateCampaign", mock.Anything, mock.Anything).Return(testCampaign(), nil)
	h.UpdateCampaign(w, req)
	assert.Equal(t, http.StatusSeeOther, w.Code)
}

// --- TESTS PauseCampaign ---
func TestPauseCampaign_Unauthorized(t *testing.T) {
	h := NewHandler(new(MockDB))
	req := httptest.NewRequest("POST", "/campaigns/123/pause", nil)
	w := httptest.NewRecorder()
	h.PauseCampaign(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestPauseCampaign_InvalidID(t *testing.T) {
	h := NewHandler(new(MockDB))
	req := httptest.NewRequest("POST", "/campaigns/invalid-uuid/pause", nil)
	addSessionCookie(req)
	w := httptest.NewRecorder()
	h.PauseCampaign(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestPauseCampaign_Success(t *testing.T) {
	mockDB := new(MockDB)
	h := NewHandler(mockDB)
	cid := testUUID()
	addPath := "/campaigns/" + cid.String() + "/pause"
	req := httptest.NewRequest("POST", addPath, nil)
	addSessionCookie(req)
	w := httptest.NewRecorder()
	mockDB.On("UpdateCampaignStatus", mock.Anything, mock.Anything).Return(testCampaign(), nil)
	h.PauseCampaign(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "Pausada")
}

// --- TESTS ActivateCampaign ---
func TestActivateCampaign_Unauthorized(t *testing.T) {
	h := NewHandler(new(MockDB))
	req := httptest.NewRequest("POST", "/campaigns/123/activate", nil)
	w := httptest.NewRecorder()
	h.ActivateCampaign(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestActivateCampaign_InvalidID(t *testing.T) {
	h := NewHandler(new(MockDB))
	req := httptest.NewRequest("POST", "/campaigns/invalid-uuid/activate", nil)
	addSessionCookie(req)
	w := httptest.NewRecorder()
	h.ActivateCampaign(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestActivateCampaign_Success(t *testing.T) {
	mockDB := new(MockDB)
	h := NewHandler(mockDB)
	cid := testUUID()
	addPath := "/campaigns/" + cid.String() + "/activate"
	req := httptest.NewRequest("POST", addPath, nil)
	addSessionCookie(req)
	w := httptest.NewRecorder()
	mockDB.On("UpdateCampaignStatus", mock.Anything, mock.Anything).Return(testCampaign(), nil)
	h.ActivateCampaign(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "Activa")
}

// --- TESTS Optimization ---
func TestOptimization_Unauthorized(t *testing.T) {
	h := NewHandler(new(MockDB))
	req := httptest.NewRequest("GET", "/optimization", nil)
	w := httptest.NewRecorder()
	h.Optimization(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestOptimization_Success(t *testing.T) {
	mockDB := new(MockDB)
	h := NewHandler(mockDB)
	addPath := "/optimization"
	req := httptest.NewRequest("GET", addPath, nil)
	addSessionCookie(req)
	w := httptest.NewRecorder()
	mockDB.On("GetCampaignPerformanceAnalysis", mock.Anything, mock.Anything).Return([]database.GetCampaignPerformanceAnalysisRow{}, nil)
	mockDB.On("GetAdPerformanceAnalysis", mock.Anything, mock.Anything).Return([]database.GetAdPerformanceAnalysisRow{}, nil)
	mockDB.On("GetBudgetAlerts", mock.Anything, mock.Anything).Return([]database.GetBudgetAlertsRow{}, nil)
	mockDB.On("GetOptimizationRecommendations", mock.Anything, mock.Anything).Return([]database.GetOptimizationRecommendationsRow{}, nil)
	h.Optimization(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}
