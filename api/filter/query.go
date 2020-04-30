package filter

import (
	"log"
	"net/http"

	"github.com/pankrator/payment/query"

	"github.com/pankrator/payment/model"
	"github.com/pankrator/payment/storage"
	"github.com/pankrator/payment/web"
)

type Query struct {
	repository storage.Storage
}

func NewQueryFilter(repository storage.Storage) *Query {
	return &Query{
		repository: repository,
	}
}

func (q *Query) Execute(rw http.ResponseWriter, req *http.Request, next http.HandlerFunc) {
	ctx := req.Context()
	user, found := web.UserFromContext(ctx)
	if !found {
		web.WriteError(rw, &web.HTTPError{
			StatusCode:  http.StatusUnauthorized,
			Description: "No user found",
		})
		return
	}
	object, err := q.repository.GetBy(model.MerchantType, "email = ?", user.Email)
	if err != nil {
		if err == storage.ErrNotFound {
			log.Printf("Merchant with email %s not found. Proceed without merchant id filter", user.Email)
			next.ServeHTTP(rw, req)
			return
		} else {
			web.WriteError(rw, err)
			return
		}
	}
	merchant := object.(*model.Merchant)

	log.Printf("Adding query on transaction for merchant %s", user.Email)
	ctx = query.AddQuery(ctx, query.Query{
		Type:      model.TransactionObjectType,
		Key:       "merchant_id",
		Operation: "=",
		Value:     merchant.UUID,
	})
	req = req.WithContext(ctx)

	next.ServeHTTP(rw, req)
}

func (q *Query) Matchers() []web.Endpoint {
	return []web.Endpoint{
		{
			Path:   "/payment",
			Method: http.MethodGet,
		},
	}
}
