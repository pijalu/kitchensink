package quietlog

import (
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/pijalu/kitchensink/mocks"
)

func TestPrintf(t *testing.T) {
	mockController := gomock.NewController(t)
	defer mockController.Finish()

	logger := mocks.NewMockLogger(mockController)
	quieter := mocks.NewMockQuieter(mockController)

	quieter.EXPECT().Quiet().Return(false)
	logger.EXPECT().Printf("hello %s", "world")

	New(logger, quieter).Printf("hello %s", "world")
}

func TestPrintfWhenQuiet(t *testing.T) {
	mockController := gomock.NewController(t)
	defer mockController.Finish()

	logger := mocks.NewMockLogger(mockController)
	quieter := mocks.NewMockQuieter(mockController)

	quieter.EXPECT().Quiet().Return(true)

	New(logger, quieter).Printf("hello %s", "world")
}
