package products

import (
	"context"
	"fmt"

	repo "github/M-b-a-s/e-comm/internal/adapters/postgresql/sqlc"

	"github.com/jackc/pgx/v5/pgtype"
)

type repository interface {
	CreateProduct(context.Context, repo.CreateProductParams) (repo.Product, error)
	UpdateProduct(context.Context, repo.UpdateProductParams) (repo.Product, error)
	DeleteProduct(context.Context, int64) (repo.Product, error)
	GetProductByID(context.Context, int64) (repo.Product, error)
	ListProducts(context.Context) ([]repo.Product, error)
}

type Service struct {
	repository repository
}

func NewService(repository repository) *Service {
	return &Service{repository: repository}
}

type Input struct {
	Name                  string
	PriceInCents          int32
	Slug                  string
	ShortName             string
	CategoryID            int32
	IsNew                 bool
	Description           string
	Features              string
	BoxIncludes           []byte
	Gallery               []byte
	CategoryImage         string
	RecommendedProductIDs []int64
}

func (s *Service) Create(ctx context.Context, input Input) (repo.Product, error) {
	product, err := s.repository.CreateProduct(ctx, productParams(input))
	if err != nil {
		return repo.Product{}, fmt.Errorf("create product: %w", err)
	}
	return product, nil
}

func (s *Service) Update(ctx context.Context, id int64, input Input) (repo.Product, error) {
	params := updateProductParams(input)
	params.ID = id

	product, err := s.repository.UpdateProduct(ctx, params)
	if err != nil {
		return repo.Product{}, fmt.Errorf("update product %d: %w", id, err)
	}
	return product, nil
}

func (s *Service) Delete(ctx context.Context, id int64) (repo.Product, error) {
	product, err := s.repository.DeleteProduct(ctx, id)
	if err != nil {
		return repo.Product{}, fmt.Errorf("delete product %d: %w", id, err)
	}
	return product, nil
}

func (s *Service) GetByID(ctx context.Context, id int64) (repo.Product, error) {
	product, err := s.repository.GetProductByID(ctx, id)
	if err != nil {
		return repo.Product{}, fmt.Errorf("get product %d: %w", id, err)
	}
	return product, nil
}

func (s *Service) List(ctx context.Context) ([]repo.Product, error) {
	products, err := s.repository.ListProducts(ctx)
	if err != nil {
		return nil, fmt.Errorf("list products: %w", err)
	}
	return products, nil
}

func productParams(input Input) repo.CreateProductParams {
	return repo.CreateProductParams{
		Name:                  input.Name,
		PriceInCents:          input.PriceInCents,
		Slug:                  pgtype.Text{String: input.Slug, Valid: true},
		ShortName:             pgtype.Text{String: input.ShortName, Valid: true},
		CategoryID:            pgtype.Int8{Int64: int64(input.CategoryID), Valid: true},
		IsNew:                 input.IsNew,
		Description:           pgtype.Text{String: input.Description, Valid: true},
		Features:              pgtype.Text{String: input.Features, Valid: true},
		BoxIncludes:           input.BoxIncludes,
		Gallery:               input.Gallery,
		CategoryImage:         pgtype.Text{String: input.CategoryImage, Valid: true},
		RecommendedProductIds: input.RecommendedProductIDs,
	}
}

func updateProductParams(input Input) repo.UpdateProductParams {
	return repo.UpdateProductParams{
		Name:                  input.Name,
		PriceInCents:          input.PriceInCents,
		Slug:                  pgtype.Text{String: input.Slug, Valid: true},
		ShortName:             pgtype.Text{String: input.ShortName, Valid: true},
		CategoryID:            pgtype.Int8{Int64: int64(input.CategoryID), Valid: true},
		IsNew:                 input.IsNew,
		Description:           pgtype.Text{String: input.Description, Valid: true},
		Features:              pgtype.Text{String: input.Features, Valid: true},
		BoxIncludes:           input.BoxIncludes,
		Gallery:               input.Gallery,
		CategoryImage:         pgtype.Text{String: input.CategoryImage, Valid: true},
		RecommendedProductIds: input.RecommendedProductIDs,
	}
}
