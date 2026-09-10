package graph

import (
	"strconv"
	"time"

	"github/M-b-a-s/e-comm/graph/model"
	repo "github/M-b-a-s/e-comm/internal/adapters/postgresql/sqlc"
	"github/M-b-a-s/e-comm/internal/scalar"

	"github.com/jackc/pgx/v5/pgtype"
)

func productToModel(product repo.Product) *model.Product {
	return &model.Product{
		ID:                    strconv.FormatInt(product.ID, 10),
		Name:                  product.Name,
		PriceInCents:          product.PriceInCents,
		CreatedAt:             productCreatedAt(product.CreatedAt),
		Slug:                  nullableString(product.Slug),
		ShortName:             nullableString(product.ShortName),
		CategoryID:            nullableID(product.CategoryID),
		IsNew:                 product.IsNew,
		Description:           nullableString(product.Description),
		Features:              nullableString(product.Features),
		BoxIncludes:           scalar.JSON(product.BoxIncludes),
		Gallery:               scalar.JSON(product.Gallery),
		CategoryImage:         nullableString(product.CategoryImage),
		RecommendedProductIds: idsToStrings(product.RecommendedProductIds),
	}
}

func idsToInt64s(values []int32) []int64 {
	result := make([]int64, len(values))
	for i, value := range values {
		result[i] = int64(value)
	}
	return result
}

func idsToStrings(values []int64) []string {
	result := make([]string, len(values))
	for i, value := range values {
		result[i] = strconv.FormatInt(value, 10)
	}
	return result
}

func nullableString(value pgtype.Text) *string {
	if !value.Valid {
		return nil
	}
	return &value.String
}

func nullableID(value pgtype.Int8) *string {
	if !value.Valid {
		return nil
	}
	id := strconv.FormatInt(value.Int64, 10)
	return &id
}

func productCreatedAt(value pgtype.Timestamp) *time.Time {
	if !value.Valid {
		return nil
	}
	return &value.Time
}
