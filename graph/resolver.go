package graph

import repo "github/M-b-a-s/e-comm/internal/adapters/postgresql/sqlc"

// Resolver holds dependencies shared by GraphQL resolvers.
type Resolver struct {
	Queries *repo.Queries
}
