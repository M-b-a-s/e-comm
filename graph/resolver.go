package graph

import (
	"github/M-b-a-s/e-comm/internal/auth"
	"github/M-b-a-s/e-comm/internal/otp"
	"github/M-b-a-s/e-comm/internal/products"
	"github/M-b-a-s/e-comm/internal/users"
)

// Resolver holds dependencies shared by GraphQL resolvers.
type Resolver struct {
	Accounts       *auth.AccountService
	OTPStore       otp.Store
	EmailSender    otp.Sender
	ProductService *products.Service
	UserService    *users.Service
}
