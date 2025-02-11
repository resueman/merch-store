package operation

import "github.com/resueman/merch-store/internal/repo"

type operationUsecase struct {
	repo repo.Operation
}

func NewOperationUsecase(repo repo.Operation) *operationUsecase {
	return &operationUsecase{repo: repo}
}
