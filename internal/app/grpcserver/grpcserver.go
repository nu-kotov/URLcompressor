package grpcserver

import (
	"context"

	"github.com/nu-kotov/URLcompressor/internal/app/api/service"
	"github.com/nu-kotov/URLcompressor/internal/app/models"
	"github.com/nu-kotov/URLcompressor/internal/app/proto"
)

// GRPCServer структура сервера gRPC-сервиса.
type GRPCServer struct {
	proto.UnimplementedURLcompressorServer
	service service.Service
}

// NewgRPCServer создаёт новый экземпляр gRPC-сервера с переданным сервисом.
func NewgRPCServer(service service.Service) *GRPCServer {
	return &GRPCServer{
		service: service,
	}
}

// PingDB - проверяет доступность базы данных.
func (s *GRPCServer) PingDB(ctx context.Context, _ *proto.PingDBRequest) (*proto.PingDBResponse, error) {
	err := s.service.PingDB()
	if err != nil {
		return nil, err
	}
	return &proto.PingDBResponse{}, nil
}

// GetShortURL - возвращает сокращенный урл пользователя по полному урлу.
func (s *GRPCServer) GetShortURL(ctx context.Context, req *proto.GetShortURLRequest) (*proto.GetShortURLResponse, error) {
	shortURL, err := s.service.GetShortURLSrv(ctx, []byte(req.OriginalUrl), req.UserId)
	if err != nil {
		return nil, err
	}
	return &proto.GetShortURLResponse{ShortUrl: shortURL.Result}, nil
}

// GetOriginalURL - возвращает оригинальный урл пользователя по сокращенному урлу.
func (s *GRPCServer) GetOriginalURL(ctx context.Context, req *proto.GetOriginalURLRequest) (*proto.GetOriginalURLResponse, error) {
	originalURL, err := s.service.SelectOriginalURLByShortURL(ctx, req.ShortUrlId)
	if err != nil {
		return nil, err
	}
	return &proto.GetOriginalURLResponse{OriginalUrl: originalURL}, nil
}

// GetShortURLsBatch - возвращает батч сокращенных урлов.
func (s *GRPCServer) GetShortURLsBatch(ctx context.Context, req *proto.GetShortURLsBatchRequest) (*proto.GetShortURLsBatchResponse, error) {
	var batch []models.GetShortURLsBatchRequest
	for _, item := range req.Items {
		batch = append(batch, models.GetShortURLsBatchRequest{CorrelationID: item.CorrelationId, OriginalURL: item.OriginalUrl})
	}
	shortURLsBatch, err := s.service.GetShortURLsBatch(ctx, batch, req.UserId)
	if err != nil {
		return nil, err
	}

	respItems := make([]*proto.GetShortURLsBatchResponseItem, len(shortURLsBatch))
	for i, item := range shortURLsBatch {
		respItems[i] = &proto.GetShortURLsBatchResponseItem{
			CorrelationId: item.CorrelationID,
			ShortUrl:      item.ShortURL,
		}
	}
	return &proto.GetShortURLsBatchResponse{Items: respItems}, nil
}

// GetUserURLs - возвращает урлы по id пользователя.
func (s *GRPCServer) GetUserURLs(ctx context.Context, req *proto.GetUserURLsRequest) (*proto.GetUserURLsResponse, error) {
	userURLs, err := s.service.GetUserURLs(ctx, req.UserId)
	if err != nil {
		return nil, err
	}

	respURLs := make([]*proto.GetUserURLItem, len(userURLs))
	for i, item := range userURLs {
		respURLs[i] = &proto.GetUserURLItem{
			ShortUrl:    item.ShortURL,
			OriginalUrl: item.OriginalURL,
		}
	}
	return &proto.GetUserURLsResponse{Urls: respURLs}, nil
}

// DeleteUserURLs помещает список ID урлов ["IhqFu4fdBD9w", "50ZT5FOYE6y"] в канал для удаления.
func (s *GRPCServer) DeleteUserURLs(ctx context.Context, req *proto.DeleteURLsRequest) (*proto.DeleteURLsResponse, error) {

	go s.service.SendURLsToDeletion(req.ShortUrls, req.UserId)

	return &proto.DeleteURLsResponse{}, nil
}

// GetStats возвращает количество пользователей и урлов в сервисе.
func (s *GRPCServer) GetStats(ctx context.Context, req *proto.StatsRequest) (*proto.StatsResponse, error) {
	stats, err := s.service.GetStats(ctx)
	if err != nil {
		return nil, err
	}

	return &proto.StatsResponse{Urls: int32(stats.URLs), Users: int32(stats.Users)}, nil
}
