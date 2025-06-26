package storage

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/nu-kotov/URLcompressor/internal/app/models"
	"github.com/nu-kotov/URLcompressor/mocks"
)

func BenchmarkCompressURLWithPGMock(b *testing.B) {
	ctrl := gomock.NewController(b)
	storage := mocks.NewMockStorage(ctrl)

	storage.EXPECT().InsertURLsData(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = storage.InsertURLsData(context.Background(), &models.URLsData{ShortURL: "https://test.com"})
	}
}
