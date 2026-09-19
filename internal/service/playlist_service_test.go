package service_test

import (
	"ZVideo/internal/domain"
	"ZVideo/internal/service"
	"ZVideo/internal/testing/mocks"
	"ZVideo/internal/testing/mother"
	"context"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type PlaylistServiceTestSuite struct {
	suite.Suite
	mockPlaylistRepo *mocks.PlaylistRepository
	mockVideoRepo    *mocks.VideoRepository
	mockChanSvc      *mocks.ChannelService
	service          service.PlaylistService
	mother           mother.PlaylistMother
	vidMother        mother.VideoMother
}

func (s *PlaylistServiceTestSuite) SetupTest() {
	s.mockPlaylistRepo = mocks.NewPlaylistRepository(s.T())
	s.mockVideoRepo = mocks.NewVideoRepository(s.T())
	s.mockChanSvc = mocks.NewChannelService(s.T())
	s.service = service.NewPlaylistService(s.mockPlaylistRepo, s.mockVideoRepo, s.mockChanSvc)
	s.mother = mother.PlaylistMother{}
	s.vidMother = mother.VideoMother{}
}

func (s *PlaylistServiceTestSuite) TestCreate_Positive_StateTransition() {
	ctx := context.Background()
	s.mockChanSvc.On("IsOwner", ctx, 1, 1).Return(true, nil)
	s.mockPlaylistRepo.On("Create", ctx, mock.AnythingOfType("*domain.Playlist")).Return(nil)

	pl, err := s.service.Create(ctx, 1, 1, "Name", "Desc")

	s.NoError(err)
	s.NotNil(pl)
	s.Equal("Name", pl.Name)
}

func (s *PlaylistServiceTestSuite) TestCreate_Negative_EquivalencePartitioning() {
	ctx := context.Background()

	pl, err := s.service.Create(ctx, 1, 1, "", "Desc")

	s.ErrorIs(err, domain.ErrPlaylistNameEmpty)
	s.Nil(pl)
}

func (s *PlaylistServiceTestSuite) TestGetByID_Positive_EquivalencePartitioning() {
	ctx := context.Background()
	expected := s.mother.PlaylistForChannel(1)
	s.mockPlaylistRepo.On("GetByID", ctx, expected.ID).Return(expected, nil)

	pl, err := s.service.GetByID(ctx, expected.ID)

	s.NoError(err)
	s.Equal(expected.ID, pl.ID)
}

func (s *PlaylistServiceTestSuite) TestGetByID_Negative_EquivalencePartitioning() {
	ctx := context.Background()
	s.mockPlaylistRepo.On("GetByID", ctx, 999).Return(nil, nil)

	pl, err := s.service.GetByID(ctx, 999)

	s.ErrorIs(err, domain.ErrPlaylistNotFound)
	s.Nil(pl)
}

func (s *PlaylistServiceTestSuite) TestListByChannel_Positive_BoundaryValueAnalysis() {
	ctx := context.Background()
	expected := []*domain.Playlist{s.mother.PlaylistForChannel(1)}
	s.mockChanSvc.On("Exists", ctx, 1).Return(true, nil)
	s.mockPlaylistRepo.On("ListByChannel", ctx, 1, 10, 0).Return(expected, nil)
	s.mockPlaylistRepo.On("CountByChannel", ctx, 1).Return(int64(len(expected)), nil)

	list, err := s.service.ListByChannel(ctx, 1, 10, 0)

	s.NoError(err)
	s.Len(list.Items, 1)
	s.Equal(int64(1), list.TotalCount)
}

func (s *PlaylistServiceTestSuite) TestListByChannel_Negative_EquivalencePartitioning() {
	ctx := context.Background()
	s.mockChanSvc.On("Exists", ctx, 999).Return(false, nil)

	list, err := s.service.ListByChannel(ctx, 999, 10, 0)

	s.ErrorIs(err, domain.ErrChannelNotFound)
	s.Nil(list)
}

func (s *PlaylistServiceTestSuite) TestUpdate_Positive_StateTransition() {
	ctx := context.Background()
	pl := s.mother.PlaylistForChannel(1)
	newName := "New Name"
	s.mockPlaylistRepo.On("GetByID", ctx, pl.ID).Return(pl, nil)
	s.mockChanSvc.On("IsOwner", ctx, pl.ChannelID, 1).Return(true, nil)
	s.mockPlaylistRepo.On("Update", ctx, pl).Return(nil)

	updated, err := s.service.Update(ctx, pl.ID, 1, &newName, nil)

	s.NoError(err)
	s.Equal(newName, updated.Name)
}

func (s *PlaylistServiceTestSuite) TestUpdate_Negative_Combinatorial() {
	ctx := context.Background()
	pl := s.mother.PlaylistForChannel(1)
	newName := "New Name"
	s.mockPlaylistRepo.On("GetByID", ctx, pl.ID).Return(pl, nil)
	s.mockChanSvc.On("IsOwner", ctx, pl.ChannelID, 999).Return(false, nil)

	updated, err := s.service.Update(ctx, pl.ID, 999, &newName, nil)

	s.ErrorIs(err, domain.ErrForbidden)
	s.Nil(updated)
}

func (s *PlaylistServiceTestSuite) TestDelete_Positive_StateTransition() {
	ctx := context.Background()
	pl := s.mother.PlaylistForChannel(1)
	s.mockPlaylistRepo.On("GetByID", ctx, pl.ID).Return(pl, nil)
	s.mockChanSvc.On("IsOwner", ctx, pl.ChannelID, 1).Return(true, nil)
	s.mockPlaylistRepo.On("Delete", ctx, pl.ID).Return(nil)

	err := s.service.Delete(ctx, pl.ID, 1)

	s.NoError(err)
}

func (s *PlaylistServiceTestSuite) TestDelete_Negative_EquivalencePartitioning() {
	ctx := context.Background()
	s.mockPlaylistRepo.On("GetByID", ctx, 999).Return(nil, nil)

	err := s.service.Delete(ctx, 999, 1)

	s.ErrorIs(err, domain.ErrPlaylistNotFound)
}

func (s *PlaylistServiceTestSuite) TestAddVideo_Positive_StateTransition() {
	ctx := context.Background()
	pl := s.mother.PlaylistForChannel(1)
	vid := s.vidMother.ReadyVideoForChannel(1)
	s.mockPlaylistRepo.On("GetByID", ctx, pl.ID).Return(pl, nil)
	s.mockChanSvc.On("IsOwner", ctx, pl.ChannelID, 1).Return(true, nil)
	s.mockVideoRepo.On("GetByID", ctx, vid.ID).Return(vid, nil)
	s.mockPlaylistRepo.On("AddVideo", ctx, pl.ID, vid.ID).Return(nil)

	err := s.service.AddVideo(ctx, pl.ID, vid.ID, 1)

	s.NoError(err)
}

func (s *PlaylistServiceTestSuite) TestAddVideo_Negative_EquivalencePartitioning() {
	ctx := context.Background()
	pl := s.mother.PlaylistForChannel(1)
	vid := s.vidMother.PendingVideoForChannel(1)
	s.mockPlaylistRepo.On("GetByID", ctx, pl.ID).Return(pl, nil)
	s.mockChanSvc.On("IsOwner", ctx, pl.ChannelID, 1).Return(true, nil)
	s.mockVideoRepo.On("GetByID", ctx, vid.ID).Return(vid, nil)

	err := s.service.AddVideo(ctx, pl.ID, vid.ID, 1)

	s.ErrorIs(err, domain.ErrVideoNotReady)
}

func (s *PlaylistServiceTestSuite) TestRemoveVideo_Positive_StateTransition() {
	ctx := context.Background()
	pl := s.mother.PlaylistForChannel(1)
	s.mockPlaylistRepo.On("GetByID", ctx, pl.ID).Return(pl, nil)
	s.mockChanSvc.On("IsOwner", ctx, pl.ChannelID, 1).Return(true, nil)
	s.mockPlaylistRepo.On("RemoveVideo", ctx, pl.ID, 1).Return(nil)

	err := s.service.RemoveVideo(ctx, pl.ID, 1, 1)

	s.NoError(err)
}

func (s *PlaylistServiceTestSuite) TestRemoveVideo_Negative_EquivalencePartitioning() {
	ctx := context.Background()
	s.mockPlaylistRepo.On("GetByID", ctx, 999).Return(nil, nil)

	err := s.service.RemoveVideo(ctx, 999, 1, 1)

	s.ErrorIs(err, domain.ErrPlaylistNotFound)
}

func (s *PlaylistServiceTestSuite) TestUpdateVideoPosition_Positive_StateTransition() {
	ctx := context.Background()
	pl := s.mother.PlaylistForChannel(1)
	s.mockPlaylistRepo.On("GetByID", ctx, pl.ID).Return(pl, nil)
	s.mockChanSvc.On("IsOwner", ctx, pl.ChannelID, 1).Return(true, nil)
	s.mockPlaylistRepo.On("GetItemsCount", ctx, pl.ID).Return(5, nil)
	s.mockPlaylistRepo.On("UpdateVideoPosition", ctx, pl.ID, 1, 3).Return(nil)

	err := s.service.UpdateVideoPosition(ctx, pl.ID, 1, 1, 3)

	s.NoError(err)
}

func (s *PlaylistServiceTestSuite) TestUpdateVideoPosition_Negative_Combinatorial() {
	ctx := context.Background()
	pl := s.mother.PlaylistForChannel(1)
	s.mockPlaylistRepo.On("GetByID", ctx, pl.ID).Return(pl, nil)
	s.mockChanSvc.On("IsOwner", ctx, pl.ChannelID, 999).Return(false, nil)

	err := s.service.UpdateVideoPosition(ctx, pl.ID, 1, 999, 3)

	s.ErrorIs(err, domain.ErrForbidden)
}

func (s *PlaylistServiceTestSuite) TestGetMyPlaylists_Positive_BoundaryValueAnalysis() {
	ctx := context.Background()
	chMother := mother.ChannelMother{}
	ch := chMother.ChannelForUser(1)
	expected := []*domain.Playlist{s.mother.PlaylistForChannel(ch.ID)}
	s.mockChanSvc.On("GetChannelByUserID", ctx, 1).Return(ch, nil)
	s.mockChanSvc.On("Exists", ctx, ch.ID).Return(true, nil)
	s.mockPlaylistRepo.On("ListByChannel", ctx, ch.ID, 10, 0).Return(expected, nil)
	s.mockPlaylistRepo.On("CountByChannel", ctx, ch.ID).Return(int64(len(expected)), nil)

	list, err := s.service.GetMyPlaylists(ctx, 1, 10, 0)

	s.NoError(err)
	s.Len(list.Items, 1)
	s.Equal(int64(1), list.TotalCount)
}

func (s *PlaylistServiceTestSuite) TestGetMyPlaylists_Negative_EquivalencePartitioning() {
	ctx := context.Background()
	s.mockChanSvc.On("GetChannelByUserID", ctx, 999).Return(nil, domain.ErrChannelNotFound)

	list, err := s.service.GetMyPlaylists(ctx, 999, 10, 0)

	s.Error(err)
	s.Nil(list)
}

func (s *PlaylistServiceTestSuite) TestGetPlaylistItems_Positive_EquivalencePartitioning() {
	ctx := context.Background()
	pl := s.mother.PlaylistForChannel(1)
	expected := []*domain.PlaylistItem{{PlaylistID: pl.ID, VideoID: 1, Number: 1}}
	s.mockPlaylistRepo.On("GetByID", ctx, pl.ID).Return(pl, nil)
	s.mockPlaylistRepo.On("ListItems", ctx, pl.ID, 10, 0).Return(expected, nil)
	s.mockPlaylistRepo.On("GetItemsCount", ctx, pl.ID).Return(len(expected), nil)

	items, err := s.service.GetPlaylistItems(ctx, pl.ID, 10, 0)

	s.NoError(err)
	s.Len(items.Items, 1)
	s.Equal(int64(1), items.TotalCount)
}

func (s *PlaylistServiceTestSuite) TestGetPlaylistItems_Negative_EquivalencePartitioning() {
	ctx := context.Background()
	s.mockPlaylistRepo.On("GetByID", ctx, 999).Return(nil, nil)

	items, err := s.service.GetPlaylistItems(ctx, 999, 10, 0)

	s.ErrorIs(err, domain.ErrPlaylistNotFound)
	s.Nil(items)
}

func TestPlaylistServiceSuite(t *testing.T) {
	suite.Run(t, new(PlaylistServiceTestSuite))
}
