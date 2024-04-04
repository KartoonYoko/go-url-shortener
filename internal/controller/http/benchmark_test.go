package http

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

// BenchmarkHandlerAPIUserURLsGET проверяет получение данных из in-memory хранилища
func BenchmarkHandlerAPIUserURLsGET(b *testing.B) {
	ctx := context.Background()
	controller := createTestMock()
	userID, err := controller.ucAuth.GetNewUserID(ctx)
	require.NoError(b, err)
	_, err = controller.uc.SaveURL(ctx, "https://music.yandex.ru/home", userID)
	require.NoError(b, err)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err = controller.uc.GetUserURLs(ctx, userID)
		require.NoError(b, err)
	}
}
